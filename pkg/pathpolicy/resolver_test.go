package pathpolicy_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/pathpolicy"
)

func TestResolverRejectsTraversalAndAbsolutePaths(t *testing.T) {
	resolver, err := pathpolicy.NewResolver(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	inputs := []string{
		"../outside.md",
		"folder/../../outside.md",
		"/etc/passwd",
		`C:\Windows\system.ini`,
		`\\server\share\note.md`,
		"~/note.md",
		"folder//note.md",
		"./note.md",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			_, resolveErr := resolver.Resolve(input, pathpolicy.ResolveOptions{})
			if !errors.Is(resolveErr, pathpolicy.ErrOutsideVault) {
				t.Fatalf("Resolve(%q) error = %v, want PATH_OUTSIDE_VAULT", input, resolveErr)
			}
		})
	}
}

func TestResolverRejectsExternalSymlinkForExistingAndNewTarget(t *testing.T) {
	root := t.TempDir()
	vault := filepath.Join(root, "vault")
	outside := filepath.Join(root, "outside")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.md"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(vault, "escape")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	resolver, err := pathpolicy.NewResolver(vault)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"escape/secret.md", "escape/new/deep.md"} {
		_, resolveErr := resolver.Resolve(input, pathpolicy.ResolveOptions{})
		if !errors.Is(resolveErr, pathpolicy.ErrOutsideVault) {
			t.Fatalf("Resolve(%q) error = %v, want PATH_OUTSIDE_VAULT", input, resolveErr)
		}
	}
}

func TestResolverAllowsInternalSymlinkUnicodeSpacesAndMissingParents(t *testing.T) {
	vault := t.TempDir()
	target := filepath.Join(vault, "真实 目录")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(vault, "入口")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	resolver, err := pathpolicy.NewResolver(vault)
	if err != nil {
		t.Fatal(err)
	}
	result, err := resolver.Resolve("入口/嵌套 目录/笔记.md", pathpolicy.ResolveOptions{})
	if err != nil {
		t.Fatal(err)
	}
	canonicalTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(canonicalTarget, "嵌套 目录", "笔记.md")
	if result.Path != want || result.Logical != "入口/嵌套 目录/笔记.md" || result.Exists || !result.ThroughSymlink {
		t.Fatalf("Resolve result = %#v, want path %q and non-existing logical target", result, want)
	}
}

func TestResolverHiddenPolicyAndRootCapability(t *testing.T) {
	resolver, err := pathpolicy.NewResolver(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.Resolve(".obsidian/app.json", pathpolicy.ResolveOptions{}); !errors.Is(err, pathpolicy.ErrOutsideVault) {
		t.Fatalf("hidden path error = %v", err)
	}
	if _, err := resolver.Resolve(".custom/note.md", pathpolicy.ResolveOptions{AllowHidden: true}); err != nil {
		t.Fatalf("audited hidden path should be allowed: %v", err)
	}
	if _, err := resolver.Resolve("", pathpolicy.ResolveOptions{}); !errors.Is(err, pathpolicy.ErrOutsideVault) {
		t.Fatalf("empty path error = %v", err)
	}
	result, err := resolver.Resolve("", pathpolicy.ResolveOptions{AllowRoot: true})
	if err != nil || result.Path != resolver.Root() {
		t.Fatalf("root capability result = %#v, error = %v", result, err)
	}
}

func TestResolverDoesNotLeakExternalSymlinkTarget(t *testing.T) {
	root := t.TempDir()
	vault := filepath.Join(root, "vault")
	outside := filepath.Join(root, "private-target")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(vault, "escape")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	resolver, err := pathpolicy.NewResolver(vault)
	if err != nil {
		t.Fatal(err)
	}
	_, resolveErr := resolver.Resolve("escape/note.md", pathpolicy.ResolveOptions{})
	if resolveErr == nil {
		t.Fatal("expected symlink escape error")
	}
	if contains := filepath.Base(outside); contains != "" && stringContains(resolveErr.Error(), contains) {
		t.Fatalf("error leaked external target: %v", resolveErr)
	}
}

func stringContains(value, fragment string) bool {
	for index := 0; index+len(fragment) <= len(value); index++ {
		if value[index:index+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
