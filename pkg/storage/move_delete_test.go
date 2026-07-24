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

func TestMoveAtomicSuccessAndRollbackBeforeSourceDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		root := t.TempDir()
		source := filepath.Join(root, "old.md")
		target := filepath.Join(root, "nested", "new.md")
		if err := os.WriteFile(source, []byte("move me"), 0o640); err != nil {
			t.Fatal(err)
		}
		snapshot, err := storage.ReadSnapshot(source)
		if err != nil {
			t.Fatal(err)
		}
		store := storage.NewStore(filepath.Join(root, "runtime", "locks"))
		result, err := store.MoveAtomic(source, target, snapshot.Revision, 0)
		if err != nil {
			t.Fatal(err)
		}
		if result.Revision != snapshot.Revision {
			t.Fatalf("move revision = %s, want %s", result.Revision, snapshot.Revision)
		}
		if _, err := os.Lstat(source); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("source still exists: %v", err)
		}
		assertBytes(t, target, "move me")
	})

	t.Run("rollback", func(t *testing.T) {
		root := t.TempDir()
		source := filepath.Join(root, "old.md")
		target := filepath.Join(root, "new.md")
		if err := os.WriteFile(source, []byte("stay"), 0o644); err != nil {
			t.Fatal(err)
		}
		snapshot, err := storage.ReadSnapshot(source)
		if err != nil {
			t.Fatal(err)
		}
		injected := errors.New("stop before source delete")
		store := storage.NewStoreWithHooks(filepath.Join(root, "runtime", "locks"), storage.Hooks{
			Checkpoint: func(stage string) error {
				if stage == "move-source-delete-before" {
					return injected
				}
				return nil
			},
		})
		_, err = store.MoveAtomic(source, target, snapshot.Revision, 0)
		if !errors.Is(err, injected) {
			t.Fatalf("MoveAtomic error = %v", err)
		}
		assertBytes(t, source, "stay")
		if _, err := os.Lstat(target); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("rollback target exists: %v", err)
		}
		assertNoTemps(t, root)
	})
}
