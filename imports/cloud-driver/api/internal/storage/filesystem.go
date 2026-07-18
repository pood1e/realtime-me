package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"example.com/cloud-drive/api/internal/domain"
)

// Filesystem stores temporary uploads, immutable content, and derived artifacts.
type Filesystem struct {
	dataRoot  string
	uploads   string
	objects   string
	artifacts string
	work      string
}

// NewFilesystem initializes private storage directories.
func NewFilesystem(dataRoot string) (*Filesystem, error) {
	filesystem := &Filesystem{
		dataRoot:  dataRoot,
		uploads:   filepath.Join(dataRoot, "uploads"),
		objects:   filepath.Join(dataRoot, "objects", "sha256"),
		artifacts: filepath.Join(dataRoot, "artifacts"),
		work:      filepath.Join(dataRoot, "work"),
	}
	for _, directory := range []string{filesystem.uploads, filesystem.objects, filesystem.artifacts, filesystem.work} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return nil, fmt.Errorf("create storage directory %q: %w", directory, err)
		}
	}
	return filesystem, nil
}

// PrepareUpload creates an empty private temporary file.
func (f *Filesystem) PrepareUpload(_ context.Context, uploadUID string) error {
	path, err := f.uploadPath(uploadUID)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("create upload temporary file: %w", err)
	}
	return file.Close()
}

// WriteChunk writes bytes at their exact upload offset.
func (f *Filesystem) WriteChunk(_ context.Context, uploadUID string, offset int64, data []byte) error {
	path, err := f.uploadPath(uploadUID)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open upload temporary file: %w", err)
	}
	defer file.Close()
	written, err := file.WriteAt(data, offset)
	if err != nil {
		return fmt.Errorf("write upload chunk: %w", err)
	}
	if written != len(data) {
		return fmt.Errorf("write upload chunk: wrote %d of %d bytes", written, len(data))
	}
	return file.Sync()
}

// SealUpload publishes a complete upload under its SHA-256 content address.
func (f *Filesystem) SealUpload(ctx context.Context, uploadUID, fileName string) (domain.SealedContent, error) {
	path, err := f.uploadPath(uploadUID)
	if err != nil {
		return domain.SealedContent{}, err
	}
	sealed, err := f.inspectFile(ctx, path, fileName)
	if err != nil {
		return domain.SealedContent{}, err
	}
	if err := f.publishFile(path, sealed.StorageKey, false); err != nil {
		return domain.SealedContent{}, err
	}
	return sealed, nil
}

// InspectLegacyObject hashes a legacy object and publishes its content-addressed link.
func (f *Filesystem) InspectLegacyObject(ctx context.Context, content domain.ContentObject) (domain.SealedContent, error) {
	path, err := f.storagePath(content.StorageKey)
	if err != nil {
		return domain.SealedContent{}, err
	}
	sealed, err := f.inspectFile(ctx, path, filepath.Base(content.StorageKey))
	if err != nil {
		return domain.SealedContent{}, err
	}
	if err := f.publishFile(path, sealed.StorageKey, false); err != nil {
		return domain.SealedContent{}, err
	}
	return sealed, nil
}

// RemoveUpload deletes an unclaimed temporary upload.
func (f *Filesystem) RemoveUpload(_ context.Context, uploadUID string) error {
	path, err := f.uploadPath(uploadUID)
	if err != nil {
		return err
	}
	return removeFile(path, "upload temporary file")
}

// Open opens immutable source content or an artifact by relative storage key.
func (f *Filesystem) Open(_ context.Context, storageKey string) (*os.File, error) {
	path, err := f.storagePath(storageKey)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("open stored file: %w", err)
	}
	return file, nil
}

// Remove deletes one stored file after metadata authorization.
func (f *Filesystem) Remove(_ context.Context, storageKey string) error {
	path, err := f.storagePath(storageKey)
	if err != nil {
		return err
	}
	return removeFile(path, "stored file")
}

