package storage

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"example.com/cloud-drive/api/internal/domain"
)

func (f *Filesystem) inspectFile(path, fileName string) (domain.SealedContent, error) {
	file, err := os.Open(path)
	if err != nil {
		return domain.SealedContent{}, fmt.Errorf("open source file: %w", err)
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return domain.SealedContent{}, fmt.Errorf("hash source file: %w", err)
	}
	digest := hash.Sum(nil)
	hexDigest := hex.EncodeToString(digest)
	contentType, err := detectContentType(path, fileName)
	if err != nil {
		return domain.SealedContent{}, err
	}
	return domain.SealedContent{
		SHA256:      digest,
		SizeBytes:   size,
		ContentType: contentType,
		StorageKey:  filepath.ToSlash(filepath.Join("objects", "sha256", hexDigest[:2], hexDigest[2:4], hexDigest)),
	}, nil
}

func (f *Filesystem) publishFile(sourcePath, storageKey string, removeSource bool) error {
	targetPath, err := f.storagePath(storageKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o700); err != nil {
		return fmt.Errorf("create storage directory: %w", err)
	}
	if _, err := os.Stat(targetPath); err == nil {
		if removeSource {
			return removeFile(sourcePath, "published source")
		}
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("stat storage target: %w", err)
	}
	if err := os.Link(sourcePath, targetPath); err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("publish stored file: %w", err)
	}
	if err := syncDirectory(filepath.Dir(targetPath)); err != nil {
		return err
	}
	if removeSource {
		return removeFile(sourcePath, "published source")
	}
	return nil
}

func (f *Filesystem) uploadPath(uploadUID string) (string, error) {
	if err := validateSegment(uploadUID); err != nil {
		return "", err
	}
	return filepath.Join(f.uploads, uploadUID+".part"), nil
}

func (f *Filesystem) storagePath(storageKey string) (string, error) {
	if storageKey == "" || filepath.IsAbs(storageKey) || strings.ContainsRune(storageKey, '\x00') {
		return "", errors.New("unsafe storage key")
	}
	clean := filepath.Clean(filepath.FromSlash(storageKey))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", errors.New("unsafe storage key")
	}
	return f.pathWithin(f.dataRoot, filepath.Join(f.dataRoot, clean))
}

func (f *Filesystem) pathWithin(root, path string) (string, error) {
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("path escapes storage root")
	}
	return path, nil
}

func validateSegment(value string) error {
	if value == "" || value == "." || value == ".." || filepath.Base(value) != value || containsPathSeparatorOrNUL(value) {
		return errors.New("unsafe storage segment")
	}
	return nil
}

func containsPathSeparatorOrNUL(value string) bool {
	return strings.ContainsAny(value, `/\`) || strings.ContainsRune(value, '\x00')
}

func detectContentType(path, fileName string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open source for type detection: %w", err)
	}
	defer file.Close()
	sample := make([]byte, 4096)
	read, err := file.Read(sample)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read source for type detection: %w", err)
	}
	sample = sample[:read]
	detected := http.DetectContentType(sample)
	extension := strings.ToLower(filepath.Ext(fileName))
	if extension == ".epub" && validEPUB(path) {
		return "application/epub+zip", nil
	}
	if extension == ".svg" && bytes.Contains(bytes.ToLower(sample), []byte("<svg")) {
		return "image/svg+xml", nil
	}
	if detected != "application/octet-stream" && !strings.HasPrefix(detected, "text/plain") {
		return strings.TrimSpace(strings.Split(detected, ";")[0]), nil
	}
	if value := mime.TypeByExtension(extension); value != "" {
		return strings.TrimSpace(strings.Split(value, ";")[0]), nil
	}
	return "application/octet-stream", nil
}

func validEPUB(path string) bool {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return false
	}
	defer archive.Close()
	for _, entry := range archive.File {
		if entry.Name != "mimetype" {
			continue
		}
		file, err := entry.Open()
		if err != nil {
			return false
		}
		body, err := io.ReadAll(io.LimitReader(file, 64))
		file.Close()
		return err == nil && strings.TrimSpace(string(body)) == "application/epub+zip"
	}
	return false
}

func removeFile(path, label string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", label, err)
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open directory for sync: %w", err)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync directory: %w", err)
	}
	return nil
}
