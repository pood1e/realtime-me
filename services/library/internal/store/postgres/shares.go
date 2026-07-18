package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pood1e/realtime-me/services/library/internal/domain"
)

func (s *Store) CreateShare(ctx context.Context, share domain.ShareLink, tokenHash []byte) (domain.ShareLink, error) {
	if _, err := s.GetItem(ctx, share.TargetUID, false); err != nil {
		return domain.ShareLink{}, err
	}
	if _, err := s.pool.Exec(ctx, `INSERT INTO share_links (uid, target_uid, token_hash, create_time, expire_time)
		VALUES ($1, $2, $3, $4, $5)`, share.UID, share.TargetUID, tokenHash, share.CreateTime, share.ExpireTime); err != nil {
		return domain.ShareLink{}, fmt.Errorf("create share link: %w", err)
	}
	return share, nil
}

// ListShareLinks lists owner-visible links.
func (s *Store) ListShareLinks(ctx context.Context, targetUID string, pageSize int, pageToken string) (domain.SharePage, error) {
	if _, err := s.GetItem(ctx, targetUID, true); err != nil {
		return domain.SharePage{}, err
	}
	cursor, err := decodeSimpleCursor(pageToken)
	if err != nil {
		return domain.SharePage{}, err
	}
	pageSize = normalizePageSize(pageSize)
	query := `SELECT uid, target_uid, create_time, expire_time, revoke_time FROM share_links WHERE target_uid = $1`
	arguments := []any{targetUID}
	if cursor != "" {
		query += " AND uid > $2"
		arguments = append(arguments, cursor)
	}
	query += fmt.Sprintf(" ORDER BY uid LIMIT $%d", len(arguments)+1)
	arguments = append(arguments, pageSize+1)
	rows, err := s.pool.Query(ctx, query, arguments...)
	if err != nil {
		return domain.SharePage{}, fmt.Errorf("list share links: %w", err)
	}
	defer rows.Close()
	page := domain.SharePage{}
	for rows.Next() {
		share, err := scanShare(rows)
		if err != nil {
			return domain.SharePage{}, fmt.Errorf("scan share link: %w", err)
		}
		page.ShareLinks = append(page.ShareLinks, share)
	}
	if len(page.ShareLinks) > pageSize {
		last := page.ShareLinks[pageSize-1]
		page.ShareLinks = page.ShareLinks[:pageSize]
		page.NextPageToken = encodeSimpleCursor(last.UID)
	}
	return page, rows.Err()
}

// GetShareByTokenHash resolves an active share.
func (s *Store) GetShareByTokenHash(ctx context.Context, tokenHash []byte) (domain.ShareLink, error) {
	share, err := scanShare(s.pool.QueryRow(ctx, `SELECT uid, target_uid, create_time, expire_time, revoke_time
		FROM share_links WHERE token_hash = $1 AND revoke_time IS NULL AND expire_time > now()`, tokenHash))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ShareLink{}, fmt.Errorf("%w: active share", domain.ErrNotFound)
	}
	if err != nil {
		return domain.ShareLink{}, fmt.Errorf("get share: %w", err)
	}
	return share, nil
}

// RevokeShare disables an active share.
func (s *Store) RevokeShare(ctx context.Context, uid string) (domain.ShareLink, error) {
	share, err := scanShare(s.pool.QueryRow(ctx, `UPDATE share_links SET revoke_time = now()
		WHERE uid = $1 AND revoke_time IS NULL RETURNING uid, target_uid, create_time, expire_time, revoke_time`, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ShareLink{}, fmt.Errorf("%w: active share", domain.ErrNotFound)
	}
	if err != nil {
		return domain.ShareLink{}, fmt.Errorf("revoke share: %w", err)
	}
	return share, nil
}

// ListSharedItems lists children within a shared directory.
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
		return domain.Page{}, fmt.Errorf("%w: shared file has no children", domain.ErrConflict)
	}
	if parentUID == nil {
		parentUID = &share.TargetUID
	}
	inside, err := s.isWithin(ctx, share.TargetUID, *parentUID)
	if err != nil || !inside {
		if err != nil {
			return domain.Page{}, err
		}
		return domain.Page{}, fmt.Errorf("%w: item is outside share", domain.ErrForbidden)
	}
	return s.ListItems(ctx, domain.DriveListQuery{ParentUID: parentUID, PageSize: pageSize, PageToken: pageToken})
}

// CanReadSharedItem verifies share scope.
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
		return domain.Item{}, fmt.Errorf("%w: item is outside share", domain.ErrForbidden)
	}
	return s.GetItem(ctx, itemUID, false)
}

func (s *Store) getActiveShareByUID(ctx context.Context, uid string) (domain.ShareLink, error) {
	share, err := scanShare(s.pool.QueryRow(ctx, `SELECT uid, target_uid, create_time, expire_time, revoke_time
		FROM share_links WHERE uid = $1 AND revoke_time IS NULL AND expire_time > now()`, uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ShareLink{}, fmt.Errorf("%w: active share", domain.ErrNotFound)
	}
	if err != nil {
		return domain.ShareLink{}, fmt.Errorf("get share: %w", err)
	}
	return share, nil
}

func scanShare(row rowScanner) (domain.ShareLink, error) {
	var share domain.ShareLink
	var revokeTime pgtype.Timestamptz
	if err := row.Scan(&share.UID, &share.TargetUID, &share.CreateTime, &share.ExpireTime, &revokeTime); err != nil {
		return domain.ShareLink{}, err
	}
	if revokeTime.Valid {
		value := revokeTime.Time.UTC()
		share.RevokeTime = &value
	}
	return share, nil
}

type cursor struct{ name, uid string }
