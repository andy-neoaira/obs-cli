package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrRevisionConflict = errors.New("REVISION_CONFLICT")
	ErrAlreadyExists    = errors.New("ALREADY_EXISTS")
	ErrPrecondition     = errors.New("write precondition is required")
)

type Preconditions struct {
	MustNotExist     bool
	ExpectedRevision string
}

type WriteOptions struct {
	Mode os.FileMode
}

type WriteResult struct {
	RevisionBefore string
	RevisionAfter  string
	Changed        bool
}

type Hooks struct {
	Checkpoint func(stage string) error
	Write      func(file *os.File, data []byte) error
}

type Store struct {
	lockDir string
	hooks   Hooks
}

func DefaultStore() *Store {
	return NewStore(filepath.Join(os.TempDir(), "obs-cli-runtime", "locks"))
}

func NewStore(lockDir string) *Store {
	return &Store{lockDir: lockDir}
}

func NewStoreWithHooks(lockDir string, hooks Hooks) *Store {
	return &Store{lockDir: lockDir, hooks: hooks}
}

func (s *Store) WriteAtomic(path string, data []byte, precondition Preconditions, options WriteOptions) (WriteResult, error) {
	if precondition.MustNotExist == (precondition.ExpectedRevision != "") {
		return WriteResult{}, ErrPrecondition
	}
	if err := os.MkdirAll(s.lockDir, 0o700); err != nil {
		return WriteResult{}, fmt.Errorf("create write lock directory: %w", err)
	}
	lock, err := acquireFileLock(filepath.Join(s.lockDir, lockName(path)))
	if err != nil {
		return WriteResult{}, err
	}
	defer lock.release()

	before, exists, err := snapshotIfExists(path)
	if err != nil {
		return WriteResult{}, err
	}
	if err := validatePrecondition(before, exists, precondition); err != nil {
		return WriteResult{}, err
	}
	if exists && before.Revision == Revision(data) {
		return WriteResult{
			RevisionBefore: before.Revision,
			RevisionAfter:  before.Revision,
			Changed:        false,
		}, nil
	}

	if err := s.checkpoint("temp-create-before"); err != nil {
		return WriteResult{}, err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".obs-write-*.tmp")
	if err != nil {
		return WriteResult{}, fmt.Errorf("create atomic write temp: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := s.checkpoint("temp-create-after"); err != nil {
		temp.Close()
		return WriteResult{}, err
	}

	mode := options.Mode.Perm()
	if exists {
		mode = before.Mode
	} else if mode == 0 {
		mode = 0o644
	}
	if err := temp.Chmod(mode); err != nil {
		temp.Close()
		return WriteResult{}, err
	}
	if err := s.write(temp, data); err != nil {
		temp.Close()
		return WriteResult{}, err
	}
	if err := s.checkpoint("write-after"); err != nil {
		temp.Close()
		return WriteResult{}, err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return WriteResult{}, err
	}
	if err := s.checkpoint("flush-after"); err != nil {
		temp.Close()
		return WriteResult{}, err
	}
	if err := temp.Close(); err != nil {
		return WriteResult{}, err
	}
	if err := s.checkpoint("close-after"); err != nil {
		return WriteResult{}, err
	}

	if err := s.checkpoint("precommit-before"); err != nil {
		return WriteResult{}, err
	}
	current, currentExists, err := snapshotIfExists(path)
	if err != nil {
		return WriteResult{}, err
	}
	if err := validatePrecondition(current, currentExists, precondition); err != nil {
		return WriteResult{}, err
	}
	if err := s.checkpoint("precommit-after"); err != nil {
		return WriteResult{}, err
	}
	if err := s.checkpoint("commit-before"); err != nil {
		return WriteResult{}, err
	}
	if precondition.MustNotExist {
		if err := commitCreateNoReplace(tempPath, path); err != nil {
			if errors.Is(err, os.ErrExist) {
				return WriteResult{}, ErrAlreadyExists
			}
			return WriteResult{}, err
		}
	} else if err := replaceFile(tempPath, path); err != nil {
		return WriteResult{}, err
	}
	if err := syncDirectory(filepath.Dir(path)); err != nil {
		return WriteResult{}, err
	}
	if err := s.checkpoint("commit-after"); err != nil {
		return WriteResult{}, err
	}

	after, err := ReadSnapshot(path)
	if err != nil {
		return WriteResult{}, err
	}
	wantRevision := Revision(data)
	if after.Revision != wantRevision {
		return WriteResult{}, fmt.Errorf("atomic write verification failed")
	}
	if err := s.checkpoint("verify-after"); err != nil {
		return WriteResult{}, err
	}
	return WriteResult{
		RevisionBefore: before.Revision,
		RevisionAfter:  after.Revision,
		Changed:        true,
	}, nil
}

func (s *Store) checkpoint(stage string) error {
	if s.hooks.Checkpoint == nil {
		return nil
	}
	return s.hooks.Checkpoint(stage)
}

func (s *Store) write(file *os.File, data []byte) error {
	if s.hooks.Write != nil {
		return s.hooks.Write(file, data)
	}
	_, err := file.Write(data)
	return err
}

func snapshotIfExists(path string) (Snapshot, bool, error) {
	snapshot, err := ReadSnapshot(path)
	if errors.Is(err, os.ErrNotExist) {
		return Snapshot{}, false, nil
	}
	return snapshot, err == nil, err
}

func validatePrecondition(snapshot Snapshot, exists bool, precondition Preconditions) error {
	if precondition.MustNotExist {
		if exists {
			return ErrAlreadyExists
		}
		return nil
	}
	if !exists || snapshot.Revision != precondition.ExpectedRevision {
		return ErrRevisionConflict
	}
	return nil
}

func lockName(path string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(path)))
	return hex.EncodeToString(sum[:]) + ".lock"
}

func commitCreateNoReplace(tempPath, targetPath string) error {
	if err := os.Link(tempPath, targetPath); err != nil {
		return err
	}
	return os.Remove(tempPath)
}
