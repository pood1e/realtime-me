package postgres

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/cloud-drive/api/internal/domain"
)

const itemColumns = `uid, parent_uid, name, kind, size_bytes, content_type, storage_key, create_time, update_time, delete_time`

// Store persists drive metadata in PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

// Open creates a pool and applies the embedded schema migrations.
func Open(ctx context.Context, databaseURL string) (*Store, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 15 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}
	store := &Store{pool: pool}
	if err := store.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if err := Migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

// Close releases database connections.
func (s *Store) Close() {
	s.pool.Close()
}

// Ping verifies database reachability.
func (s *Store) Ping(ctx context.Context) error {
	if err := s.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}
	return nil
}

// GetItem returns an item, optionally including trashed items.
func (s *Store) GetItem(ctx context.Context, uid string, includeTrashed bool) (domain.Item, error) {
	query := "SELECT " + itemColumns + " FROM drive_items WHERE uid = $1"
	if !includeTrashed {
		query += " AND delete_time IS NULL"
	}
	item, err := scanItem(s.pool.QueryRow(ctx, query, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Item{}, fmt.Errorf("%w: drive item", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Item{}, fmt.Errorf("get drive item: %w", err)
	}
	return item, nil
}

// ListItems lists direct children with a stable opaque cursor.
func (s *Store) ListItems(ctx context.Context, parentUID *string, includeTrashed bool, pageSize int, pageToken string) (domain.Page, error) {
	cursor, err := decodeCursor(pageToken)
	if err != nil {
		return domain.Page{}, err
	}
	pageSize = normalizePageSize(pageSize)
	query := "SELECT " + itemColumns + " FROM drive_items WHERE parent_uid IS NOT DISTINCT FROM $1"
	arguments := []any{nullableString(parentUID)}
	if !includeTrashed {
		query += " AND delete_time IS NULL"
	}
	if cursor != nil {
		query += fmt.Sprintf(" AND (name, uid) > ($%d, $%d)", len(arguments)+1, len(arguments)+2)
		arguments = append(arguments, cursor.name, cursor.uid)
	}
	query += fmt.Sprintf(" ORDER BY name ASC, uid ASC LIMIT $%d", len(arguments)+1)
	arguments = append(arguments, pageSize+1)
	return s.queryItemPage(ctx, query, arguments, pageSize)
}

// ListTrashedItems lists only trash roots so nested deleted items are not duplicated.
func (s *Store) ListTrashedItems(ctx context.Context, pageSize int, pageToken string) (domain.Page, error) {
	cursor, err := decodeCursor(pageToken)
	if err != nil {
		return domain.Page{}, err
	}
	pageSize = normalizePageSize(pageSize)
	query := `SELECT ` + itemColumns + ` FROM drive_items item
		WHERE item.delete_time IS NOT NULL
		AND (item.parent_uid IS NULL OR NOT EXISTS (
			SELECT 1 FROM drive_items parent WHERE parent.uid = item.parent_uid AND parent.delete_time IS NOT NULL
		))`
	arguments := []any{}
	if cursor != nil {
		query += fmt.Sprintf(" AND (item.name, item.uid) > ($%d, $%d)", len(arguments)+1, len(arguments)+2)
		arguments = append(arguments, cursor.name, cursor.uid)
	}
	query += fmt.Sprintf(" ORDER BY item.name ASC, item.uid ASC LIMIT $%d", len(arguments)+1)
	arguments = append(arguments, pageSize+1)
	return s.queryItemPage(ctx, query, arguments, pageSize)
}

// SearchItems searches visible item names with a stable opaque cursor.
func (s *Store) SearchItems(ctx context.Context, queryText string, pageSize int, pageToken string) (domain.Page, error) {
	cursor, err := decodeCursor(pageToken)
	if err != nil {
		return domain.Page{}, err
	}
	pageSize = normalizePageSize(pageSize)
	query := "SELECT " + itemColumns + " FROM drive_items WHERE delete_time IS NULL AND name ILIKE '%' || $1 || '%'"
	arguments := []any{queryText}
	if cursor != nil {
		query += fmt.Sprintf(" AND (name, uid) > ($%d, $%d)", len(arguments)+1, len(arguments)+2)
		arguments = append(arguments, cursor.name, cursor.uid)
	}
	query += fmt.Sprintf(" ORDER BY name ASC, uid ASC LIMIT $%d", len(arguments)+1)
	arguments = append(arguments, pageSize+1)
	return s.queryItemPage(ctx, query, arguments, pageSize)
}

// CreateDirectory creates a visible directory under the supplied parent.
func (s *Store) CreateDirectory(ctx context.Context, parentUID *string, name string) (domain.Item, error) {
	if err := s.validateParent(ctx, parentUID); err != nil {
		return domain.Item{}, err
	}
	now := time.Now().UTC()
	item := domain.Item{
		UID:        uuid.NewString(),
		ParentUID:  copyString(parentUID),
		Name:       name,
		Kind:       domain.ItemKindDirectory,
		CreateTime: now,
		UpdateTime: now,
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO drive_items
		(uid, parent_uid, name, kind, size_bytes, content_type, storage_key, create_time, update_time)
		VALUES ($1, $2, $3, $4, 0, '', '', $5, $5)`, item.UID, nullableString(item.ParentUID), item.Name, string(item.Kind), now)
	if err != nil {
		return domain.Item{}, fmt.Errorf("create directory: %w", err)
	}
	return item, nil
}

// RenameItem changes a visible item's basename.
func (s *Store) RenameItem(ctx context.Context, uid, name string) (domain.Item, error) {
	item, err := scanItem(s.pool.QueryRow(ctx, `UPDATE drive_items SET name = $2, update_time = now()
		WHERE uid = $1 AND delete_time IS NULL RETURNING `+itemColumns, uid, name))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Item{}, fmt.Errorf("%w: drive item", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Item{}, fmt.Errorf("rename drive item: %w", err)
	}
	return item, nil
}

// MoveItem moves a visible item, rejecting directory cycles.
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
	updated, err := scanItem(s.pool.QueryRow(ctx, `UPDATE drive_items SET parent_uid = $2, update_time = now()
		WHERE uid = $1 AND delete_time IS NULL RETURNING `+itemColumns, uid, nullableString(parentUID)))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Item{}, fmt.Errorf("%w: drive item", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Item{}, fmt.Errorf("move drive item: %w", err)
	}
	return updated, nil
}

// TrashItem places an item and descendants in the trash.
func (s *Store) TrashItem(ctx context.Context, uid string) (domain.Item, error) {
	if _, err := s.GetItem(ctx, uid, false); err != nil {
		return domain.Item{}, err
	}
	_, err := s.pool.Exec(ctx, `WITH RECURSIVE subtree AS (
		SELECT uid FROM drive_items WHERE uid = $1
		UNION ALL
		SELECT child.uid FROM drive_items child JOIN subtree ON child.parent_uid = subtree.uid
	)
	UPDATE drive_items SET delete_time = now(), update_time = now()
	WHERE uid IN (SELECT uid FROM subtree)`, uid)
	if err != nil {
		return domain.Item{}, fmt.Errorf("trash drive item: %w", err)
	}
	return s.GetItem(ctx, uid, true)
}

// RestoreItem restores an item and its descendants. A deleted parent restores to root.
func (s *Store) RestoreItem(ctx context.Context, uid string) (domain.Item, error) {
	item, err := s.GetItem(ctx, uid, true)
	if err != nil {
		return domain.Item{}, err
	}
	if item.DeleteTime == nil {
		return domain.Item{}, fmt.Errorf("%w: drive item is not trashed", domain.ErrConflict)
	}
	if item.ParentUID != nil {
		parent, err := s.GetItem(ctx, *item.ParentUID, true)
		if err == nil && parent.DeleteTime != nil {
			if _, err := s.pool.Exec(ctx, "UPDATE drive_items SET parent_uid = NULL, update_time = now() WHERE uid = $1", uid); err != nil {
				return domain.Item{}, fmt.Errorf("restore item parent: %w", err)
			}
		}
	}
	_, err = s.pool.Exec(ctx, `WITH RECURSIVE subtree AS (
		SELECT uid FROM drive_items WHERE uid = $1
		UNION ALL
		SELECT child.uid FROM drive_items child JOIN subtree ON child.parent_uid = subtree.uid
	)
	UPDATE drive_items SET delete_time = NULL, update_time = now()
	WHERE uid IN (SELECT uid FROM subtree)`, uid)
	if err != nil {
		return domain.Item{}, fmt.Errorf("restore drive item: %w", err)
	}
	return s.GetItem(ctx, uid, false)
}

// CreateUpload persists an active upload session after its temporary file exists.
func (s *Store) CreateUpload(ctx context.Context, upload domain.Upload) (domain.Upload, error) {
	if err := s.validateParent(ctx, upload.ParentUID); err != nil {
		return domain.Upload{}, err
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO uploads
		(uid, item_uid, parent_uid, file_name, content_type, total_size_bytes, received_bytes, chunk_size_bytes, status, create_time, expire_time)
		VALUES ($1, $2, $3, $4, $5, $6, 0, $7, $8, $9, $10)`,
		upload.UID, upload.ItemUID, nullableString(upload.ParentUID), upload.FileName, upload.ContentType,
		upload.TotalSizeBytes, upload.ChunkSizeBytes, string(upload.Status), upload.CreateTime, upload.ExpireTime)
	if err != nil {
		return domain.Upload{}, fmt.Errorf("create upload: %w", err)
	}
	return s.GetUpload(ctx, upload.UID)
}

// ReservedUploadBytes returns the remaining byte capacity held by active sessions.
func (s *Store) ReservedUploadBytes(ctx context.Context) (int64, error) {
	var reserved int64
	if err := s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(total_size_bytes - received_bytes), 0)
		FROM uploads WHERE status = 'active' AND expire_time > now()`).Scan(&reserved); err != nil {
		return 0, fmt.Errorf("sum reserved upload bytes: %w", err)
	}
	return reserved, nil
}

// GetUpload returns session metadata and acknowledged ranges.
func (s *Store) GetUpload(ctx context.Context, uid string) (domain.Upload, error) {
	upload, err := scanUpload(s.pool.QueryRow(ctx, `SELECT uid, item_uid, parent_uid, file_name, content_type,
		total_size_bytes, received_bytes, chunk_size_bytes, status, create_time, expire_time
		FROM uploads WHERE uid = $1`, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Upload{}, fmt.Errorf("%w: upload", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Upload{}, fmt.Errorf("get upload: %w", err)
	}
	if upload.Status == domain.UploadStatusActive && !upload.ExpireTime.After(time.Now().UTC()) {
		if _, err := s.pool.Exec(ctx, "UPDATE uploads SET status = 'expired' WHERE uid = $1 AND status = 'active'", uid); err != nil {
			return domain.Upload{}, fmt.Errorf("expire upload: %w", err)
		}
		upload.Status = domain.UploadStatusExpired
	}
	rows, err := s.pool.Query(ctx, `SELECT start_offset, end_offset, checksum FROM upload_chunks WHERE upload_uid = $1 ORDER BY start_offset`, uid)
	if err != nil {
		return domain.Upload{}, fmt.Errorf("get upload chunks: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var chunk domain.UploadChunk
		if err := rows.Scan(&chunk.StartOffset, &chunk.EndOffset, &chunk.Checksum); err != nil {
			return domain.Upload{}, fmt.Errorf("scan upload chunk: %w", err)
		}
		upload.Chunks = append(upload.Chunks, chunk)
	}
	if err := rows.Err(); err != nil {
		return domain.Upload{}, fmt.Errorf("iterate upload chunks: %w", err)
	}
	return upload, nil
}

// RecordUploadChunk atomically acknowledges one non-overlapping chunk.
func (s *Store) RecordUploadChunk(ctx context.Context, uploadUID string, chunk domain.UploadChunk) (domain.Upload, error) {
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Upload{}, fmt.Errorf("begin chunk transaction: %w", err)
	}
	defer transaction.Rollback(ctx)

	var status string
	var expireTime time.Time
	if err := transaction.QueryRow(ctx, "SELECT status, expire_time FROM uploads WHERE uid = $1 FOR UPDATE", uploadUID).Scan(&status, &expireTime); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Upload{}, fmt.Errorf("%w: upload", domain.ErrNotFound)
		}
		return domain.Upload{}, fmt.Errorf("lock upload: %w", err)
	}
	if domain.UploadStatus(status) != domain.UploadStatusActive || !expireTime.After(time.Now().UTC()) {
		if domain.UploadStatus(status) == domain.UploadStatusActive {
			_, _ = transaction.Exec(ctx, "UPDATE uploads SET status = 'expired' WHERE uid = $1", uploadUID)
		}
		return domain.Upload{}, fmt.Errorf("%w: upload is not active", domain.ErrConflict)
	}

	var existingEnd int64
	var existingChecksum []byte
	err = transaction.QueryRow(ctx, `SELECT end_offset, checksum FROM upload_chunks
		WHERE upload_uid = $1 AND start_offset = $2`, uploadUID, chunk.StartOffset).Scan(&existingEnd, &existingChecksum)
	if err == nil {
		if existingEnd != chunk.EndOffset || !bytes.Equal(existingChecksum, chunk.Checksum) {
			return domain.Upload{}, fmt.Errorf("%w: chunk offset already has different content", domain.ErrConflict)
		}
		if err := transaction.Commit(ctx); err != nil {
			return domain.Upload{}, fmt.Errorf("commit existing chunk: %w", err)
		}
		return s.GetUpload(ctx, uploadUID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Upload{}, fmt.Errorf("get existing chunk: %w", err)
	}
	var overlaps bool
	if err := transaction.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM upload_chunks WHERE upload_uid = $1 AND start_offset < $3 AND end_offset > $2
	)`, uploadUID, chunk.StartOffset, chunk.EndOffset).Scan(&overlaps); err != nil {
		return domain.Upload{}, fmt.Errorf("check chunk overlap: %w", err)
	}
	if overlaps {
		return domain.Upload{}, fmt.Errorf("%w: chunk overlaps an acknowledged range", domain.ErrConflict)
	}
	if _, err := transaction.Exec(ctx, `INSERT INTO upload_chunks (upload_uid, start_offset, end_offset, checksum)
		VALUES ($1, $2, $3, $4)`, uploadUID, chunk.StartOffset, chunk.EndOffset, chunk.Checksum); err != nil {
		return domain.Upload{}, fmt.Errorf("insert upload chunk: %w", err)
	}
	if _, err := transaction.Exec(ctx, `UPDATE uploads SET received_bytes = (
		SELECT COALESCE(SUM(end_offset - start_offset), 0) FROM upload_chunks WHERE upload_uid = $1
	) WHERE uid = $1`, uploadUID); err != nil {
		return domain.Upload{}, fmt.Errorf("update upload received bytes: %w", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return domain.Upload{}, fmt.Errorf("commit upload chunk: %w", err)
	}
	return s.GetUpload(ctx, uploadUID)
}

// CompleteUpload publishes the metadata for a contiguous, fully written file.
func (s *Store) CompleteUpload(ctx context.Context, uploadUID string) (domain.Item, error) {
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Item{}, fmt.Errorf("begin complete upload transaction: %w", err)
	}
	defer transaction.Rollback(ctx)

	upload, err := scanUpload(transaction.QueryRow(ctx, `SELECT uid, item_uid, parent_uid, file_name, content_type,
		total_size_bytes, received_bytes, chunk_size_bytes, status, create_time, expire_time
		FROM uploads WHERE uid = $1 FOR UPDATE`, uploadUID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Item{}, fmt.Errorf("%w: upload", domain.ErrNotFound)
	}
	if err != nil {
		return domain.Item{}, fmt.Errorf("lock upload for completion: %w", err)
	}
	if upload.Status == domain.UploadStatusCompleted {
		if err := transaction.Commit(ctx); err != nil {
			return domain.Item{}, fmt.Errorf("commit completed upload read: %w", err)
		}
		return s.GetItem(ctx, upload.ItemUID, false)
	}
	if upload.Status != domain.UploadStatusActive || !upload.ExpireTime.After(time.Now().UTC()) {
		return domain.Item{}, fmt.Errorf("%w: upload is not active", domain.ErrConflict)
	}
	rows, err := transaction.Query(ctx, `SELECT start_offset, end_offset FROM upload_chunks WHERE upload_uid = $1 ORDER BY start_offset`, uploadUID)
	if err != nil {
		return domain.Item{}, fmt.Errorf("read upload ranges: %w", err)
	}
	var expectedOffset int64
	for rows.Next() {
		var startOffset, endOffset int64
		if err := rows.Scan(&startOffset, &endOffset); err != nil {
			rows.Close()
			return domain.Item{}, fmt.Errorf("scan upload range: %w", err)
		}
		if startOffset != expectedOffset {
			rows.Close()
			return domain.Item{}, fmt.Errorf("%w: upload has missing byte ranges", domain.ErrConflict)
		}
		expectedOffset = endOffset
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return domain.Item{}, fmt.Errorf("iterate upload ranges: %w", err)
	}
	rows.Close()
	if expectedOffset != upload.TotalSizeBytes {
		return domain.Item{}, fmt.Errorf("%w: upload size is incomplete", domain.ErrConflict)
	}
	if err := s.validateParentTx(ctx, transaction, upload.ParentUID); err != nil {
		return domain.Item{}, err
	}
	now := time.Now().UTC()
	if _, err := transaction.Exec(ctx, `INSERT INTO drive_items
		(uid, parent_uid, name, kind, size_bytes, content_type, storage_key, create_time, update_time)
		VALUES ($1, $2, $3, 'file', $4, $5, $1, $6, $6)`, upload.ItemUID, nullableString(upload.ParentUID), upload.FileName, upload.TotalSizeBytes, upload.ContentType, now); err != nil {
		return domain.Item{}, fmt.Errorf("create completed drive item: %w", err)
	}
	if _, err := transaction.Exec(ctx, "UPDATE uploads SET status = 'completed', complete_time = $2 WHERE uid = $1", uploadUID, now); err != nil {
		return domain.Item{}, fmt.Errorf("mark upload completed: %w", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return domain.Item{}, fmt.Errorf("commit completed upload: %w", err)
	}
	return s.GetItem(ctx, upload.ItemUID, false)
}

// CreateShare persists a hashed-token share link for a visible target item.
func (s *Store) CreateShare(ctx context.Context, share domain.ShareLink, tokenHash []byte) (domain.ShareLink, error) {
	if _, err := s.GetItem(ctx, share.TargetUID, false); err != nil {
		return domain.ShareLink{}, err
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO share_links (uid, target_uid, token_hash, create_time, expire_time)
		VALUES ($1, $2, $3, $4, $5)`, share.UID, share.TargetUID, tokenHash, share.CreateTime, share.ExpireTime)
	if err != nil {
		return domain.ShareLink{}, fmt.Errorf("create share link: %w", err)
	}
	return share, nil
}

// ListShareLinks lists owner-visible share metadata without bearer tokens.
func (s *Store) ListShareLinks(ctx context.Context, targetUID string, pageSize int, pageToken string) (domain.SharePage, error) {
	if _, err := s.GetItem(ctx, targetUID, true); err != nil {
		return domain.SharePage{}, err
	}
	cursor, err := decodeShareCursor(pageToken)
	if err != nil {
		return domain.SharePage{}, err
	}
	pageSize = normalizePageSize(pageSize)
	query := `SELECT uid, target_uid, create_time, expire_time, revoke_time FROM share_links WHERE target_uid = $1`
	arguments := []any{targetUID}
	if cursor != "" {
		query += fmt.Sprintf(" AND uid > $%d", len(arguments)+1)
		arguments = append(arguments, cursor)
	}
	query += fmt.Sprintf(" ORDER BY uid ASC LIMIT $%d", len(arguments)+1)
	arguments = append(arguments, pageSize+1)
	rows, err := s.pool.Query(ctx, query, arguments...)
	if err != nil {
		return domain.SharePage{}, fmt.Errorf("query share links: %w", err)
	}
	defer rows.Close()
	page := domain.SharePage{ShareLinks: make([]domain.ShareLink, 0, pageSize)}
	for rows.Next() {
		share, err := scanShare(rows)
		if err != nil {
			return domain.SharePage{}, fmt.Errorf("scan share link: %w", err)
		}
		page.ShareLinks = append(page.ShareLinks, share)
	}
	if err := rows.Err(); err != nil {
		return domain.SharePage{}, fmt.Errorf("iterate share links: %w", err)
	}
	if len(page.ShareLinks) > pageSize {
		last := page.ShareLinks[pageSize-1]
		page.ShareLinks = page.ShareLinks[:pageSize]
		page.NextPageToken = encodeShareCursor(last.UID)
	}
	return page, nil
}

// GetShareByTokenHash resolves an active, non-expired share by its token hash.
func (s *Store) GetShareByTokenHash(ctx context.Context, tokenHash []byte) (domain.ShareLink, error) {
	share, err := scanShare(s.pool.QueryRow(ctx, `SELECT uid, target_uid, create_time, expire_time, revoke_time
		FROM share_links WHERE token_hash = $1 AND revoke_time IS NULL AND expire_time > now()`, tokenHash))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ShareLink{}, fmt.Errorf("%w: active share link", domain.ErrNotFound)
	}
	if err != nil {
		return domain.ShareLink{}, fmt.Errorf("get share link: %w", err)
	}
	return share, nil
}

// RevokeShare disables a link immediately.
func (s *Store) RevokeShare(ctx context.Context, uid string) (domain.ShareLink, error) {
	share, err := scanShare(s.pool.QueryRow(ctx, `UPDATE share_links SET revoke_time = now()
		WHERE uid = $1 AND revoke_time IS NULL RETURNING uid, target_uid, create_time, expire_time, revoke_time`, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ShareLink{}, fmt.Errorf("%w: active share link", domain.ErrNotFound)
	}
	if err != nil {
		return domain.ShareLink{}, fmt.Errorf("revoke share link: %w", err)
	}
	return share, nil
}

// ListSharedItems lists direct children within an active shared directory.
func (s *Store) ListSharedItems(ctx context.Context, shareUID string, parentUID *string, pageSize int, pageToken string) (domain.Page, error) {
	share, err := s.getActiveShareByUID(ctx, shareUID)
	if err != nil {
		return domain.Page{}, err
	}
	target, err := s.GetItem(ctx, share.TargetUID, false)
	if err != nil {
		return domain.Page{}, err
	}
	if target.Kind != domain.ItemKindDirectory {
		return domain.Page{}, fmt.Errorf("%w: a shared file has no children", domain.ErrConflict)
	}
	if parentUID == nil {
		parentUID = &share.TargetUID
	}
	inside, err := s.isWithin(ctx, share.TargetUID, *parentUID)
	if err != nil {
		return domain.Page{}, err
	}
	if !inside {
		return domain.Page{}, fmt.Errorf("%w: item is outside the share", domain.ErrForbidden)
	}
	parent, err := s.GetItem(ctx, *parentUID, false)
	if err != nil {
		return domain.Page{}, err
	}
	if parent.Kind != domain.ItemKindDirectory {
		return domain.Page{}, fmt.Errorf("%w: parent is not a directory", domain.ErrConflict)
	}
	return s.ListItems(ctx, parentUID, false, pageSize, pageToken)
}

// CanReadSharedItem verifies that an active share covers a visible item.
func (s *Store) CanReadSharedItem(ctx context.Context, shareUID, itemUID string) (domain.Item, error) {
	share, err := s.getActiveShareByUID(ctx, shareUID)
	if err != nil {
		return domain.Item{}, err
	}
	inside, err := s.isWithin(ctx, share.TargetUID, itemUID)
	if err != nil {
		return domain.Item{}, err
	}
	if !inside {
		return domain.Item{}, fmt.Errorf("%w: item is outside the share", domain.ErrForbidden)
	}
	return s.GetItem(ctx, itemUID, false)
}

// ListExpiredUploads returns sessions whose temporary files can be removed.
func (s *Store) ListExpiredUploads(ctx context.Context, cutoff time.Time) ([]domain.Upload, error) {
	rows, err := s.pool.Query(ctx, `SELECT uid, item_uid, parent_uid, file_name, content_type,
		total_size_bytes, received_bytes, chunk_size_bytes, status, create_time, expire_time
		FROM uploads WHERE status <> 'completed' AND expire_time <= $1 ORDER BY expire_time ASC`, cutoff)
	if err != nil {
		return nil, fmt.Errorf("query expired uploads: %w", err)
	}
	defer rows.Close()
	uploads := make([]domain.Upload, 0)
	for rows.Next() {
		upload, err := scanUpload(rows)
		if err != nil {
			return nil, fmt.Errorf("scan expired upload: %w", err)
		}
		uploads = append(uploads, upload)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expired uploads: %w", err)
	}
	return uploads, nil
}

// DeleteUpload deletes metadata after its temporary file has been removed.
func (s *Store) DeleteUpload(ctx context.Context, uid string) error {
	command, err := s.pool.Exec(ctx, "DELETE FROM uploads WHERE uid = $1 AND status <> 'completed'", uid)
	if err != nil {
		return fmt.Errorf("delete expired upload: %w", err)
	}
	if command.RowsAffected() == 0 {
		return fmt.Errorf("%w: expired upload", domain.ErrNotFound)
	}
	return nil
}

// PurgeTrashedItems atomically removes eligible metadata, returning blobs for later orphan cleanup.
func (s *Store) PurgeTrashedItems(ctx context.Context, cutoff time.Time) ([]domain.Item, error) {
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin trash purge: %w", err)
	}
	defer transaction.Rollback(ctx)
	if _, err := transaction.Exec(ctx, `DELETE FROM share_links
		WHERE target_uid IN (SELECT uid FROM drive_items WHERE delete_time IS NOT NULL AND delete_time <= $1)`, cutoff); err != nil {
		return nil, fmt.Errorf("purge expired share links: %w", err)
	}
	rows, err := transaction.Query(ctx, "DELETE FROM drive_items WHERE delete_time IS NOT NULL AND delete_time <= $1 RETURNING "+itemColumns, cutoff)
	if err != nil {
		return nil, fmt.Errorf("purge expired drive items: %w", err)
	}
	items := make([]domain.Item, 0)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan purged drive item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate purged drive items: %w", err)
	}
	rows.Close()
	if err := transaction.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit trash purge: %w", err)
	}
	return items, nil
}

func (s *Store) queryItemPage(ctx context.Context, query string, arguments []any, pageSize int) (domain.Page, error) {
	rows, err := s.pool.Query(ctx, query, arguments...)
	if err != nil {
		return domain.Page{}, fmt.Errorf("query drive item page: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Item, 0, pageSize)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return domain.Page{}, fmt.Errorf("scan drive item page: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return domain.Page{}, fmt.Errorf("iterate drive item page: %w", err)
	}
	page := domain.Page{Items: items}
	if len(items) > pageSize {
		last := items[pageSize-1]
		page.Items = items[:pageSize]
		page.NextPageToken = encodeCursor(last.Name, last.UID)
	}
	return page, nil
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

func (s *Store) validateParentTx(ctx context.Context, transaction pgx.Tx, parentUID *string) error {
	if parentUID == nil {
		return nil
	}
	item, err := scanItem(transaction.QueryRow(ctx, "SELECT "+itemColumns+" FROM drive_items WHERE uid = $1 AND delete_time IS NULL", *parentUID))
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: parent drive item", domain.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("get upload parent: %w", err)
	}
	if item.Kind != domain.ItemKindDirectory {
		return fmt.Errorf("%w: parent is not a directory", domain.ErrConflict)
	}
	return nil
}

func (s *Store) isWithin(ctx context.Context, rootUID, itemUID string) (bool, error) {
	var inside bool
	err := s.pool.QueryRow(ctx, `WITH RECURSIVE ancestry AS (
		SELECT uid, parent_uid FROM drive_items WHERE uid = $1
		UNION ALL
		SELECT parent.uid, parent.parent_uid FROM drive_items parent JOIN ancestry ON ancestry.parent_uid = parent.uid
	)
	SELECT EXISTS (SELECT 1 FROM ancestry WHERE uid = $2)`, itemUID, rootUID).Scan(&inside)
	if err != nil {
		return false, fmt.Errorf("check item ancestry: %w", err)
	}
	return inside, nil
}

func (s *Store) getActiveShareByUID(ctx context.Context, uid string) (domain.ShareLink, error) {
	share, err := scanShare(s.pool.QueryRow(ctx, `SELECT uid, target_uid, create_time, expire_time, revoke_time
		FROM share_links WHERE uid = $1 AND revoke_time IS NULL AND expire_time > now()`, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ShareLink{}, fmt.Errorf("%w: active share link", domain.ErrNotFound)
	}
	if err != nil {
		return domain.ShareLink{}, fmt.Errorf("get active share link: %w", err)
	}
	return share, nil
}

type rowScanner interface {
	Scan(...any) error
}

func scanItem(row rowScanner) (domain.Item, error) {
	var item domain.Item
	var parentUID pgtype.Text
	var deleteTime pgtype.Timestamptz
	var kind string
	if err := row.Scan(&item.UID, &parentUID, &item.Name, &kind, &item.SizeBytes, &item.ContentType, &item.StorageKey,
		&item.CreateTime, &item.UpdateTime, &deleteTime); err != nil {
		return domain.Item{}, err
	}
	item.Kind = domain.ItemKind(kind)
	if parentUID.Valid {
		item.ParentUID = copyString(&parentUID.String)
	}
	if deleteTime.Valid {
		deleteTimeUTC := deleteTime.Time.UTC()
		item.DeleteTime = &deleteTimeUTC
	}
	return item, nil
}

func scanUpload(row rowScanner) (domain.Upload, error) {
	var upload domain.Upload
	var parentUID pgtype.Text
	var status string
	if err := row.Scan(&upload.UID, &upload.ItemUID, &parentUID, &upload.FileName, &upload.ContentType,
		&upload.TotalSizeBytes, &upload.ReceivedBytes, &upload.ChunkSizeBytes, &status, &upload.CreateTime, &upload.ExpireTime); err != nil {
		return domain.Upload{}, err
	}
	upload.Status = domain.UploadStatus(status)
	if parentUID.Valid {
		upload.ParentUID = copyString(&parentUID.String)
	}
	return upload, nil
}

func scanShare(row rowScanner) (domain.ShareLink, error) {
	var share domain.ShareLink
	var revokeTime pgtype.Timestamptz
	if err := row.Scan(&share.UID, &share.TargetUID, &share.CreateTime, &share.ExpireTime, &revokeTime); err != nil {
		return domain.ShareLink{}, err
	}
	if revokeTime.Valid {
		revokeTimeUTC := revokeTime.Time.UTC()
		share.RevokeTime = &revokeTimeUTC
	}
	return share, nil
}

type cursor struct {
	name string
	uid  string
}

func encodeCursor(name, uid string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(name + "\x00" + uid))
}

func decodeCursor(token string) (*cursor, error) {
	if token == "" {
		return nil, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("%w: malformed page token", domain.ErrInvalidArgument)
	}
	parts := strings.Split(string(decoded), "\x00")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("%w: malformed page token", domain.ErrInvalidArgument)
	}
	return &cursor{name: parts[0], uid: parts[1]}, nil
}

func encodeShareCursor(uid string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(uid))
}

func decodeShareCursor(token string) (string, error) {
	if token == "" {
		return "", nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(decoded) == 0 {
		return "", fmt.Errorf("%w: malformed page token", domain.ErrInvalidArgument)
	}
	return string(decoded), nil
}

func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return 100
	}
	if pageSize > 200 {
		return 200
	}
	return pageSize
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func copyString(value *string) *string {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