// WalkStoredFiles visits source objects and derived artifacts without following links.
func (f *Filesystem) WalkStoredFiles(ctx context.Context, visit func(domain.StoredFile) error) error {
	for _, root := range []string{f.objects, f.artifacts} {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			relative, err := filepath.Rel(f.dataRoot, path)
			if err != nil {
				return fmt.Errorf("resolve stored file key: %w", err)
			}
			return visit(domain.StoredFile{
				StorageKey:   filepath.ToSlash(relative),
				ActivityTime: fileActivityTime(info),
			})
		})
		if err != nil {
			return fmt.Errorf("walk stored files: %w", err)
		}
	}
	return nil
}

// RemoveStoredFileIfOlder rechecks age under an advisory file lock before deletion.
func (f *Filesystem) RemoveStoredFileIfOlder(_ context.Context, storageKey string, cutoff time.Time) error {
	path, err := f.storagePath(storageKey)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open stored file for reconciliation: %w", err)
	}
	defer file.Close()
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("lock stored file for reconciliation: %w", err)
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat stored file for reconciliation: %w", err)
	}
	current, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("recheck stored file for reconciliation: %w", err)
	}
	if !os.SameFile(info, current) || !fileActivityTime(info).Before(cutoff) {
		return nil
	}
	return removeFile(path, "orphaned stored file")
}

// NewWorkDir creates an isolated directory for one processing job.
func (f *Filesystem) NewWorkDir(jobUID string) (string, error) {
	if err := validateSegment(jobUID); err != nil {
		return "", fmt.Errorf("invalid processing job id: %w", err)
	}
	path := filepath.Join(f.work, jobUID)
	if err := os.RemoveAll(path); err != nil {
		return "", fmt.Errorf("reset processing directory: %w", err)
	}
	if err := os.Mkdir(path, 0o700); err != nil {
		return "", fmt.Errorf("create processing directory: %w", err)
	}
	return path, nil
}

// RemoveWorkDir removes a processing job's temporary directory.
func (f *Filesystem) RemoveWorkDir(path string) error {
	clean, err := f.pathWithin(f.work, path)
	if err != nil {
		return err
	}
	if clean == f.work {
		return errors.New("refusing to remove work root")
	}
	return os.RemoveAll(clean)
}

// PublishArtifact atomically installs a generated artifact file.
func (f *Filesystem) PublishArtifact(sourcePath string, contentSHA256 []byte, kind, variant, extension string) (string, error) {
	if len(contentSHA256) != sha256.Size {
		return "", errors.New("invalid content hash")
	}
	if err := validateSegment(kind); err != nil {
		return "", fmt.Errorf("invalid artifact kind: %w", err)
	}
	if err := validateSegment(variant); err != nil {
		return "", fmt.Errorf("invalid artifact variant: %w", err)
	}
	extension = strings.TrimPrefix(strings.ToLower(extension), ".")
	if extension == "" || containsPathSeparatorOrNUL(extension) {
		return "", errors.New("unsafe artifact extension")
	}
	contentAddress := hex.EncodeToString(contentSHA256)
	storageKey := filepath.ToSlash(filepath.Join("artifacts", contentAddress, kind, variant+"."+extension))
	if err := f.replaceFile(sourcePath, storageKey); err != nil {
		return "", err
	}
	return storageKey, nil
}

// PublishSource installs a worker-created source file in content-addressed storage.
func (f *Filesystem) PublishSource(ctx context.Context, sourcePath, fileName string) (domain.SealedContent, error) {
	sealed, err := f.inspectFile(ctx, sourcePath, fileName)
	if err != nil {
		return domain.SealedContent{}, err
	}
	if err := f.publishFile(sourcePath, sealed.StorageKey, false); err != nil {
		return domain.SealedContent{}, err
	}
	return sealed, nil
}

// FreeBytes returns currently available bytes on the data filesystem.
func (f *Filesystem) FreeBytes() (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(f.dataRoot, &stat); err != nil {
		return 0, fmt.Errorf("stat storage filesystem: %w", err)
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}
