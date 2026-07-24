//go:build !windows

package storage

import (
	"fmt"
	"os"
	"syscall"
)

type fileLock struct {
	file *os.File
}

func acquireFileLock(path string) (*fileLock, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return nil, fmt.Errorf("acquire write lock: %w", err)
	}
	return &fileLock{file: file}, nil
}

func (l *fileLock) release() {
	_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	_ = l.file.Close()
}

func replaceFile(source, target string) error {
	return os.Rename(source, target)
}

func syncDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
