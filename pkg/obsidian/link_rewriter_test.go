package obsidian_test

import (
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/stretchr/testify/assert"
)

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
	withoutBase, _, err := rewriter.RewriteStructuredLinks(
		content, "links.md", "Folder/Old", "Archive/New", false,
	)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "[[Old|Alias]] [[Archive/New#H]]", string(withoutBase))

	withBase, _, err := rewriter.RewriteStructuredLinks(
		content, "links.md", "Folder/Old", "Archive/New", true,
	)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "[[New|Alias]] [[Archive/New#H]]", string(withBase))
}
