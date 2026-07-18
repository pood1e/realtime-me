//go:build !darwin && !linux

package storage

import (
	"os"
	"time"
)

func fileActivityTime(info os.FileInfo) time.Time {
	return info.ModTime().UTC()
}
