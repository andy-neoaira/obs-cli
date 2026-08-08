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

func TestTransactionCreateRewriteDelete(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "old.md")
	link := filepath.Join(root, "links.md")
	target := filepath.Join(root, "new.md")
	if err := os.WriteFile(source, []byte("source"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(link, []byte("[[old]]"), 0o644); err != nil {
		t.Fatal(err)
	}
	sourceSnapshot, _ := storage.ReadSnapshot(source)
	linkSnapshot, _ := storage.ReadSnapshot(link)
	store := storage.NewStore(filepath.Join(root, "runtime", "locks"))
	result, err := store.ApplyTransaction([]storage.Mutation{
		{Path: target, Data: sourceSnapshot.Data, Precondition: storage.Preconditions{MustNotExist: true}, Mode: sourceSnapshot.Mode},
		{Path: link, Data: []byte("[[new]]"), Precondition: storage.Preconditions{ExpectedRevision: linkSnapshot.Revision}},
		{Path: source, Delete: true, Precondition: storage.Preconditions{ExpectedRevision: sourceSnapshot.Revision}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Revisions[source] != "" || result.Revisions[target] != sourceSnapshot.Revision {
		t.Fatalf("unexpected revisions: %#v", result.Revisions)
	}
	if _, err := os.Stat(source); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source exists: %v", err)
	}
	assertBytes(t, target, "source")
	assertBytes(t, link, "[[new]]")
}

func TestTransactionRejectsInvalidMutationSets(t *testing.T) {
	root := t.TempDir()
	store := storage.NewStore(filepath.Join(root, "runtime", "locks"))
	if _, err := store.ApplyTransaction(nil); err == nil {
		t.Fatal("empty transaction should fail")
	}

	path := filepath.Join(root, "note.md")
	if _, err := store.ApplyTransaction([]storage.Mutation{
		{Path: path, Data: []byte("one"), Precondition: storage.Preconditions{MustNotExist: true}},
		{Path: path, Data: []byte("two"), Precondition: storage.Preconditions{MustNotExist: true}},
	}); err == nil {
		t.Fatal("duplicate transaction path should fail")
	}
	if _, err := store.ApplyTransaction([]storage.Mutation{{
		Path: path, Delete: true, Precondition: storage.Preconditions{MustNotExist: true},
	}}); !errors.Is(err, storage.ErrAlreadyExists) && !errors.Is(err, storage.ErrPrecondition) {
		t.Fatalf("invalid delete precondition error = %v", err)
	}
	if _, err := store.ApplyTransaction([]storage.Mutation{{
		Path: path, Data: []byte("update"), Precondition: storage.Preconditions{ExpectedRevision: storage.Revision(nil)},
	}}); !errors.Is(err, storage.ErrRevisionConflict) {
		t.Fatalf("missing update target error = %v", err)
	}
	missingParent := filepath.Join(root, "missing", "new.md")
	if _, err := store.ApplyTransaction([]storage.Mutation{{
		Path: missingParent, Data: []byte("create"), Precondition: storage.Preconditions{MustNotExist: true},
	}}); err == nil {
		t.Fatal("staging below a missing parent should fail")
	}
}

func TestTransactionRollbackRemovesNewlyCreatedFile(t *testing.T) {
	root := t.TempDir()
	created := filepath.Join(root, "a-new.md")
	existing := filepath.Join(root, "z-existing.md")
	if err := os.WriteFile(existing, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	existingSnapshot, err := storage.ReadSnapshot(existing)
	if err != nil {
		t.Fatal(err)
	}
	injected := errors.New("second commit failed")
	store := storage.NewStoreWithHooks(filepath.Join(root, "runtime", "locks"), storage.Hooks{
		Checkpoint: func(stage string) error {
			if stage == "transaction-commit:2:before" {
				return injected
			}
			return nil
		},
	})
	_, err = store.ApplyTransaction([]storage.Mutation{
		{Path: created, Data: []byte("new"), Precondition: storage.Preconditions{MustNotExist: true}},
		{Path: existing, Data: []byte("changed"), Precondition: storage.Preconditions{ExpectedRevision: existingSnapshot.Revision}},
	})
	if !errors.Is(err, injected) {
		t.Fatalf("ApplyTransaction error = %v", err)
	}
	if _, err := os.Stat(created); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("created file survived rollback: %v", err)
	}
	assertBytes(t, existing, "old")
}

func TestTransactionDetectsLateConflictCommitFailureAndVerificationFailure(t *testing.T) {
	t.Run("late conflict rolls back prior commit", func(t *testing.T) {
		root := t.TempDir()
		first, second := transactionFiles(t, root)
		mutations := transactionMutations(t, first, second)
		store := storage.NewStoreWithHooks(filepath.Join(root, "runtime", "locks"), storage.Hooks{
			Checkpoint: func(stage string) error {
				if stage == "transaction-commit:1:before" {
					return os.WriteFile(second, []byte("external"), 0o644)
				}
				return nil
			},
		})
		if _, err := store.ApplyTransaction(mutations); !errors.Is(err, storage.ErrRevisionConflict) {
			t.Fatalf("late conflict error = %v", err)
		}
		assertBytes(t, first, "old-one")
		assertBytes(t, second, "external")
	})

	t.Run("filesystem commit error", func(t *testing.T) {
		root := t.TempDir()
		first, second := transactionFiles(t, root)
		store := storage.NewStoreWithHooks(filepath.Join(root, "runtime", "locks"), storage.Hooks{
			Checkpoint: func(stage string) error {
				if stage == "transaction-commit:1:before" {
					if err := os.Remove(first); err != nil {
						return err
					}
					return os.Mkdir(first, 0o755)
				}
				return nil
			},
		})
		if _, err := store.ApplyTransaction(transactionMutations(t, first, second)); err == nil {
			t.Fatal("commit over a directory should fail")
		}
	})

	t.Run("post-commit verification detects external edit", func(t *testing.T) {
		root := t.TempDir()
		first, second := transactionFiles(t, root)
		store := storage.NewStoreWithHooks(filepath.Join(root, "runtime", "locks"), storage.Hooks{
			Checkpoint: func(stage string) error {
				if stage == "transaction-commit:2:before" {
					return os.WriteFile(first, []byte("external-after-commit"), 0o644)
				}
				return nil
			},
		})
		if _, err := store.ApplyTransaction(transactionMutations(t, first, second)); !errors.Is(err, storage.ErrPartialFailure) {
			t.Fatalf("verification error = %v", err)
		}
		assertBytes(t, first, "external-after-commit")
		assertBytes(t, second, "new-two")
	})
}

func TestTransactionJournalCreationFailureCleansStagedArtifacts(t *testing.T) {
	root := t.TempDir()
	first, _ := transactionFiles(t, root)
	runtimeRoot := filepath.Join(root, "runtime")
	if err := os.MkdirAll(runtimeRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runtimeRoot, "journals"), []byte("blocker"), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err := storage.ReadSnapshot(first)
	if err != nil {
		t.Fatal(err)
	}
	store := storage.NewStore(filepath.Join(runtimeRoot, "locks"))
	if _, err := store.ApplyTransaction([]storage.Mutation{{
		Path: first, Data: []byte("new"), Precondition: storage.Preconditions{ExpectedRevision: snapshot.Revision},
	}}); err == nil {
		t.Fatal("blocked journal directory should fail")
	}
	assertBytes(t, first, "old-one")
}

func TestSnapshotRejectsNonRegularFile(t *testing.T) {
	if _, err := storage.ReadSnapshot(t.TempDir()); err == nil {
		t.Fatal("directory snapshot should fail")
	}
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
