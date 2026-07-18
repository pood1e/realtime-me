package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"example.com/cloud-drive/api/internal/app"
	"example.com/cloud-drive/api/internal/config"
	"example.com/cloud-drive/api/internal/storage"
	"example.com/cloud-drive/api/internal/store/postgres"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()
	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.LoadMigration()
	if err != nil {
		return err
	}
	store, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer store.Close()
	if err := store.Migrate(ctx); err != nil {
		return err
	}
	files, err := storage.NewFilesystem(cfg.DataRoot)
	if err != nil {
		return err
	}
	migration := app.NewContentMigrationService(store, files)
	if err := migration.MigrateLegacyContent(ctx); err != nil {
		return err
	}
	_, err = store.AdoptDriveBooks(ctx)
	return err
}
