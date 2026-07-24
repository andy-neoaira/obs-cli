package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type DeleteResult struct {
	RevisionBefore string
	RecoveryPath   string
}

type MoveResult struct {
	Revision string
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

func (s *Store) MoveAtomic(source, target, expectedRevision string, mode os.FileMode) (MoveResult, error) {
	if expectedRevision == "" {
		return MoveResult{}, ErrPrecondition
	}
	locks, err := s.lockMany([]string{source, target})
	if err != nil {
		return MoveResult{}, err
	}
	defer releaseLocks(locks)

	sourceSnapshot, err := ReadSnapshot(source)
	if err != nil {
		return MoveResult{}, err
	}
	if sourceSnapshot.Revision != expectedRevision {
		return MoveResult{}, ErrRevisionConflict
	}
	if _, err := os.Lstat(target); err == nil {
		return MoveResult{}, ErrAlreadyExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return MoveResult{}, err
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return MoveResult{}, err
	}
	temp, err := os.CreateTemp(filepath.Dir(target), ".obs-write-move-*.tmp")
	if err != nil {
		return MoveResult{}, err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if mode == 0 {
		mode = sourceSnapshot.Mode
	}
	if err := temp.Chmod(mode.Perm()); err != nil {
		temp.Close()
		return MoveResult{}, err
	}
	if _, err := temp.Write(sourceSnapshot.Data); err != nil {
		temp.Close()
		return MoveResult{}, err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return MoveResult{}, err
	}
	if err := temp.Close(); err != nil {
		return MoveResult{}, err
	}

	current, err := ReadSnapshot(source)
	if err != nil || current.Revision != expectedRevision {
		if err == nil || errors.Is(err, os.ErrNotExist) {
			return MoveResult{}, ErrRevisionConflict
		}
		return MoveResult{}, err
	}
	if _, err := os.Lstat(target); err == nil {
		return MoveResult{}, ErrAlreadyExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return MoveResult{}, err
	}
	if err := s.checkpoint("move-commit-before"); err != nil {
		return MoveResult{}, err
	}
	if err := commitCreateNoReplace(tempPath, target); err != nil {
		if errors.Is(err, os.ErrExist) {
			return MoveResult{}, ErrAlreadyExists
		}
		return MoveResult{}, err
	}
	rollbackTarget := true
	defer func() {
		if rollbackTarget {
			_ = os.Remove(target)
		}
	}()
	if err := s.checkpoint("move-source-delete-before"); err != nil {
		return MoveResult{}, err
	}
	if err := os.Remove(source); err != nil {
		return MoveResult{}, err
	}
	// 源删除后移动已经提交；即使后续 durability/verification 报错，也保留
	// 完整目标，不能再由 defer 删除唯一剩余副本。
	rollbackTarget = false
	if err := syncDirectory(filepath.Dir(source)); err != nil {
		return MoveResult{}, err
	}
	if filepath.Dir(source) != filepath.Dir(target) {
		if err := syncDirectory(filepath.Dir(target)); err != nil {
			return MoveResult{}, err
		}
	}
	after, err := ReadSnapshot(target)
	if err != nil || after.Revision != expectedRevision {
		return MoveResult{}, fmt.Errorf("atomic move verification failed")
	}
	return MoveResult{Revision: after.Revision}, nil
}

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
	for index := len(locks) - 1; index >= 0; index-- {
		locks[index].release()
	}
}
