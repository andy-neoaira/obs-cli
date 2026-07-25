package noteops_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/noteops"
	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

func TestMoveRewritesStructuredLinksAndPreservesUnrelatedText(t *testing.T) {
	service, root := newService(t)
	source, err := service.Create("Folder/Old", []byte("# Old\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create("Refs/links", []byte(
		"[[Folder/Old#Heading|Alias]]\n[relative](../Folder/Old.md#part)\n"+
			"`[[Folder/Old]]`\nplain Folder/Old\n",
	)); err != nil {
		t.Fatal(err)
	}
	plan, err := service.PlanMove("Folder/Old", "Archive/New", source.RevisionAfter)
	if err != nil {
		t.Fatal(err)
	}
	replanned, err := service.PlanMove("Folder/Old", "Archive/New", source.RevisionAfter)
	if err != nil {
		t.Fatal(err)
	}
	if !storage.IsRevision(plan.PlanHash) || replanned.PlanHash != plan.PlanHash {
		t.Fatalf("move plan hash is not stable: first=%q second=%q", plan.PlanHash, replanned.PlanHash)
	}
	if len(plan.Changes) != 2 || len(plan.Changes[1].LinkEdits) != 2 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	result, err := service.ApplyMovePlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	if result.TransactionID == "" || result.RevisionAfter == "" || result.PlanHash != plan.PlanHash {
		t.Fatalf("unexpected result: %#v", result)
	}
	if _, err := os.Stat(filepath.Join(root, "Folder", "Old.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source still exists: %v", err)
	}
	assertContent(t, filepath.Join(root, "Archive", "New.md"), "# Old\n")
	assertContent(
		t,
		filepath.Join(root, "Refs", "links.md"),
		"[[Archive/New#Heading|Alias]]\n[relative](../Archive/New.md#part)\n"+
			"`[[Folder/Old]]`\nplain Folder/Old\n",
	)
}

func TestMovePlanRejectsTargetAndApplyRejectsExternalChange(t *testing.T) {
	service, root := newService(t)
	source, err := service.Create("Old", []byte("source"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create("Existing", []byte("target")); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PlanMove("Old", "Existing", source.RevisionAfter); !errors.Is(err, storage.ErrAlreadyExists) {
		t.Fatalf("existing target error = %v", err)
	}
	if _, err := service.Create("links", []byte("[[Old]]")); err != nil {
		t.Fatal(err)
	}
	plan, err := service.PlanMove("Old", "New", source.RevisionAfter)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "links.md"), []byte("external"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ApplyMovePlan(plan); !errors.Is(err, storage.ErrRevisionConflict) {
		t.Fatalf("external change error = %v", err)
	}
	assertContent(t, filepath.Join(root, "Old.md"), "source")
	assertContent(t, filepath.Join(root, "links.md"), "external")
	if _, err := os.Stat(filepath.Join(root, "New.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("target created after rejected plan: %v", err)
	}
}

func TestMoveRollbackRestoresEveryFile(t *testing.T) {
	root := t.TempDir()
	store := storage.NewStoreWithHooks(filepath.Join(root, "runtime", "locks"), storage.Hooks{
		Checkpoint: func(stage string) error {
			if stage == "transaction-commit:3:before" {
				return errors.New("injected commit failure")
			}
			return nil
		},
	})
	service, err := noteops.NewService(root, store)
	if err != nil {
		t.Fatal(err)
	}
	source, err := service.Create("Old", []byte("source"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create("links", []byte("[[Old]]")); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Move("Old", "New", source.RevisionAfter); err == nil {
		t.Fatal("expected injected failure")
	}
	assertContent(t, filepath.Join(root, "Old.md"), "source")
	assertContent(t, filepath.Join(root, "links.md"), "[[Old]]")
	if _, err := os.Stat(filepath.Join(root, "New.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("target survived rollback: %v", err)
	}
}

func TestMovePartialFailureUsesLogicalRecoveryPaths(t *testing.T) {
	root := t.TempDir()
	store := storage.NewStoreWithHooks(filepath.Join(root, "runtime", "locks"), storage.Hooks{
		Checkpoint: func(stage string) error {
			switch stage {
			case "transaction-commit:3:before":
				return errors.New("injected commit failure")
			case "transaction-rollback:2:before":
				return errors.New("injected rollback failure")
			}
			return nil
		},
	})
	service, err := noteops.NewService(root, store)
	if err != nil {
		t.Fatal(err)
	}
	source, _ := service.Create("Old", []byte("source"))
	_, _ = service.Create("links", []byte("[[Old]]"))
	_, err = service.Move("Old", "New", source.RevisionAfter)
	var partial *noteops.MovePartialFailure
	if !errors.As(err, &partial) {
		t.Fatalf("error = %v, want MovePartialFailure", err)
	}
	for _, paths := range [][]string{partial.Completed, partial.Failed, partial.RollbackFailed} {
		for _, value := range paths {
			if filepath.IsAbs(value) {
				t.Fatalf("absolute recovery path leaked: %q", value)
			}
		}
	}
}
