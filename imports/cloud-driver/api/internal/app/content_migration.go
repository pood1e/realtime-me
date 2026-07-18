package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"example.com/cloud-drive/api/internal/domain"
)

// ContentMigrationService performs offline legacy object convergence.
type ContentMigrationService struct {
	contents domain.ContentStore
	files    ContentMigrationFiles
}

// NewContentMigrationService constructs the standalone migration boundary.
func NewContentMigrationService(contents domain.ContentStore, files ContentMigrationFiles) *ContentMigrationService {
	return &ContentMigrationService{contents: contents, files: files}
}

// MigrateLegacyContent converts all pre-suite blobs to content-addressed storage.
func (s *ContentMigrationService) MigrateLegacyContent(ctx context.Context) error {
	for {
		objects, err := s.contents.ListUnhashedContent(ctx, 100)
		if err != nil {
			return err
		}
		if len(objects) == 0 {
			break
		}
		for _, object := range objects {
			sealed, err := s.files.InspectLegacyObject(ctx, object)
			if err != nil {
				return fmt.Errorf("migrate content %s: %w", object.UID, err)
			}
			if _, err := s.contents.CommitContentMigration(ctx, object.UID, sealed); err != nil {
				return err
			}
			if object.StorageKey != sealed.StorageKey {
				if err := s.files.Remove(ctx, object.StorageKey); err != nil && !errors.Is(err, fs.ErrNotExist) {
					return err
				}
			}
		}
	}
	return s.contents.FinalizeContentMigration(ctx)
}
