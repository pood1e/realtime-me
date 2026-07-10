package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

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
		applied_time TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
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
		version := strings.TrimSuffix(entry.Name(), ".sql")
		var applied bool
		if err := connection.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if applied {
			continue
		}
		source, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}
		transaction, err := connection.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", version, err)
		}
		if _, err := transaction.Exec(ctx, string(source)); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if _, err := transaction.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := transaction.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}
	}
	return nil
}
