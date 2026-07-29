package storage_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

func TestDeleteAtomicCreatesRecoveryAndChecksRevision(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "note.md")
	if err := os.WriteFile(target, []byte("recover me"), 0o644); err != nil {
		t.Fatal(err)
	}
	snapshot, err := storage.ReadSnapshot(target)
	if err != nil {
		t.Fatal(err)
	}
	store := storage.NewStore(filepath.Join(root, "runtime", "locks"))

	if _, err := store.DeleteAtomic(target, storage.Revision([]byte("stale"))); !errors.Is(err, storage.ErrRevisionConflict) {
		t.Fatalf("DeleteAtomic stale error = %v", err)
	}
	assertBytes(t, target, "recover me")

	result, err := store.DeleteAtomic(target, snapshot.Revision)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("deleted target still exists: %v", err)
	}
	assertBytes(t, result.RecoveryPath, "recover me")
	info, err := os.Stat(result.RecoveryPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("recovery mode = %o", info.Mode().Perm())
	}
}
