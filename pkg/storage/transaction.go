package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

var ErrPartialFailure = errors.New("PARTIAL_FAILURE")

type Mutation struct {
	Path         string
	Data         []byte
	Delete       bool
	Precondition Preconditions
	Mode         os.FileMode
}

type TransactionResult struct {
	ID        string
	Revisions map[string]string
}

type PartialFailureError struct {
	TransactionID   string
	Completed       []string
	Failed          []string
	RolledBack      []string
	RollbackFailed  []string
	RecoveryActions []string
	Cause           error
}

func (e *PartialFailureError) Error() string {
	return fmt.Sprintf("%s: transaction %s requires recovery", ErrPartialFailure, e.TransactionID)
}

func (e *PartialFailureError) Unwrap() error {
	return ErrPartialFailure
}

type stagedMutation struct {
	mutation  Mutation
	before    Snapshot
	existed   bool
	tempPath  string
	backup    string
	committed bool
}

type transactionJournal struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Paths  []string `json:"paths"`
}

func (s *Store) ApplyTransaction(mutations []Mutation) (TransactionResult, error) {
	if len(mutations) == 0 {
		return TransactionResult{}, errors.New("transaction requires mutations")
	}
	ordered := append([]Mutation(nil), mutations...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Path < ordered[j].Path })
	for index := 1; index < len(ordered); index++ {
		if ordered[index-1].Path == ordered[index].Path {
			return TransactionResult{}, fmt.Errorf("duplicate transaction path")
		}
	}
	paths := make([]string, len(ordered))
	for index := range ordered {
		paths[index] = ordered[index].Path
	}
	locks, err := s.lockMany(paths)
	if err != nil {
		return TransactionResult{}, err
	}
	defer releaseLocks(locks)

	id := transactionID()
	staged := make([]stagedMutation, 0, len(ordered))
	retainArtifacts := false
	cleanup := func() {
		if retainArtifacts {
			return
		}
		for _, item := range staged {
			_ = os.Remove(item.tempPath)
			_ = os.Remove(item.backup)
		}
	}
	defer cleanup()

	for _, mutation := range ordered {
		before, exists, err := snapshotIfExists(mutation.Path)
		if err != nil {
			return TransactionResult{}, err
		}
		if err := validatePrecondition(before, exists, mutation.Precondition); err != nil {
			return TransactionResult{}, err
		}
		if mutation.Delete && (!exists || mutation.Precondition.ExpectedRevision == "") {
			return TransactionResult{}, ErrPrecondition
		}
		var temp string
		if !mutation.Delete {
			temp, err = stageFile(mutation.Path, mutation.Data, mutation.Mode, before, exists)
			if err != nil {
				return TransactionResult{}, err
			}
		}
		item := stagedMutation{mutation: mutation, before: before, existed: exists, tempPath: temp}
		if exists {
			backup, err := s.writeRecovery(id, before.Data)
			if err != nil {
				return TransactionResult{}, err
			}
			item.backup = backup
		}
		staged = append(staged, item)
	}

	journalPath, err := s.writeJournal(transactionJournal{ID: id, Status: "staged", Paths: paths})
	if err != nil {
		return TransactionResult{}, err
	}
	removeJournal := true
	defer func() {
		if removeJournal {
			_ = os.Remove(journalPath)
		}
	}()

	for index := range staged {
		item := &staged[index]
		if err := s.checkpoint(fmt.Sprintf("transaction-commit:%d:before", index+1)); err != nil {
			rolledBack, rollbackFailed, rollbackErr := s.rollback(staged, index-1)
			if rollbackErr != nil {
				removeJournal = false
				retainArtifacts = true
				return TransactionResult{}, partialFailure(id, staged, index, err, rolledBack, rollbackFailed)
			}
			return TransactionResult{}, err
		}
		current, exists, err := snapshotIfExists(item.mutation.Path)
		if err != nil || validatePrecondition(current, exists, item.mutation.Precondition) != nil {
			rolledBack, rollbackFailed, rollbackErr := s.rollback(staged, index-1)
			if rollbackErr != nil {
				removeJournal = false
				retainArtifacts = true
				return TransactionResult{}, partialFailure(id, staged, index, ErrRevisionConflict, rolledBack, rollbackFailed)
			}
			return TransactionResult{}, ErrRevisionConflict
		}
		if item.mutation.Delete {
			err = os.Remove(item.mutation.Path)
		} else if item.mutation.Precondition.MustNotExist {
			err = commitCreateNoReplace(item.tempPath, item.mutation.Path)
		} else {
			err = replaceFile(item.tempPath, item.mutation.Path)
		}
		if err != nil {
			rolledBack, rollbackFailed, rollbackErr := s.rollback(staged, index-1)
			if rollbackErr != nil {
				removeJournal = false
				retainArtifacts = true
				return TransactionResult{}, partialFailure(id, staged, index, err, rolledBack, rollbackFailed)
			}
			return TransactionResult{}, err
		}
		item.committed = true
	}

	revisions := make(map[string]string, len(staged))
	for _, item := range staged {
		if item.mutation.Delete {
			if _, err := os.Lstat(item.mutation.Path); !errors.Is(err, os.ErrNotExist) {
				removeJournal = false
				retainArtifacts = true
				return TransactionResult{}, partialFailure(id, staged, len(staged), errors.New("deleted path still exists"), nil, nil)
			}
			revisions[item.mutation.Path] = ""
			continue
		}
		after, err := ReadSnapshot(item.mutation.Path)
		if err != nil || after.Revision != Revision(item.mutation.Data) {
			removeJournal = false
			retainArtifacts = true
			return TransactionResult{}, partialFailure(id, staged, len(staged), errors.New("transaction verification failed"), nil, nil)
		}
		revisions[item.mutation.Path] = after.Revision
	}
	return TransactionResult{ID: id, Revisions: revisions}, nil
}

