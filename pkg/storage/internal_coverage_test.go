package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInternalStagingRecoveryAndJournalHelpers(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "note.md")
	temp, err := stageFile(target, []byte("new"), 0, Snapshot{}, false)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(temp)
	info, err := os.Stat(temp)
	if err != nil || info.Mode().Perm() != 0o644 {
		t.Fatalf("staged mode = %v, %v", info, err)
	}

	tempExisting, err := stageFile(target, []byte("old"), 0o777, Snapshot{Mode: 0o600}, true)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempExisting)
	info, err = os.Stat(tempExisting)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("existing staged mode = %v, %v", info, err)
	}

	store := NewStore(filepath.Join(root, "runtime", "locks"))
	backup, err := store.writeRecovery("txn-test", []byte("backup"))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(backup)
	journal := transactionJournal{ID: "txn-test", Status: "staged", Paths: []string{target}}
	journalPath, err := store.writeJournal(journal)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(journalPath)
	if _, err := store.writeJournal(journal); err == nil {
		t.Fatal("duplicate journal should fail")
	}
}

func TestInternalStorageErrorBranches(t *testing.T) {
	partial := &PartialFailureError{TransactionID: "txn-test"}
	if !strings.Contains(partial.Error(), "txn-test") || !errors.Is(partial, ErrPartialFailure) {
		t.Fatalf("partial failure methods failed: %v", partial)
	}
	if DefaultStore() == nil {
		t.Fatal("DefaultStore returned nil")
	}

	root := t.TempDir()
	blocker := filepath.Join(root, "runtime")
	if err := os.WriteFile(blocker, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := NewStore(filepath.Join(blocker, "locks"))
	if _, err := store.lockOne(filepath.Join(root, "note.md")); err == nil {
		t.Fatal("lockOne with blocked runtime should fail")
	}
	if _, err := store.lockMany([]string{filepath.Join(root, "a.md")}); err == nil {
		t.Fatal("lockMany with blocked runtime should fail")
	}
	if _, err := store.writeRecovery("txn", []byte("data")); err == nil {
		t.Fatal("recovery directory creation should fail")
	}
	if _, err := store.writeJournal(transactionJournal{ID: "txn"}); err == nil {
		t.Fatal("journal directory creation should fail")
	}
	if _, err := stageFile(filepath.Join(root, "missing", "note.md"), nil, 0, Snapshot{}, false); err == nil {
		t.Fatal("staging in a missing directory should fail")
	}
	if _, err := acquireFileLock(filepath.Join(root, "missing-locks", "note.lock")); err == nil {
		t.Fatal("locking in a missing directory should fail")
	}
	if err := syncDirectory(filepath.Join(root, "missing-directory")); err == nil {
		t.Fatal("syncing a missing directory should fail")
	}
	source := filepath.Join(root, "source.tmp")
	target := filepath.Join(root, "target.md")
	if err := os.WriteFile(source, []byte("source"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("target"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := commitCreateNoReplace(source, target); !errors.Is(err, os.ErrExist) {
		t.Fatalf("no-replace conflict error = %v", err)
	}
}
