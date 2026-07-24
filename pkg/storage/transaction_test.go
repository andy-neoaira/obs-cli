package storage_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

func TestTransactionCommitAndRollbackInjection(t *testing.T) {
	t.Run("commit", func(t *testing.T) {
		root := t.TempDir()
		first, second := transactionFiles(t, root)
		store := storage.NewStore(filepath.Join(root, "runtime", "locks"))
		result, err := store.ApplyTransaction(transactionMutations(t, first, second))
		if err != nil {
			t.Fatal(err)
		}
		if result.ID == "" || len(result.Revisions) != 2 {
			t.Fatalf("unexpected transaction result: %#v", result)
		}
		assertBytes(t, first, "new-one")
		assertBytes(t, second, "new-two")
		assertRuntimeClean(t, root)
	})

	t.Run("second commit failure rolls back first", func(t *testing.T) {
		root := t.TempDir()
		first, second := transactionFiles(t, root)
		injected := errors.New("commit two failed")
		store := storage.NewStoreWithHooks(filepath.Join(root, "runtime", "locks"), storage.Hooks{
			Checkpoint: func(stage string) error {
				if stage == "transaction-commit:2:before" {
					return injected
				}
				return nil
			},
		})
		_, err := store.ApplyTransaction(transactionMutations(t, first, second))
		if !errors.Is(err, injected) {
			t.Fatalf("ApplyTransaction error = %v", err)
		}
		assertBytes(t, first, "old-one")
		assertBytes(t, second, "old-two")
		assertRuntimeClean(t, root)
	})

	t.Run("rollback failure retains recovery artifacts", func(t *testing.T) {
		root := t.TempDir()
		first, second := transactionFiles(t, root)
		store := storage.NewStoreWithHooks(filepath.Join(root, "runtime", "locks"), storage.Hooks{
			Checkpoint: func(stage string) error {
				switch stage {
				case "transaction-commit:2:before":
					return errors.New("commit failed")
				case "transaction-rollback:1:before":
					return errors.New("rollback failed")
				}
				return nil
			},
		})
		_, err := store.ApplyTransaction(transactionMutations(t, first, second))
		if !errors.Is(err, storage.ErrPartialFailure) {
			t.Fatalf("ApplyTransaction error = %v, want PARTIAL_FAILURE", err)
		}
		assertBytes(t, first, "new-one")
		assertBytes(t, second, "old-two")
		journals, _ := filepath.Glob(filepath.Join(root, "runtime", "journals", "*.json"))
		backups, _ := filepath.Glob(filepath.Join(root, "runtime", "recovery", "*.bak"))
		if len(journals) != 1 || len(backups) != 2 {
			t.Fatalf("recovery artifacts journals=%v backups=%v", journals, backups)
		}
	})
}

func transactionFiles(t *testing.T, root string) (string, string) {
	t.Helper()
	first := filepath.Join(root, "one.md")
	second := filepath.Join(root, "two.md")
	if err := os.WriteFile(first, []byte("old-one"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("old-two"), 0o644); err != nil {
		t.Fatal(err)
	}
	return first, second
}

func transactionMutations(t *testing.T, first, second string) []storage.Mutation {
	t.Helper()
	firstSnapshot, err := storage.ReadSnapshot(first)
	if err != nil {
		t.Fatal(err)
	}
	secondSnapshot, err := storage.ReadSnapshot(second)
	if err != nil {
		t.Fatal(err)
	}
	return []storage.Mutation{
		{Path: first, Data: []byte("new-one"), Precondition: storage.Preconditions{ExpectedRevision: firstSnapshot.Revision}},
		{Path: second, Data: []byte("new-two"), Precondition: storage.Preconditions{ExpectedRevision: secondSnapshot.Revision}},
	}
}

func assertRuntimeClean(t *testing.T, root string) {
	t.Helper()
	journals, _ := filepath.Glob(filepath.Join(root, "runtime", "journals", "*.json"))
	backups, _ := filepath.Glob(filepath.Join(root, "runtime", "recovery", "*.bak"))
	if len(journals) != 0 || len(backups) != 0 {
		t.Fatalf("runtime artifacts leaked journals=%v backups=%v", journals, backups)
	}
}
