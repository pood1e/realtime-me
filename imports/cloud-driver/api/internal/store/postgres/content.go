package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"example.com/cloud-drive/api/internal/domain"
)

func (s *Store) ListUnhashedContent(ctx context.Context, limit int) ([]domain.ContentObject, error) {
	rows, err := s.pool.Query(ctx, `SELECT uid, sha256, size_bytes, content_type, storage_key, create_time
		FROM content_objects WHERE sha256 IS NULL ORDER BY uid LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list legacy content: %w", err)
	}
	defer rows.Close()
	var objects []domain.ContentObject
	for rows.Next() {
		object, err := scanContent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan legacy content: %w", err)
		}
		objects = append(objects, object)
	}
	return objects, rows.Err()
}

// CommitContentMigration records a content address and collapses duplicates.
func (s *Store) CommitContentMigration(ctx context.Context, legacyUID string, sealed domain.SealedContent) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin content migration: %w", err)
	}
	defer tx.Rollback(ctx)
	var currentUID string
	if err := tx.QueryRow(ctx, "SELECT uid FROM content_objects WHERE uid = $1 AND sha256 IS NULL FOR UPDATE", legacyUID).Scan(&currentUID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return legacyUID, nil
		}
		return "", fmt.Errorf("lock legacy content: %w", err)
	}
	var canonicalUID string
	err = tx.QueryRow(ctx, "SELECT uid FROM content_objects WHERE sha256 = $1", sealed.SHA256).Scan(&canonicalUID)
	if errors.Is(err, pgx.ErrNoRows) {
		canonicalUID = legacyUID
		if _, err := tx.Exec(ctx, `UPDATE content_objects SET sha256 = $2, size_bytes = $3,
			content_type = $4, storage_key = $5 WHERE uid = $1`, legacyUID, sealed.SHA256,
			sealed.SizeBytes, sealed.ContentType, sealed.StorageKey); err != nil {
			return "", fmt.Errorf("update migrated content: %w", err)
		}
	} else if err != nil {
		return "", fmt.Errorf("find duplicate content: %w", err)
	} else {
		for _, table := range []string{"drive_items", "books", "tracks", "images"} {
			if _, err := tx.Exec(ctx, "UPDATE "+table+" SET content_uid = $2 WHERE content_uid = $1", legacyUID, canonicalUID); err != nil {
				return "", fmt.Errorf("rewire duplicate content in %s: %w", table, err)
			}
		}
		if _, err := tx.Exec(ctx, "DELETE FROM content_objects WHERE uid = $1", legacyUID); err != nil {
			return "", fmt.Errorf("delete duplicate legacy content: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit content migration: %w", err)
	}
	return canonicalUID, nil
}

// FinalizeContentMigration enforces hashes after all legacy objects are converted.
func (s *Store) FinalizeContentMigration(ctx context.Context) error {
	var remaining bool
	if err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM content_objects WHERE sha256 IS NULL)").Scan(&remaining); err != nil {
		return fmt.Errorf("check content migration: %w", err)
	}
	if remaining {
		return fmt.Errorf("%w: content migration is incomplete", domain.ErrConflict)
	}
	if _, err := s.pool.Exec(ctx, "ALTER TABLE content_objects ALTER COLUMN sha256 SET NOT NULL"); err != nil {
		return fmt.Errorf("finalize content migration: %w", err)
	}
	return nil
}

// GetContent returns immutable content metadata.
func (s *Store) GetContent(ctx context.Context, uid string) (domain.ContentObject, error) {
	content, err := scanContent(s.pool.QueryRow(ctx, `SELECT uid, sha256, size_bytes, content_type, storage_key, create_time
		FROM content_objects WHERE uid = $1`, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ContentObject{}, fmt.Errorf("%w: content", domain.ErrNotFound)
	}
	if err != nil {
		return domain.ContentObject{}, fmt.Errorf("get content: %w", err)
	}
	return content, nil
}

// ListUnreferencedContent returns content with no owning application resource.
func (s *Store) ListUnreferencedContent(ctx context.Context, limit int) ([]domain.ContentObject, error) {
	rows, err := s.pool.Query(ctx, `SELECT content.uid, content.sha256, content.size_bytes, content.content_type,
		content.storage_key, content.create_time FROM content_objects content
		WHERE NOT EXISTS (SELECT 1 FROM drive_items WHERE content_uid = content.uid)
		AND NOT EXISTS (SELECT 1 FROM books WHERE content_uid = content.uid)
		AND NOT EXISTS (SELECT 1 FROM tracks WHERE content_uid = content.uid)
		AND NOT EXISTS (SELECT 1 FROM images WHERE content_uid = content.uid)
		AND NOT EXISTS (SELECT 1 FROM uploads WHERE sealed_content_uid = content.uid)
		ORDER BY content.create_time LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list unreferenced content: %w", err)
	}
	defer rows.Close()
	var objects []domain.ContentObject
	for rows.Next() {
		object, err := scanContent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan unreferenced content: %w", err)
		}
		objects = append(objects, object)
	}
	return objects, rows.Err()
}

