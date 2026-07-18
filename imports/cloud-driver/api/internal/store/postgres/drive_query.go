package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"example.com/cloud-drive/api/internal/domain"
)

const (
	itemColumns = `item.uid, item.parent_uid, item.content_uid, item.name, item.kind,
        COALESCE(content.size_bytes, 0), COALESCE(content.content_type, ''), COALESCE(content.storage_key, ''),
        item.create_time, item.update_time, item.delete_time`
	itemFrom = `drive_items item LEFT JOIN content_objects content ON content.uid = item.content_uid`
)

func (s *Store) GetItem(ctx context.Context, uid string, includeTrashed bool) (domain.Item, error) {
	query := "SELECT " + itemColumns + " FROM " + itemFrom + " WHERE item.uid = $1"
	if !includeTrashed {
		query += " AND item.delete_time IS NULL"
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

// ListItems lists direct children with a stable cursor.
func (s *Store) ListItems(ctx context.Context, filter domain.DriveListQuery) (domain.Page, error) {
	cursor, err := decodeCursor(filter.PageToken)
	if err != nil {
		return domain.Page{}, err
	}
	pageSize := normalizePageSize(filter.PageSize)
	query := "SELECT " + itemColumns + " FROM " + itemFrom + " WHERE item.parent_uid IS NOT DISTINCT FROM $1"
	arguments := []any{nullableString(filter.ParentUID)}
	if !filter.IncludeTrashed {
		query += " AND item.delete_time IS NULL"
	}
	if cursor != nil {
		query += fmt.Sprintf(" AND (item.name, item.uid) > ($%d, $%d)", len(arguments)+1, len(arguments)+2)
		arguments = append(arguments, cursor.name, cursor.uid)
	}
	query += fmt.Sprintf(" ORDER BY item.name, item.uid LIMIT $%d", len(arguments)+1)
	arguments = append(arguments, pageSize+1)
	return s.queryItemPage(ctx, query, arguments, pageSize)
}

// ListTrashedItems lists trash roots.
func (s *Store) ListTrashedItems(ctx context.Context, pageSize int, pageToken string) (domain.Page, error) {
	cursor, err := decodeCursor(pageToken)
	if err != nil {
		return domain.Page{}, err
	}
	pageSize = normalizePageSize(pageSize)
	query := `SELECT ` + itemColumns + ` FROM ` + itemFrom + `
		WHERE item.delete_time IS NOT NULL AND (item.parent_uid IS NULL OR NOT EXISTS (
			SELECT 1 FROM drive_items parent WHERE parent.uid = item.parent_uid AND parent.delete_time IS NOT NULL))`
	arguments := []any{}
	if cursor != nil {
		query += fmt.Sprintf(" AND (item.name, item.uid) > ($%d, $%d)", len(arguments)+1, len(arguments)+2)
		arguments = append(arguments, cursor.name, cursor.uid)
	}
	query += fmt.Sprintf(" ORDER BY item.name, item.uid LIMIT $%d", len(arguments)+1)
	arguments = append(arguments, pageSize+1)
	return s.queryItemPage(ctx, query, arguments, pageSize)
}

// SearchItems searches visible basenames.
func (s *Store) SearchItems(ctx context.Context, queryText string, pageSize int, pageToken string) (domain.Page, error) {
	cursor, err := decodeCursor(pageToken)
	if err != nil {
		return domain.Page{}, err
	}
	pageSize = normalizePageSize(pageSize)
	query := "SELECT " + itemColumns + " FROM " + itemFrom + " WHERE item.delete_time IS NULL AND item.name ILIKE '%' || $1 || '%'"
	arguments := []any{queryText}
	if cursor != nil {
		query += fmt.Sprintf(" AND (item.name, item.uid) > ($%d, $%d)", len(arguments)+1, len(arguments)+2)
		arguments = append(arguments, cursor.name, cursor.uid)
	}
	query += fmt.Sprintf(" ORDER BY item.name, item.uid LIMIT $%d", len(arguments)+1)
	arguments = append(arguments, pageSize+1)
	return s.queryItemPage(ctx, query, arguments, pageSize)
}

// CreateDirectory creates a visible directory.
func (s *Store) queryItemPage(ctx context.Context, query string, arguments []any, pageSize int) (domain.Page, error) {
	rows, err := s.pool.Query(ctx, query, arguments...)
	if err != nil {
		return domain.Page{}, fmt.Errorf("query drive items: %w", err)
	}
	defer rows.Close()
	page := domain.Page{}
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return domain.Page{}, fmt.Errorf("scan drive item: %w", err)
		}
		page.Items = append(page.Items, item)
	}
	if len(page.Items) > pageSize {
		last := page.Items[pageSize-1]
		page.Items = page.Items[:pageSize]
		page.NextPageToken = encodeCursor(last.Name, last.UID)
	}
	return page, rows.Err()
}
