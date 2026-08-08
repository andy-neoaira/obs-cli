package storage_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

func TestWriteAtomicCreateReplaceAndConflict(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "note.md")
	store := storage.NewStore(filepath.Join(root, "locks"))

	created, err := store.WriteAtomic(
		target,
		[]byte("one"),
		storage.Preconditions{MustNotExist: true},
		storage.WriteOptions{Mode: 0o640},
	)
	if err != nil {
		t.Fatal(err)
	}
	if created.RevisionBefore != "" || created.RevisionAfter != storage.Revision([]byte("one")) || !created.Changed {
		t.Fatalf("unexpected create result: %#v", created)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("mode = %o, want 640", info.Mode().Perm())
	}

	if _, err := store.WriteAtomic(
		target,
		[]byte("must not replace"),
		storage.Preconditions{MustNotExist: true},
		storage.WriteOptions{},
	); !errors.Is(err, storage.ErrAlreadyExists) {
		t.Fatalf("create conflict error = %v", err)
	}
	assertBytes(t, target, "one")

	replaced, err := store.WriteAtomic(
		target,
		[]byte("two"),
		storage.Preconditions{ExpectedRevision: created.RevisionAfter},
		storage.WriteOptions{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if replaced.RevisionBefore != created.RevisionAfter || replaced.RevisionAfter != storage.Revision([]byte("two")) {
		t.Fatalf("unexpected replace result: %#v", replaced)
	}

	if _, err := store.WriteAtomic(
		target,
		[]byte("stale"),
		storage.Preconditions{ExpectedRevision: created.RevisionAfter},
		storage.WriteOptions{},
	); !errors.Is(err, storage.ErrRevisionConflict) {
		t.Fatalf("revision conflict error = %v", err)
	}
	assertBytes(t, target, "two")

	unchanged, err := store.WriteAtomic(
		target,
		[]byte("two"),
		storage.Preconditions{ExpectedRevision: replaced.RevisionAfter},
		storage.WriteOptions{},
	)
	if err != nil || unchanged.Changed || unchanged.RevisionAfter != replaced.RevisionAfter {
		t.Fatalf("unchanged write = %#v, %v", unchanged, err)
	}
	if _, err := store.WriteAtomic(target, nil, storage.Preconditions{}, storage.WriteOptions{}); !errors.Is(err, storage.ErrPrecondition) {
		t.Fatalf("missing precondition error = %v", err)
	}
}

func TestWriteAtomicPartialWriteFailureLeavesOldBytesAndCleansTemp(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "note.md")
	if err := os.WriteFile(target, []byte("old complete"), 0o644); err != nil {
		t.Fatal(err)
	}
	before, err := storage.ReadSnapshot(target)
	if err != nil {
		t.Fatal(err)
	}
	injected := errors.New("injected partial write")
	store := storage.NewStoreWithHooks(filepath.Join(root, "locks"), storage.Hooks{
		Write: func(file *os.File, data []byte) error {
			if _, err := file.Write(data[:len(data)/2]); err != nil {
				return err
			}
			return injected
		},
	})
	_, err = store.WriteAtomic(
		target,
		[]byte("new complete bytes"),
		storage.Preconditions{ExpectedRevision: before.Revision},
		storage.WriteOptions{},
	)
	if !errors.Is(err, injected) {
		t.Fatalf("WriteAtomic error = %v", err)
	}
	assertBytes(t, target, "old complete")
	assertNoTemps(t, root)
}

func TestWriteAtomicFailureInjectionCleanup(t *testing.T) {
	stages := []string{
		"temp-create-before",
		"temp-create-after",
		"write-after",
		"flush-after",
		"close-after",
		"precommit-before",
		"precommit-after",
		"commit-before",
	}
	for _, stage := range stages {
		t.Run(stage, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "note.md")
			if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
				t.Fatal(err)
			}
			before, err := storage.ReadSnapshot(target)
			if err != nil {
				t.Fatal(err)
			}
			injected := errors.New("injected " + stage)
			store := storage.NewStoreWithHooks(filepath.Join(root, "locks"), storage.Hooks{
				Checkpoint: func(current string) error {
					if current == stage {
						return injected
					}
					return nil
				},
			})
			_, err = store.WriteAtomic(
				target,
				[]byte("new"),
				storage.Preconditions{ExpectedRevision: before.Revision},
				storage.WriteOptions{},
			)
			if !errors.Is(err, injected) {
				t.Fatalf("WriteAtomic error = %v", err)
			}
			assertBytes(t, target, "old")
			assertNoTemps(t, root)
		})
	}
}