// TombstoneContent removes unreferenced metadata while retaining every physical path for delayed deletion.
func (s *Store) TombstoneContent(ctx context.Context, uid string, deleteAfter time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin content tombstone: %w", err)
	}
	defer tx.Rollback(ctx)
	var storageKeys []string
	err = tx.QueryRow(ctx, `SELECT ARRAY(
		SELECT storage_key FROM content_objects WHERE uid = $1
		UNION
		SELECT storage_key FROM content_artifacts WHERE content_uid = $1
		ORDER BY storage_key
	) FROM content_objects WHERE uid = $1 FOR UPDATE`, uid).Scan(&storageKeys)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: content", domain.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("collect content storage keys: %w", err)
	}
	command, err := tx.Exec(ctx, `DELETE FROM content_objects content WHERE uid = $1
		AND NOT EXISTS (SELECT 1 FROM drive_items WHERE content_uid = content.uid)
		AND NOT EXISTS (SELECT 1 FROM books WHERE content_uid = content.uid)
		AND NOT EXISTS (SELECT 1 FROM tracks WHERE content_uid = content.uid)
		AND NOT EXISTS (SELECT 1 FROM images WHERE content_uid = content.uid)
		AND NOT EXISTS (SELECT 1 FROM uploads WHERE sealed_content_uid = content.uid)`, uid)
	if err != nil {
		return fmt.Errorf("delete content metadata: %w", err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("%w: content is referenced", domain.ErrConflict)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO content_tombstones (content_uid, storage_keys, delete_after)
		VALUES ($1, $2, $3)`, uid, storageKeys, deleteAfter); err != nil {
		return fmt.Errorf("create content tombstone: %w", err)
	}
	return tx.Commit(ctx)
}

// ListExpiredContentTombstones returns backup-safe physical deletions.
func (s *Store) ListExpiredContentTombstones(ctx context.Context, cutoff time.Time, limit int) ([]domain.ContentTombstone, error) {
	rows, err := s.pool.Query(ctx, `SELECT content_uid, storage_keys, delete_after
		FROM content_tombstones WHERE delete_after <= $1 ORDER BY delete_after, content_uid LIMIT $2`, cutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("list content tombstones: %w", err)
	}
	defer rows.Close()
	var tombstones []domain.ContentTombstone
	for rows.Next() {
		var tombstone domain.ContentTombstone
		if err := rows.Scan(&tombstone.ContentUID, &tombstone.StorageKeys, &tombstone.DeleteAfter); err != nil {
			return nil, fmt.Errorf("scan content tombstone: %w", err)
		}
		tombstones = append(tombstones, tombstone)
	}
	return tombstones, rows.Err()
}

// DeleteContentTombstone acknowledges successful physical deletion.
func (s *Store) DeleteContentTombstone(ctx context.Context, contentUID string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM content_tombstones WHERE content_uid = $1", contentUID)
	return wrapDatabaseError("delete content tombstone", err)
}

// ReferencedStorageKeys returns the subset still owned by durable metadata.
func (s *Store) ReferencedStorageKeys(ctx context.Context, storageKeys []string) ([]string, error) {
	if len(storageKeys) == 0 {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `WITH candidates(storage_key) AS (
		SELECT unnest($1::text[])
	), referenced(storage_key) AS (
		SELECT storage_key FROM content_objects
		UNION SELECT storage_key FROM content_artifacts
		UNION SELECT sealed_storage_key FROM uploads WHERE sealed_storage_key IS NOT NULL
		UNION SELECT unnest(storage_keys) FROM content_tombstones
	)
	SELECT DISTINCT candidates.storage_key FROM candidates
	JOIN referenced USING (storage_key)`, storageKeys)
	if err != nil {
		return nil, fmt.Errorf("find referenced storage keys: %w", err)
	}
	defer rows.Close()
	referenced := make([]string, 0, len(storageKeys))
	for rows.Next() {
		var storageKey string
		if err := rows.Scan(&storageKey); err != nil {
			return nil, fmt.Errorf("scan referenced storage key: %w", err)
		}
		referenced = append(referenced, storageKey)
	}
	return referenced, rows.Err()
}

// CreateShare persists a hashed-token drive share.
func scanContent(row rowScanner) (domain.ContentObject, error) {
	var content domain.ContentObject
	if err := row.Scan(&content.UID, &content.SHA256, &content.SizeBytes, &content.ContentType,
		&content.StorageKey, &content.CreateTime); err != nil {
		return domain.ContentObject{}, err
	}
	return content, nil
}