func partialFailure(
	id string,
	staged []stagedMutation,
	failedIndex int,
	cause error,
	rolledBack, rollbackFailed []string,
) error {
	completed := make([]string, 0)
	for _, item := range staged {
		if item.committed {
			completed = append(completed, item.mutation.Path)
		}
	}
	failed := []string{}
	if failedIndex >= 0 && failedIndex < len(staged) {
		failed = append(failed, staged[failedIndex].mutation.Path)
	}
	return &PartialFailureError{
		TransactionID:   id,
		Completed:       completed,
		Failed:          failed,
		RolledBack:      rolledBack,
		RollbackFailed:  rollbackFailed,
		RecoveryActions: []string{"inspect the retained transaction journal and recovery copies before retrying"},
		Cause:           cause,
	}
}

func (s *Store) rollback(staged []stagedMutation, last int) ([]string, []string, error) {
	rolledBack := make([]string, 0)
	for index := last; index >= 0; index-- {
		item := staged[index]
		if !item.committed {
			continue
		}
		if err := s.checkpoint(fmt.Sprintf("transaction-rollback:%d:before", index+1)); err != nil {
			return rolledBack, []string{item.mutation.Path}, err
		}
		if item.existed {
			temp, err := stageFile(item.mutation.Path, item.before.Data, item.before.Mode, Snapshot{}, false)
			if err != nil {
				return rolledBack, []string{item.mutation.Path}, err
			}
			if err := replaceFile(temp, item.mutation.Path); err != nil {
				_ = os.Remove(temp)
				return rolledBack, []string{item.mutation.Path}, err
			}
		} else if err := os.Remove(item.mutation.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return rolledBack, []string{item.mutation.Path}, err
		}
		rolledBack = append(rolledBack, item.mutation.Path)
	}
	return rolledBack, []string{}, nil
}

func stageFile(path string, data []byte, mode os.FileMode, before Snapshot, exists bool) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(path), ".obs-write-txn-*.tmp")
	if err != nil {
		return "", err
	}
	name := file.Name()
	if exists {
		mode = before.Mode
	} else if mode == 0 {
		mode = 0o644
	}
	if err := file.Chmod(mode.Perm()); err != nil {
		file.Close()
		_ = os.Remove(name)
		return "", err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		_ = os.Remove(name)
		return "", err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		_ = os.Remove(name)
		return "", err
	}
	return name, file.Close()
}

func (s *Store) writeRecovery(id string, data []byte) (string, error) {
	dir := filepath.Join(filepath.Dir(s.lockDir), "recovery")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	file, err := os.CreateTemp(dir, id+"-*.bak")
	if err != nil {
		return "", err
	}
	name := file.Name()
	_ = file.Chmod(0o600)
	if _, err := file.Write(data); err != nil {
		file.Close()
		return name, err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return name, err
	}
	return name, file.Close()
}

func (s *Store) writeJournal(journal transactionJournal) (string, error) {
	dir := filepath.Join(filepath.Dir(s.lockDir), "journals")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	data, err := json.Marshal(journal)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, journal.ID+".json")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return "", err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	return path, nil
}

func transactionID() string {
	value := make([]byte, 12)
	if _, err := rand.Read(value); err != nil {
		return "txn-unavailable"
	}
	return "txn-" + hex.EncodeToString(value)
}
