package actions_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/mocks"
	"github.com/andy-neoaira/obs-cli/pkg/actions"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/andy-neoaira/obs-cli/pkg/pathpolicy"
)

func TestCreateDeleteAndMoveRejectSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	vaultPath := filepath.Join(root, "vault")
	outsidePath := filepath.Join(root, "outside")
	if err := os.MkdirAll(vaultPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outsidePath, 0o755); err != nil {
		t.Fatal(err)
	}
	externalNote := filepath.Join(outsidePath, "victim.md")
	if err := os.WriteFile(externalNote, []byte("do not change"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsidePath, filepath.Join(vaultPath, "escape")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := os.Symlink(externalNote, filepath.Join(vaultPath, "victim.md")); err != nil {
		t.Skipf("file symlink unavailable: %v", err)
	}

	vault := &mocks.MockVaultOperator{Name: "vault", PathValue: vaultPath}
	createErr := actions.CreateNote(vault, &mocks.MockUriManager{}, actions.CreateParams{
		NoteName: "escape/new-note",
		Content:  "escaped",
	})
	if !errors.Is(createErr, pathpolicy.ErrOutsideVault) {
		t.Fatalf("CreateNote error = %v, want PATH_OUTSIDE_VAULT", createErr)
	}

	deleteErr := actions.DeleteNote(vault, &obsidian.Note{}, actions.DeleteParams{NotePath: "victim"})
	if !errors.Is(deleteErr, pathpolicy.ErrOutsideVault) {
		t.Fatalf("DeleteNote error = %v, want PATH_OUTSIDE_VAULT", deleteErr)
	}

	moveErr := actions.MoveNote(
		vault,
		&obsidian.Note{},
		&obsidian.LinkRewriter{},
		&mocks.MockUriManager{},
		actions.MoveParams{CurrentNoteName: "victim", NewNoteName: "renamed"},
	)
	if !errors.Is(moveErr, pathpolicy.ErrOutsideVault) {
		t.Fatalf("MoveNote error = %v, want PATH_OUTSIDE_VAULT", moveErr)
	}

	content, err := os.ReadFile(externalNote)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "do not change" {
		t.Fatalf("external file changed: %q", content)
	}
	if _, err := os.Stat(filepath.Join(outsidePath, "new-note.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("create escaped Vault: %v", err)
	}
}
