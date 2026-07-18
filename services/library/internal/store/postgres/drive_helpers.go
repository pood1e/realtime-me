package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

func (s *Store) purgeDriveItems(ctx context.Context, candidates string, arguments ...any) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin drive purge: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "DELETE FROM share_links WHERE target_uid IN ("+candidates+")", arguments...); err != nil {
		return fmt.Errorf("delete drive shares: %w", err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM drive_items WHERE uid IN ("+candidates+")", arguments...); err != nil {
		return fmt.Errorf("delete drive items: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit drive purge: %w", err)
	}
	return nil
}

func (s *Store) validateParent(ctx context.Context, parentUID *string) error {
	if parentUID == nil {
		return nil
	}
	parent, err := s.GetItem(ctx, *parentUID, false)
	if err != nil {
		return err
	}
	if parent.Kind != domain.ItemKindDirectory {
		return fmt.Errorf("%w: parent is not a directory", domain.ErrConflict)
	}
	return nil
}

func (s *Store) validateParentTx(ctx context.Context, tx pgx.Tx, parentUID *string) error {
	if parentUID == nil {
		return nil
	}
	var kind string
	if err := tx.QueryRow(ctx, `SELECT kind FROM drive_items WHERE uid = $1 AND delete_time IS NULL`, *parentUID).Scan(&kind); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: parent", domain.ErrNotFound)
		}
		return fmt.Errorf("get parent: %w", err)
	}
	if domain.ItemKind(kind) != domain.ItemKindDirectory {
		return fmt.Errorf("%w: parent is not a directory", domain.ErrConflict)
	}
	return nil
}

func (s *Store) isWithin(ctx context.Context, rootUID, itemUID string) (bool, error) {
	var inside bool
	if err := s.pool.QueryRow(ctx, `WITH RECURSIVE ancestry AS (
		SELECT uid, parent_uid FROM drive_items WHERE uid = $1
		UNION ALL SELECT parent.uid, parent.parent_uid FROM drive_items parent JOIN ancestry ON ancestry.parent_uid = parent.uid)
		SELECT EXISTS(SELECT 1 FROM ancestry WHERE uid = $2)`, itemUID, rootUID).Scan(&inside); err != nil {
		return false, fmt.Errorf("check drive ancestry: %w", err)
	}
	return inside, nil
}

func scanItem(row rowScanner) (domain.Item, error) {
	var item domain.Item
	var parentUID, contentUID pgtype.Text
	var deleteTime pgtype.Timestamptz
	var kind string
	if err := row.Scan(&item.UID, &parentUID, &contentUID, &item.Name, &kind, &item.SizeBytes, &item.ContentType,
		&item.StorageKey, &item.CreateTime, &item.UpdateTime, &deleteTime); err != nil {
		return domain.Item{}, err
	}
	item.Kind = domain.ItemKind(kind)
	if parentUID.Valid {
		item.ParentUID = copyString(&parentUID.String)
	}
	if contentUID.Valid {
		item.ContentUID = copyString(&contentUID.String)
	}
	if deleteTime.Valid {
		value := deleteTime.Time.UTC()
		item.DeleteTime = &value
	}
	return item, nil
}
