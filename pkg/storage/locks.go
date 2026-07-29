package storage

import (
	"os"
	"path/filepath"
	"sort"
)

func (s *Store) lockOne(path string) (*fileLock, error) {
	if err := os.MkdirAll(s.lockDir, 0o700); err != nil {
		return nil, err
	}
	return acquireFileLock(filepath.Join(s.lockDir, lockName(path)))
}

func (s *Store) lockMany(paths []string) ([]*fileLock, error) {
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)

	locks := make([]*fileLock, 0, len(sorted))
	for _, path := range sorted {
		lock, err := s.lockOne(path)
		if err != nil {
			releaseLocks(locks)
			return nil, err
		}
		locks = append(locks, lock)
	}
	return locks, nil
}

func releaseLocks(locks []*fileLock) {
	for i := len(locks) - 1; i >= 0; i-- {
		locks[i].release()
	}
}
