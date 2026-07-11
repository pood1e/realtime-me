package app

import (
	"context"
	"errors"
	"io/fs"
	"time"

	"example.com/cloud-drive/api/internal/domain"
)

const (
	claimedUploadRetention = 7 * 24 * time.Hour
	contentDeleteDelay     = 48 * time.Hour
	storageOrphanDelay     = 48 * time.Hour
	storageReconcileBatch  = 256
)

// PurgeExpiredUploads removes temporary sessions past their expiry.
func (s *ContentService) PurgeExpiredUploads(ctx context.Context) error {
	now := s.clock.Now().UTC()
	uploads, err := s.uploads.ListDiscardableUploads(ctx, now, now.Add(-claimedUploadRetention))
	if err != nil {
		return err
	}
	for _, upload := range uploads {
		if err := s.files.RemoveUpload(ctx, upload.UID); err != nil {
			return err
		}
		if err := s.uploads.DeleteRetainedUpload(ctx, upload.UID); err != nil {
			return err
		}
	}
	return nil
}

// CollectGarbage deletes unreferenced content metadata and source bytes.
func (s *ContentService) CollectGarbage(ctx context.Context) error {
	for {
		objects, err := s.contents.ListUnreferencedContent(ctx, 100)
		if err != nil {
			return err
		}
		if len(objects) == 0 {
			break
		}
		for _, object := range objects {
			if err := s.contents.TombstoneContent(ctx, object.UID, s.clock.Now().UTC().Add(contentDeleteDelay)); err != nil {
				if errors.Is(err, domain.ErrConflict) {
					continue
				}
				return err
			}
		}
	}
	for {
		tombstones, err := s.contents.ListExpiredContentTombstones(ctx, s.clock.Now().UTC(), 100)
		if err != nil {
			return err
		}
		if len(tombstones) == 0 {
			break
		}
		for _, tombstone := range tombstones {
			for _, storageKey := range tombstone.StorageKeys {
				if err := s.files.Remove(ctx, storageKey); err != nil && !errors.Is(err, fs.ErrNotExist) {
					return err
				}
			}
			if err := s.contents.DeleteContentTombstone(ctx, tombstone.ContentUID); err != nil {
				return err
			}
		}
	}
	return nil
}

// ReconcileStorage removes old physical files that have no database owner.
// The age delay and filesystem recheck protect in-flight publishers.
func (s *ContentService) ReconcileStorage(ctx context.Context) error {
	cutoff := s.clock.Now().UTC().Add(-storageOrphanDelay)
	batch := make([]domain.StoredFile, 0, storageReconcileBatch)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		keys := make([]string, len(batch))
		for index, entry := range batch {
			keys[index] = entry.StorageKey
		}
		referenced, err := s.contents.ReferencedStorageKeys(ctx, keys)
		if err != nil {
			return err
		}
		known := make(map[string]struct{}, len(referenced))
		for _, key := range referenced {
			known[key] = struct{}{}
		}
		for _, entry := range batch {
			if _, exists := known[entry.StorageKey]; exists {
				continue
			}
			if err := s.files.RemoveStoredFileIfOlder(ctx, entry.StorageKey, cutoff); err != nil {
				return err
			}
		}
		batch = batch[:0]
		return nil
	}
	err := s.files.WalkStoredFiles(ctx, func(entry domain.StoredFile) error {
		if !entry.ActivityTime.Before(cutoff) {
			return nil
		}
		batch = append(batch, entry)
		if len(batch) == storageReconcileBatch {
			return flush()
		}
		return nil
	})
	if err != nil {
		return err
	}
	return flush()
}
