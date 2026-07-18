//go:build linux

package storage

import (
	"os"
	"syscall"
	"time"
)

func fileActivityTime(info os.FileInfo) time.Time {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return info.ModTime().UTC()
	}
	changed := stat.Ctim
	changeTime := time.Unix(changed.Sec, changed.Nsec).UTC()
	if changeTime.After(info.ModTime()) {
		return changeTime
	}
	return info.ModTime().UTC()
}
