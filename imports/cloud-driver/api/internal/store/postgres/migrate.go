package postgres

import (
	"context"
	"crypto/sha256"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrate applies embedded, append-only schema migrations.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	connection, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer connection.Release()

	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock(341019146727)"); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	defer connection.Exec(context.Background(), "SELECT pg_advisory_unlock(341019146727)")

	if _, err := connection.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		checksum TEXT,
		applied_time TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}
	if _, err := connection.Exec(ctx, "ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS checksum TEXT"); err != nil {
		return fmt.Errorf("upgrade migration table: %w", err)
	}

	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Slice(entries, func(left, right int) bool { return entries[left].Name() < entries[right].Name() })
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		source, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		version := strings.TrimSuffix(entry.Name(), ".sql")
		checksum := fmt.Sprintf("%x", sha256.Sum256(source))
		var storedChecksum *string
		err = connection.QueryRow(ctx, "SELECT checksum FROM schema_migrations WHERE version = $1", version).Scan(&storedChecksum)
		if err == nil {
			if storedChecksum == nil {
				if _, err := connection.Exec(ctx, "UPDATE schema_migrations SET checksum = $2 WHERE version = $1", version, checksum); err != nil {
					return fmt.Errorf("record migration %s checksum: %w", version, err)
				}
				continue
			}
			if *storedChecksum != checksum {
				return fmt.Errorf("migration %s checksum mismatch", version)
			}
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		transaction, err := connection.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", version, err)
		}
		if _, err := transaction.Exec(ctx, string(source)); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if _, err := transaction.Exec(ctx, "INSERT INTO schema_migrations (version, checksum) VALUES ($1, $2)", version, checksum); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := transaction.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}
	}
	return nil
}
