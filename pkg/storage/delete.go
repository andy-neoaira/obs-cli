package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type DeleteResult struct {
	RevisionBefore string
	RecoveryPath   string
}

func (s *Store) DeleteAtomic(path, expectedRevision string) (DeleteResult, error) {
	if expectedRevision == "" {
		return DeleteResult{}, ErrPrecondition
	}
	lock, err := s.lockOne(path)
	if err != nil {
		return DeleteResult{}, err
	}
	defer lock.release()

	before, err := ReadSnapshot(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DeleteResult{}, ErrRevisionConflict
		}
		return DeleteResult{}, err
	}
	if before.Revision != expectedRevision {
		return DeleteResult{}, ErrRevisionConflict
	}

	recoveryDir := filepath.Join(filepath.Dir(s.lockDir), "recovery")
	if err := os.MkdirAll(recoveryDir, 0o700); err != nil {
		return DeleteResult{}, err
	}
	recovery, err := os.CreateTemp(recoveryDir, "deleted-*.bak")
	if err != nil {
		return DeleteResult{}, err
	}
	recoveryPath := recovery.Name()
	cleanupRecovery := true
	defer func() {
		if cleanupRecovery {
			_ = os.Remove(recoveryPath)
		}
	}()
	if err := recovery.Chmod(0o600); err != nil {
		recovery.Close()
		return DeleteResult{}, err
	}
	if _, err := recovery.Write(before.Data); err != nil {
		recovery.Close()
		return DeleteResult{}, err
	}
	if err := recovery.Sync(); err != nil {
		recovery.Close()
		return DeleteResult{}, err
	}
	if err := recovery.Close(); err != nil {
		return DeleteResult{}, err
	}

	current, err := ReadSnapshot(path)
	if err != nil || current.Revision != expectedRevision {
		if err == nil || errors.Is(err, os.ErrNotExist) {
			return DeleteResult{}, ErrRevisionConflict
		}
		return DeleteResult{}, err
	}
	if err := s.checkpoint("delete-before"); err != nil {
		return DeleteResult{}, err
	}
	if err := os.Remove(path); err != nil {
		return DeleteResult{}, err
	}
	// 从此处开始删除已经提交；后续验证失败时必须保留恢复副本。
	cleanupRecovery = false
	if err := syncDirectory(filepath.Dir(path)); err != nil {
		return DeleteResult{}, err
	}
	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		return DeleteResult{}, fmt.Errorf("atomic delete verification failed")
	}
	return DeleteResult{RevisionBefore: before.Revision, RecoveryPath: recoveryPath}, nil
}
