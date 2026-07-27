package noteops_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/noteops"
	"github.com/andy-neoaira/obs-cli/pkg/pathpolicy"
	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

func newService(t *testing.T) (*noteops.Service, string) {
	t.Helper()
	root := t.TempDir()
	service, err := noteops.NewService(root, storage.NewStore(filepath.Join(t.TempDir(), "locks")))
	if err != nil {
		t.Fatal(err)
	}
	return service, root
}

func TestNoteCreateGetListAndNoOverwrite(t *testing.T) {
	service, root := newService(t)
	content := []byte("---\ntitle: Demo\n---\n# Demo\n")
	created, err := service.Create("Projects/demo", content)
	if err != nil {
		t.Fatal(err)
	}
	if created.Path != "Projects/demo.md" || created.RevisionAfter != storage.Revision(content) {
		t.Fatalf("unexpected create result: %#v", created)
	}
	got, err := service.Get("Projects/demo")
	if err != nil {
		t.Fatal(err)
	}
	if got.Content != string(content) || got.Frontmatter["title"] != "Demo" || got.Revision != created.RevisionAfter {
		t.Fatalf("unexpected note: %#v", got)
	}
	notes, err := service.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || notes[0] != "Projects/demo.md" {
		t.Fatalf("notes = %#v", notes)
	}
	if _, err := service.Create("Projects/demo", []byte("overwrite")); !errors.Is(err, storage.ErrAlreadyExists) {
		t.Fatalf("overwrite error = %v", err)
	}
	assertContent(t, filepath.Join(root, "Projects", "demo.md"), string(content))
}