func TestWriteAtomicPostCommitFailureLeavesNewCompleteBytes(t *testing.T) {
	for _, stage := range []string{"commit-after", "verify-after"} {
		t.Run(stage, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "note.md")
			if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
				t.Fatal(err)
			}
			before, err := storage.ReadSnapshot(target)
			if err != nil {
				t.Fatal(err)
			}
			injected := errors.New("injected " + stage)
			store := storage.NewStoreWithHooks(filepath.Join(root, "locks"), storage.Hooks{
				Checkpoint: func(current string) error {
					if current == stage {
						return injected
					}
					return nil
				},
			})
			_, err = store.WriteAtomic(
				target,
				[]byte("new"),
				storage.Preconditions{ExpectedRevision: before.Revision},
				storage.WriteOptions{},
			)
			if !errors.Is(err, injected) {
				t.Fatalf("WriteAtomic error = %v", err)
			}
			assertBytes(t, target, "new")
			assertNoTemps(t, root)
		})
	}
}

func TestWriteAtomicPrecommitExternalChangeReturnsConflict(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "note.md")
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	before, err := storage.ReadSnapshot(target)
	if err != nil {
		t.Fatal(err)
	}
	store := storage.NewStoreWithHooks(filepath.Join(root, "locks"), storage.Hooks{
		Checkpoint: func(stage string) error {
			if stage == "precommit-before" {
				return os.WriteFile(target, []byte("external"), 0o644)
			}
			return nil
		},
	})
	_, err = store.WriteAtomic(
		target,
		[]byte("new"),
		storage.Preconditions{ExpectedRevision: before.Revision},
		storage.WriteOptions{},
	)
	if !errors.Is(err, storage.ErrRevisionConflict) {
		t.Fatalf("WriteAtomic error = %v, want REVISION_CONFLICT", err)
	}
	assertBytes(t, target, "external")
	assertNoTemps(t, root)
}

func TestWriteAtomicConcurrentExpectedRevisionHasOneWinner(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "note.md")
	if err := os.WriteFile(target, []byte("base"), 0o644); err != nil {
		t.Fatal(err)
	}
	before, err := storage.ReadSnapshot(target)
	if err != nil {
		t.Fatal(err)
	}
	store := storage.NewStore(filepath.Join(root, "locks"))

	var wait sync.WaitGroup
	results := make(chan error, 2)
	for _, value := range []string{"left", "right"} {
		wait.Add(1)
		go func(content string) {
			defer wait.Done()
			_, writeErr := store.WriteAtomic(
				target,
				[]byte(content),
				storage.Preconditions{ExpectedRevision: before.Revision},
				storage.WriteOptions{},
			)
			results <- writeErr
		}(value)
	}
	wait.Wait()
	close(results)

	successes := 0
	conflicts := 0
	for result := range results {
		if result == nil {
			successes++
		} else if errors.Is(result, storage.ErrRevisionConflict) {
			conflicts++
		} else {
			t.Fatalf("unexpected concurrent result: %v", result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
}

func TestReplaceAtomicCreateReplaceAndMode(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "config.json")
	store := storage.NewStore(filepath.Join(root, "locks"))

	if err := store.ReplaceAtomic(target, []byte("one"), storage.WriteOptions{Mode: 0o600}); err != nil {
		t.Fatal(err)
	}
	assertBytes(t, target, "one")
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}

	if err := store.ReplaceAtomic(target, []byte("two"), storage.WriteOptions{Mode: 0o600}); err != nil {
		t.Fatal(err)
	}
	assertBytes(t, target, "two")
	assertNoReplaceTemps(t, root)
}

func TestDefaultModes(t *testing.T) {
	root := t.TempDir()
	store := storage.NewStore(filepath.Join(root, "locks"))
	note := filepath.Join(root, "note.md")
	if _, err := store.WriteAtomic(note, []byte("note"), storage.Preconditions{MustNotExist: true}, storage.WriteOptions{}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(note)
	if err != nil || info.Mode().Perm() != 0o644 {
		t.Fatalf("default note mode = %v, %v", info, err)
	}
	config := filepath.Join(root, "config.json")
	if err := store.ReplaceAtomic(config, []byte("{}"), storage.WriteOptions{}); err != nil {
		t.Fatal(err)
	}
	info, err = os.Stat(config)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("default replacement mode = %v, %v", info, err)
	}
}

