package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

func (s *Store) CreateDirectory(ctx context.Context, parentUID *string, name string) (domain.Item, error) {
	if err := s.validateParent(ctx, parentUID); err != nil {
		return domain.Item{}, err
	}
	now := time.Now().UTC()
	uid := uuid.NewString()
	if _, err := s.pool.Exec(ctx, `INSERT INTO drive_items
		(uid, parent_uid, name, kind, create_time, update_time) VALUES ($1, $2, $3, 'directory', $4, $4)`,
		uid, nullableString(parentUID), name, now); err != nil {
		return domain.Item{}, fmt.Errorf("create directory: %w", err)
	}
	return s.GetItem(ctx, uid, false)
}

// RenameItem changes a visible basename.
func (s *Store) RenameItem(ctx context.Context, uid, name string) (domain.Item, error) {
	command, err := s.pool.Exec(ctx, `UPDATE drive_items SET name = $2, update_time = now()
		WHERE uid = $1 AND delete_time IS NULL`, uid, name)
	if err != nil {
		return domain.Item{}, fmt.Errorf("rename drive item: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.Item{}, fmt.Errorf("%w: drive item", domain.ErrNotFound)
	}
	return s.GetItem(ctx, uid, false)
}

// MoveItem moves an item while rejecting directory cycles.
func (s *Store) MoveItem(ctx context.Context, uid string, parentUID *string) (domain.Item, error) {
	item, err := s.GetItem(ctx, uid, false)
	if err != nil {
		return domain.Item{}, err
	}
	if err := s.validateParent(ctx, parentUID); err != nil {
		return domain.Item{}, err
	}
	if item.Kind == domain.ItemKindDirectory && parentUID != nil {
		inside, err := s.isWithin(ctx, item.UID, *parentUID)
		if err != nil {
			return domain.Item{}, err
		}
		if inside {
			return domain.Item{}, fmt.Errorf("%w: directory cannot be moved into itself", domain.ErrConflict)
		}
	}
	if _, err := s.pool.Exec(ctx, `UPDATE drive_items SET parent_uid = $2, update_time = now()
		WHERE uid = $1 AND delete_time IS NULL`, uid, nullableString(parentUID)); err != nil {
		return domain.Item{}, fmt.Errorf("move drive item: %w", err)
	}
	return s.GetItem(ctx, uid, false)
}

// TrashItem places an item hierarchy in the trash.
func (s *Store) TrashItem(ctx context.Context, uid string) (domain.Item, error) {
	if _, err := s.GetItem(ctx, uid, false); err != nil {
		return domain.Item{}, err
	}
	if _, err := s.pool.Exec(ctx, `WITH RECURSIVE subtree AS (
		SELECT uid FROM drive_items WHERE uid = $1
		UNION ALL SELECT child.uid FROM drive_items child JOIN subtree ON child.parent_uid = subtree.uid)
		UPDATE drive_items SET delete_time = now(), update_time = now() WHERE uid IN (SELECT uid FROM subtree)`, uid); err != nil {
		return domain.Item{}, fmt.Errorf("trash drive item: %w", err)
	}
	return s.GetItem(ctx, uid, true)
}

// RestoreItem restores an item hierarchy, moving it to root when its parent remains trashed.
func (s *Store) RestoreItem(ctx context.Context, uid string) (domain.Item, error) {
	item, err := s.GetItem(ctx, uid, true)
	if err != nil {
		return domain.Item{}, err
	}
	if item.DeleteTime == nil {
		return domain.Item{}, fmt.Errorf("%w: drive item is not trashed", domain.ErrConflict)
	}
	if item.ParentUID != nil {
		parent, parentErr := s.GetItem(ctx, *item.ParentUID, true)
		if parentErr == nil && parent.DeleteTime != nil {
			if _, err := s.pool.Exec(ctx, "UPDATE drive_items SET parent_uid = NULL WHERE uid = $1", uid); err != nil {
				return domain.Item{}, fmt.Errorf("restore item parent: %w", err)
			}
		}
	}
	if _, err := s.pool.Exec(ctx, `WITH RECURSIVE subtree AS (
		SELECT uid FROM drive_items WHERE uid = $1
		UNION ALL SELECT child.uid FROM drive_items child JOIN subtree ON child.parent_uid = subtree.uid)
		UPDATE drive_items SET delete_time = NULL, update_time = now() WHERE uid IN (SELECT uid FROM subtree)`, uid); err != nil {
		return domain.Item{}, fmt.Errorf("restore drive item: %w", err)
	}
	return s.GetItem(ctx, uid, false)
}

// PurgeTrashedItem permanently removes one trashed hierarchy.
func (s *Store) PurgeTrashedItem(ctx context.Context, uid string) error {
	item, err := s.GetItem(ctx, uid, true)
	if err != nil {
		return err
	}
	if item.DeleteTime == nil {
		return fmt.Errorf("%w: drive item is not trashed", domain.ErrConflict)
	}
	return s.purgeDriveItems(ctx, `WITH RECURSIVE subtree AS (
		SELECT uid FROM drive_items WHERE uid = $1 AND delete_time IS NOT NULL
		UNION ALL SELECT child.uid FROM drive_items child JOIN subtree ON child.parent_uid = subtree.uid
		WHERE child.delete_time IS NOT NULL) SELECT uid FROM subtree`, uid)
}

// EmptyTrash permanently removes all trashed drive items.
func (s *Store) EmptyTrash(ctx context.Context) error {
	return s.purgeDriveItems(ctx, "SELECT uid FROM drive_items WHERE delete_time IS NOT NULL")
}

// PurgeTrashedItems removes drive items older than the retention cutoff.
func (s *Store) PurgeTrashedItems(ctx context.Context, cutoff time.Time) error {
	return s.purgeDriveItems(ctx, "SELECT uid FROM drive_items WHERE delete_time IS NOT NULL AND delete_time <= $1", cutoff)
}

// ImportDriveFile claims one complete upload as a drive file.
func (s *Store) ImportDriveFile(ctx context.Context, uploadUID string, parentUID *string, name string, sealed domain.SealedContent) (domain.Item, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Item{}, fmt.Errorf("begin drive import: %w", err)
	}
	defer tx.Rollback(ctx)
	upload, err := lockCompleteUpload(ctx, tx, uploadUID)
	if err != nil {
		return domain.Item{}, err
	}
	if upload.Status == domain.UploadStatusClaimed {
		if err := tx.Commit(ctx); err != nil {
			return domain.Item{}, fmt.Errorf("commit repeated drive import: %w", err)
		}
		return s.GetItem(ctx, upload.ClaimedUID, false)
	}
	if err := s.validateParentTx(ctx, tx, parentUID); err != nil {
		return domain.Item{}, err
	}
	content, err := contentForUpload(ctx, tx, upload, sealed)
	if err != nil {
		return domain.Item{}, err
	}
	now := time.Now().UTC()
	itemUID := uuid.NewString()
	if _, err := tx.Exec(ctx, `INSERT INTO drive_items
		(uid, parent_uid, content_uid, name, kind, create_time, update_time)
		VALUES ($1, $2, $3, $4, 'file', $5, $5)`, itemUID, nullableString(parentUID), content.UID, name, now); err != nil {
		return domain.Item{}, fmt.Errorf("create drive file: %w", err)
	}
	if err := markUploadClaimed(ctx, tx, uploadUID, itemUID); err != nil {
		return domain.Item{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Item{}, fmt.Errorf("commit drive import: %w", err)
	}
	return s.GetItem(ctx, itemUID, false)
}

// CreateUpload persists an application-neutral active upload.