func TestNoteListIgnoresHiddenFilesAndDirectories(t *testing.T) {
	service, root := newService(t)
	files := map[string]string{
		"visible.md":                    "visible",
		"Folder/nested.md":              "nested",
		".draft.md":                     "hidden root",
		"Folder/.private.md":            "hidden nested",
		".obsidian/workspace.md":        "hidden directory",
		"Folder/.private/cache/note.md": "hidden subtree",
	}
	for logical, content := range files {
		path := filepath.Join(root, filepath.FromSlash(logical))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	notes, err := service.List()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Folder/nested.md", "visible.md"}
	if len(notes) != len(want) {
		t.Fatalf("notes = %#v, want %#v", notes, want)
	}
	for index := range want {
		if notes[index] != want[index] {
			t.Fatalf("notes = %#v, want %#v", notes, want)
		}
	}
	if _, err := service.Get(".draft.md"); !errors.Is(err, pathpolicy.ErrOutsideVault) {
		t.Fatalf("explicit hidden get error = %v, want PATH_OUTSIDE_VAULT", err)
	}
}

func TestNoteAppendBoundaryAndSection(t *testing.T) {
	service, root := newService(t)
	if _, err := service.Create("demo", []byte("# Demo\nintro\n## Tasks\none\n## Later\nafter")); err != nil {
		t.Fatal(err)
	}
	appended, err := service.Append("demo", []byte("two"), "Tasks", "")
	if err != nil {
		t.Fatal(err)
	}
	want := "# Demo\nintro\n## Tasks\none\ntwo\n## Later\nafter"
	assertContent(t, filepath.Join(root, "demo.md"), want)
	if appended.RevisionAfter != storage.Revision([]byte(want)) {
		t.Fatalf("append result = %#v", appended)
	}
	if _, err := service.Append("demo", []byte("tail"), "", appended.RevisionAfter); err != nil {
		t.Fatal(err)
	}
	assertContent(t, filepath.Join(root, "demo.md"), want+"\ntail")
}

func TestNoteAppendSectionIgnoresHeadingsInsideFences(t *testing.T) {
	service, root := newService(t)
	content := "# Demo\n```md\n## Tasks\n```\n## Tasks\none\n## Later\nafter"
	if _, err := service.Create("demo", []byte(content)); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Append("demo", []byte("two"), "Tasks", ""); err != nil {
		t.Fatal(err)
	}
	want := "# Demo\n```md\n## Tasks\n```\n## Tasks\none\ntwo\n## Later\nafter"
	assertContent(t, filepath.Join(root, "demo.md"), want)
}

func TestNotePatchRequiresUniqueContextAndDoesNotModifyOnFailure(t *testing.T) {
	service, root := newService(t)
	content := []byte("alpha\nunique\nomega\n")
	created, err := service.Create("demo", content)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Patch("demo", []byte("missing"), []byte("new"), created.RevisionAfter); !errors.Is(err, noteops.ErrPatchContextMismatch) {
		t.Fatalf("missing context error = %v", err)
	}
	assertContent(t, filepath.Join(root, "demo.md"), string(content))

	ambiguous := []byte("same\nsame\n")
	replaced, err := service.Replace("demo", ambiguous, created.RevisionAfter)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Patch("demo", []byte("same"), []byte("new"), replaced.RevisionAfter); !errors.Is(err, noteops.ErrPatchContextAmbiguous) {
		t.Fatalf("ambiguous context error = %v", err)
	}
	assertContent(t, filepath.Join(root, "demo.md"), string(ambiguous))

	patched, err := service.Patch("demo", []byte("same\nsame"), []byte("only"), replaced.RevisionAfter)
	if err != nil {
		t.Fatal(err)
	}
	if patched.RevisionBefore != replaced.RevisionAfter {
		t.Fatalf("patch result = %#v", patched)
	}
	assertContent(t, filepath.Join(root, "demo.md"), "only\n")
}

func TestNoteReplaceDeleteRevisionAndDryRun(t *testing.T) {
	service, root := newService(t)
	path := filepath.Join(root, "demo.md")
	created, err := service.Create("demo", []byte("one"))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := service.PlanReplace("demo", []byte("two"), created.RevisionAfter)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Changed || plan.RevisionAfter != storage.Revision([]byte("two")) {
		t.Fatalf("replace plan = %#v", plan)
	}
	assertContent(t, path, "one")
	staleRevision := storage.Revision([]byte("stale"))
	if _, err := service.Replace("demo", []byte("stale"), staleRevision); !errors.Is(err, storage.ErrRevisionConflict) {
		t.Fatalf("replace conflict = %v", err)
	}
	replaced, err := service.Replace("demo", []byte("two"), created.RevisionAfter)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.PlanDelete("demo", created.RevisionAfter); !errors.Is(err, storage.ErrRevisionConflict) {
		t.Fatalf("delete conflict = %v", err)
	}
	deletePlan, err := service.PlanDelete("demo", replaced.RevisionAfter)
	if err != nil {
		t.Fatal(err)
	}
	assertContent(t, path, "two")
	if _, err := service.Delete("demo", deletePlan.RevisionBefore); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("note still exists: %v", err)
	}
}

func TestNotePlanCreateDoesNotCreateDirectories(t *testing.T) {
	service, root := newService(t)
	if _, err := service.PlanCreate("nested/demo", []byte("content")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "nested")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("plan created directory: %v", err)
	}
}

func TestNoteOperationsEnforceVaultPathAndSymlinkPolicy(t *testing.T) {
	service, root := newService(t)
	if _, err := service.PlanCreate("../escape", []byte("content")); !errors.Is(err, pathpolicy.ErrOutsideVault) {
		t.Fatalf("traversal error = %v", err)
	}
	target := filepath.Join(root, "target.md")
	if err := os.WriteFile(target, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.md")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	revision := storage.Revision([]byte("content"))
	if _, err := service.PlanReplace("link", []byte("changed"), revision); !errors.Is(err, pathpolicy.ErrOutsideVault) {
		t.Fatalf("symlink mutation error = %v", err)
	}
	assertContent(t, target, "content")
}

func TestNoteGetRejectsInvalidFrontmatter(t *testing.T) {
	service, _ := newService(t)
	if _, err := service.Create("demo", []byte("---\ninvalid: [\n---\nbody")); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Get("demo"); !errors.Is(err, noteops.ErrInvalidFrontmatter) {
		t.Fatalf("frontmatter error = %v", err)
	}
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want {
		t.Fatalf("content = %q, want %q", data, want)
	}
}