func TestAtomicOperationsReportFilesystemSetupErrors(t *testing.T) {
	root := t.TempDir()
	blocker := filepath.Join(root, "blocker")
	if err := os.WriteFile(blocker, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	badLockStore := storage.NewStore(filepath.Join(blocker, "locks"))
	if _, err := badLockStore.WriteAtomic(
		filepath.Join(root, "note.md"), []byte("note"),
		storage.Preconditions{MustNotExist: true}, storage.WriteOptions{},
	); err == nil {
		t.Fatal("write with unusable lock directory should fail")
	}
	if _, err := badLockStore.DeleteAtomic(filepath.Join(root, "note.md"), storage.Revision(nil)); err == nil {
		t.Fatal("delete with unusable lock directory should fail")
	}

	store := storage.NewStore(filepath.Join(root, "locks"))
	missingTarget := filepath.Join(root, "missing", "note.md")
	if _, err := store.WriteAtomic(
		missingTarget, []byte("note"), storage.Preconditions{MustNotExist: true}, storage.WriteOptions{},
	); err == nil {
		t.Fatal("write below missing parent should fail")
	}
	if err := store.ReplaceAtomic(missingTarget, []byte("config"), storage.WriteOptions{}); err == nil {
		t.Fatal("replace below missing parent should fail")
	}

	directoryTarget := filepath.Join(root, "directory-target")
	if err := os.Mkdir(directoryTarget, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceAtomic(directoryTarget, []byte("config"), storage.WriteOptions{}); err == nil {
		t.Fatal("replace over directory should fail")
	}
}

func TestReplaceAtomicFailureInjectionBeforeCommitPreservesOldBytes(t *testing.T) {
	for _, stage := range []string{
		"temp-create-before",
		"temp-create-after",
		"write-after",
		"flush-after",
		"close-after",
		"commit-before",
	} {
		t.Run(stage, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "config.json")
			if err := os.WriteFile(target, []byte("old"), 0o600); err != nil {
				t.Fatal(err)
			}
			injected := errors.New("injected " + stage)
			store := storage.NewStoreWithHooks(filepath.Join(root, "locks"), storage.Hooks{
				Checkpoint: func(current string) error {
					if current == stage {
						return injected
					}
					return nil
				},
			})
			err := store.ReplaceAtomic(target, []byte("new"), storage.WriteOptions{Mode: 0o600})
			if !errors.Is(err, injected) {
				t.Fatalf("ReplaceAtomic error = %v", err)
			}
			assertBytes(t, target, "old")
			assertNoReplaceTemps(t, root)
		})
	}
}

func TestReplaceAtomicPostCommitFailureLeavesNewCompleteBytes(t *testing.T) {
	for _, stage := range []string{
		"directory-sync-before",
		"directory-sync-after",
		"commit-after",
		"verify-after",
	} {
		t.Run(stage, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "config.json")
			if err := os.WriteFile(target, []byte("old"), 0o600); err != nil {
				t.Fatal(err)
			}
			injected := errors.New("injected " + stage)
			store := storage.NewStoreWithHooks(filepath.Join(root, "locks"), storage.Hooks{
				Checkpoint: func(current string) error {
					if current == stage {
						return injected
					}
					return nil
				},
			})
			err := store.ReplaceAtomic(target, []byte("new"), storage.WriteOptions{Mode: 0o600})
			if !errors.Is(err, injected) {
				t.Fatalf("ReplaceAtomic error = %v", err)
			}
			assertBytes(t, target, "new")
			assertNoReplaceTemps(t, root)
		})
	}
}

func TestReplaceAtomicPartialWriteFailureCleansTemp(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "config.json")
	if err := os.WriteFile(target, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	injected := errors.New("injected partial write")
	store := storage.NewStoreWithHooks(filepath.Join(root, "locks"), storage.Hooks{
		Write: func(file *os.File, data []byte) error {
			if _, err := file.Write(data[:len(data)/2]); err != nil {
				return err
			}
			return injected
		},
	})
	err := store.ReplaceAtomic(target, []byte("new complete"), storage.WriteOptions{Mode: 0o600})
	if !errors.Is(err, injected) {
		t.Fatalf("ReplaceAtomic error = %v", err)
	}
	assertBytes(t, target, "old")
	assertNoReplaceTemps(t, root)
}

func assertBytes(t *testing.T, path, expected string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != expected {
		t.Fatalf("file bytes = %q, want %q", data, expected)
	}
}

func assertNoTemps(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".obs-write-") {
			t.Fatalf("temporary file leaked: %s", entry.Name())
		}
	}
}

func assertNoReplaceTemps(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".obs-replace-") {
			t.Fatalf("temporary file leaked: %s", entry.Name())
		}
	}
}
