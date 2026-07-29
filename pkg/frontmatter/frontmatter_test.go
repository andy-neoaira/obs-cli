package frontmatter_test

import (
	"strings"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/frontmatter"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Run("Parse valid frontmatter", func(t *testing.T) {
		content := "---\ntitle: Test\ntags:\n  - a\n  - b\n---\nBody content"
		fm, body, err := frontmatter.Parse(content)
		assert.NoError(t, err)
		assert.Equal(t, "Test", fm["title"])
		assert.Equal(t, "Body content", body)
	})

	t.Run("Parse empty frontmatter", func(t *testing.T) {
		content := "---\n---\nBody content"
		fm, body, err := frontmatter.Parse(content)
		assert.NoError(t, err)
		assert.Empty(t, fm)
		assert.Equal(t, "Body content", body)
	})

	t.Run("No frontmatter returns empty map", func(t *testing.T) {
		content := "Just body content"
		fm, body, err := frontmatter.Parse(content)
		assert.NoError(t, err)
		assert.Empty(t, fm)
		assert.Equal(t, "Just body content", body)
	})

	t.Run("Invalid YAML returns error", func(t *testing.T) {
		content := "---\ninvalid: [unclosed\n---\nBody"
		_, _, err := frontmatter.Parse(content)
		assert.Error(t, err)
	})
}

func TestHasFrontmatter(t *testing.T) {
	t.Run("Has frontmatter", func(t *testing.T) {
		content := "---\ntitle: Test\n---\nBody"
		assert.True(t, frontmatter.HasFrontmatter(content))
	})

	t.Run("No frontmatter", func(t *testing.T) {
		content := "Just body content"
		assert.False(t, frontmatter.HasFrontmatter(content))
	})

	t.Run("Empty content", func(t *testing.T) {
		assert.False(t, frontmatter.HasFrontmatter(""))
	})
}

func TestSetKey(t *testing.T) {
	t.Run("Add key to existing frontmatter", func(t *testing.T) {
		content := "---\ntitle: Test\n---\nBody"
		result, err := frontmatter.SetKey(content, "author", "John")
		assert.NoError(t, err)
		assert.Contains(t, result, "author: John")
		assert.Contains(t, result, "title: Test")
		assert.Contains(t, result, "Body")
	})

	t.Run("Update existing key", func(t *testing.T) {
		content := "---\ntitle: Old\n---\nBody"
		result, err := frontmatter.SetKey(content, "title", "New")
		assert.NoError(t, err)
		assert.Contains(t, result, "title: New")
		assert.NotContains(t, result, "title: Old")
	})

	t.Run("Create frontmatter when none exists", func(t *testing.T) {
		content := "Just body content"
		result, err := frontmatter.SetKey(content, "title", "New")
		assert.NoError(t, err)
		assert.True(t, strings.HasPrefix(result, "---\n"))
		assert.Contains(t, result, "title: New")
		assert.Contains(t, result, "Just body content")
	})

	t.Run("Parse boolean value true", func(t *testing.T) {
		content := "---\n---\nBody"
		result, err := frontmatter.SetKey(content, "draft", "true")
		assert.NoError(t, err)
		assert.Contains(t, result, "draft: true")
	})

	t.Run("Parse boolean value false", func(t *testing.T) {
		content := "---\n---\nBody"
		result, err := frontmatter.SetKey(content, "published", "false")
		assert.NoError(t, err)
		assert.Contains(t, result, "published: false")
	})

	t.Run("Parse array value", func(t *testing.T) {
		content := "---\n---\nBody"
		result, err := frontmatter.SetKey(content, "tags", "[one, two, three]")
		assert.NoError(t, err)
		assert.Contains(t, result, "tags:")
		assert.Contains(t, result, "- one")
		assert.Contains(t, result, "- two")
		assert.Contains(t, result, "- three")
	})

	t.Run("Parse empty array value", func(t *testing.T) {
		content := "---\n---\nBody"
		result, err := frontmatter.SetKey(content, "tags", "[]")
		assert.NoError(t, err)
		assert.Contains(t, result, "tags: []")
	})
}
