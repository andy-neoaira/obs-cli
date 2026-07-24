package obsidian_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/andy-neoaira/obs-cli/pkg/pathpolicy"
	"github.com/stretchr/testify/assert"
)

func TestLinkRewriterRejectsExternalMarkdownSymlink(t *testing.T) {
	root := t.TempDir()
	vault := filepath.Join(root, "vault")
	outside := filepath.Join(root, "outside.md")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("[[oldNote]]"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(vault, "linked.md")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	err := (&obsidian.LinkRewriter{}).UpdateLinks(vault, "oldNote", "newNote")
	if !errors.Is(err, pathpolicy.ErrOutsideVault) {
		t.Fatalf("UpdateLinks error = %v, want PATH_OUTSIDE_VAULT", err)
	}
	content, readErr := os.ReadFile(outside)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "[[oldNote]]" {
		t.Fatalf("external Markdown changed: %q", content)
	}
}

func TestLinkRewriterRejectsInternalMarkdownAliasMutation(t *testing.T) {
	vault := t.TempDir()
	target := filepath.Join(vault, "Inside.md")
	if err := os.WriteFile(target, []byte("[[oldNote]]"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(vault, "InternalAlias.md")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	err := (&obsidian.LinkRewriter{}).UpdateLinks(vault, "oldNote", "newNote")
	if !errors.Is(err, obsidian.ErrPhysicalPathConflict) {
		t.Fatalf("UpdateLinks error = %v, want physical identity conflict", err)
	}
	content, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "[[oldNote]]" {
		t.Fatalf("preflight did not prevent partial mutation: %q", content)
	}
}

func TestLinkRewriter_GenerateReplacements(t *testing.T) {
	t.Run("Can disable basename wikilink replacement", func(t *testing.T) {
		rewriter := obsidian.LinkRewriter{}

		replacements := rewriter.GenerateReplacements("folder/oldNote", "folder/newNote", obsidian.LinkRewriteOptions{
			IncludeBaseLinks: false,
		})

		assert.NotContains(t, replacements, "[[oldNote]]")
		assert.Equal(t, "[[folder/newNote]]", replacements["[[folder/oldNote]]"])
		assert.Equal(t, "](folder/newNote.md)", replacements["](folder/oldNote.md)"])
	})

	t.Run("Can include basename wikilink replacement", func(t *testing.T) {
		rewriter := obsidian.LinkRewriter{}

		replacements := rewriter.GenerateReplacements("folder/oldNote", "folder/newNote", obsidian.LinkRewriteOptions{
			IncludeBaseLinks: true,
		})

		assert.Equal(t, "[[newNote]]", replacements["[[oldNote]]"])
		assert.Equal(t, "[[folder/newNote]]", replacements["[[folder/oldNote]]"])
	})
}

func TestLinkRewriter_UpdateLinks(t *testing.T) {
	t.Run("Updates links without going through Note", func(t *testing.T) {
		tmpDir := t.TempDir()
		notePath := filepath.Join(tmpDir, "links.md")
		assert.NoError(t, os.WriteFile(notePath, []byte("See [[oldNote]] and [md](oldNote.md)"), 0644))

		rewriter := obsidian.LinkRewriter{}
		err := rewriter.UpdateLinks(tmpDir, "oldNote", "newNote")

		assert.NoError(t, err)
		content, readErr := os.ReadFile(notePath)
		assert.NoError(t, readErr)
		assert.Equal(t, "See [[newNote]] and [md](newNote.md)", string(content))
	})
}

func TestRewriteStructuredLinksPreservesContextAndRelativeTargets(t *testing.T) {
	content := []byte(`---
example: "[[Folder/Old|frontmatter]]"
---
[[Folder/Old#Heading|Alias]]
[relative](../Folder/Old.md#part)
[encoded](../Folder/Old%2Emd)
` + "`[[Folder/Old]]`" + `
<!-- [[Folder/Old]] -->
~~~md
[[Folder/Old]]
~~~
plain Folder/Old
`)
	rewritten, edits, err := (&obsidian.LinkRewriter{}).RewriteStructuredLinks(
		content,
		"Refs/links.md",
		"Folder/Old.md",
		"Archive/New.md",
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	result := string(rewritten)
	assert.Contains(t, result, "[[Archive/New#Heading|Alias]]")
	assert.Contains(t, result, "[relative](../Archive/New.md#part)")
	assert.Contains(t, result, "[encoded](../Archive/New.md)")
	assert.Contains(t, result, `example: "[[Folder/Old|frontmatter]]"`)
	assert.Contains(t, result, "`[[Folder/Old]]`")
	assert.Contains(t, result, "<!-- [[Folder/Old]] -->")
	assert.Contains(t, result, "~~~md\n[[Folder/Old]]\n~~~")
	assert.Contains(t, result, "plain Folder/Old")
	assert.Len(t, edits, 3)
}

func TestRewriteStructuredLinksBasenameAmbiguityPolicy(t *testing.T) {
	rewriter := &obsidian.LinkRewriter{}
	content := []byte("[[Old|Alias]] [[Folder/Old#H]]")
	withoutBase, _, err := rewriter.RewriteStructuredLinks(content, "links.md", "Folder/Old", "Archive/New", false)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "[[Old|Alias]] [[Archive/New#H]]", string(withoutBase))
	withBase, _, err := rewriter.RewriteStructuredLinks(content, "links.md", "Folder/Old", "Archive/New", true)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "[[New|Alias]] [[Archive/New#H]]", string(withBase))
}
