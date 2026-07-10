package storage

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Filesystem stores temporary uploads and immutable completed blobs.
type Filesystem struct {
	dataRoot string
	uploads  string
	blobs    string
}

// NewFilesystem initializes the required private storage directories.
func NewFilesystem(dataRoot string) (*Filesystem, error) {
	filesystem := &Filesystem{
		dataRoot: dataRoot,
		uploads:  filepath.Join(dataRoot, "uploads"),
		blobs:    filepath.Join(dataRoot, "blobs"),
	}
	for _, directory := range []string{filesystem.uploads, filesystem.blobs} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return nil, fmt.Errorf("create storage directory %q: %w", directory, err)
		}
	}
	return filesystem, nil
}

// PrepareUpload creates an empty private temporary file for an upload session.
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

// FinalizeUpload atomically publishes a complete temporary file as an immutable blob.
func (f *Filesystem) FinalizeUpload(_ context.Context, uploadUID, itemUID string) error {
	temporaryPath, err := f.uploadPath(uploadUID)
	if err != nil {
		return err
	}
	blobPath, err := f.blobPath(itemUID)
	if err != nil {
		return err
	}
	if _, err := os.Stat(blobPath); err == nil {
		return f.syncBlobDirectory()
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("stat final blob: %w", err)
	}
	if err := os.Rename(temporaryPath, blobPath); err != nil {
		return fmt.Errorf("publish upload blob: %w", err)
	}
	return f.syncBlobDirectory()
}

// RemoveUpload deletes a temporary upload file after failed initialization.
func (f *Filesystem) RemoveUpload(_ context.Context, uploadUID string) error {
	path, err := f.uploadPath(uploadUID)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove upload temporary file: %w", err)
	}
	return nil
}

// OpenBlob opens an immutable blob for a range-capable HTTP response.
func (f *Filesystem) OpenBlob(_ context.Context, storageKey string) (*os.File, error) {
	path, err := f.blobPath(storageKey)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("open blob: %w", err)
	}
	return file, nil
}

// RemoveBlob removes an immutable blob during retention cleanup.
func (f *Filesystem) RemoveBlob(_ context.Context, storageKey string) error {
	path, err := f.blobPath(storageKey)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove blob: %w", err)
	}
	return nil
}

// FreeBytes returns currently available bytes on the data filesystem.
func (f *Filesystem) FreeBytes() (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(f.dataRoot, &stat); err != nil {
		return 0, fmt.Errorf("stat storage filesystem: %w", err)
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}

func (f *Filesystem) uploadPath(uploadUID string) (string, error) {
	if err := validateKey(uploadUID); err != nil {
		return "", err
	}
	return filepath.Join(f.uploads, uploadUID+".part"), nil
}

func (f *Filesystem) blobPath(storageKey string) (string, error) {
	if err := validateKey(storageKey); err != nil {
		return "", err
	}
	return filepath.Join(f.blobs, storageKey), nil
}

func (f *Filesystem) syncBlobDirectory() error {
	directory, err := os.Open(f.blobs)
	if err != nil {
		return fmt.Errorf("open blob directory for sync: %w", err)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync blob directory: %w", err)
	}
	return nil
}

func validateKey(value string) error {
	if value == "" || value == "." || value == ".." || filepath.Base(value) != value || strings.ContainsRune(value, '\x00') {
		return errors.New("unsafe storage key")
	}
	return nil
}
