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
