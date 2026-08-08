package noteops_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/noteops"
)

func TestSearchPaginationScopeTruncationAndSnippets(t *testing.T) {
	service, root := newService(t)
	writeNoteFixture(t, root, "Projects/Alpha.md", "needle\n"+strings.Repeat("界", 260)+" needle")
	writeNoteFixture(t, root, "Projects/Needle Name.md", "other")
	writeNoteFixture(t, root, "Archive/Old.md", "needle")
	writeNoteFixture(t, root, "Projects/Binary.md", "needle\x00ignored")

	page, err := service.Search("needle", "Projects", 1, 2, 20)
	if err != nil {
		t.Fatal(err)
	}
	if page.TotalResults != 3 || len(page.Results) != 2 || !page.HasMore || page.ScannedFiles != 3 {
		t.Fatalf("unexpected first page: %#v", page)
	}
	foundBoundedSnippet := false
	for _, result := range page.Results {
		if len([]rune(result.Snippet)) == 241 && strings.HasSuffix(result.Snippet, "…") {
			foundBoundedSnippet = true
		}
	}
	second, err := service.Search("needle", "Projects", 2, 2, 20)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range second.Results {
		if len([]rune(result.Snippet)) == 241 && strings.HasSuffix(result.Snippet, "…") {
			foundBoundedSnippet = true
		}
	}
	if !foundBoundedSnippet || len(second.Results) != 1 || second.HasMore {
		t.Fatalf("unexpected bounded/second-page result: %#v", second)
	}

	truncated, err := service.Search("needle", "", 1, 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !truncated.Truncated || truncated.ScannedFiles != 1 {
		t.Fatalf("search was not bounded: %#v", truncated)
	}
	if _, err := service.Search("  ", "", 1, 10, 10); err == nil {
		t.Fatal("empty query should fail")
	}
	if _, err := service.Search("needle", "Missing", 1, 10, 10); !errors.Is(err, noteops.ErrNoteNotFound) {
		t.Fatalf("missing scope error = %v", err)
	}
}

func TestBacklinksRecognizeSupportedLinksAndBoundScanning(t *testing.T) {
	service, root := newService(t)
	writeNoteFixture(t, root, "Projects/Target Note.md", "target")
	writeNoteFixture(t, root, "Projects/Wiki.md", "[[Target Note#Heading|alias]]\n[[Projects/Target Note]]")
	writeNoteFixture(t, root, "Refs/Markdown.md", "[target](../Projects/Target%20Note.md#part)\n[web](https://example.com/Target%20Note.md)")
	writeNoteFixture(t, root, "Refs/Other.md", "[[Other]]")

	report, err := service.Backlinks("Projects/Target Note", "", 20)
	if err != nil {
		t.Fatal(err)
	}
	if !report.TargetExists || report.Target != "Projects/Target Note.md" || len(report.Results) != 3 {
		t.Fatalf("unexpected backlinks: %#v", report)
	}
	kinds := map[string]int{}
	for _, result := range report.Results {
		kinds[result.Kind]++
		if result.Revision == "" || result.Line < 1 {
			t.Fatalf("backlink lacks evidence: %#v", result)
		}
	}
	if kinds["wikilink"] != 2 || kinds["markdown"] != 1 {
		t.Fatalf("backlink kinds = %#v", kinds)
	}

	missing, err := service.Backlinks("Projects/Future", "Refs", 1)
	if err != nil {
		t.Fatal(err)
	}
	if missing.TargetExists || !missing.Truncated || missing.ScannedFiles != 1 {
		t.Fatalf("unexpected missing/truncated report: %#v", missing)
	}
}

func writeNoteFixture(t *testing.T, root, logical, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(logical))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
