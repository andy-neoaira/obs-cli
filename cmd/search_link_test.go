package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/noteops"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

func TestSearchReturnsBoundedRevisionEvidence(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	files := map[string]string{
		"Notes/One.md": "---\nbroken: [\n---\nneedle first\n",
		"Notes/Two.md": "needle second\nneedle third\n",
		"Notes/No.md":  "unrelated\n",
	}
	for name, content := range files {
		full := filepath.Join(vaultRoot, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(vaultRoot, "image.png"), []byte("needle binary"), 0o644); err != nil {
		t.Fatal(err)
	}

	command := func() *cobra.Command { return newSearchCommand(registryFactory, serviceFactory) }
	first := executeTestCommand(t, command(), "", "content", "needle",
		"--scope", "Notes", "--page", "1", "--page-size", "2", "--max-files", "10")
	var data struct {
		Search noteops.SearchPage `json:"search"`
	}
	if err := json.Unmarshal(first.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.Search.TotalResults != 3 || len(data.Search.Results) != 2 || !data.Search.HasMore {
		t.Fatalf("search page = %#v", data.Search)
	}
	for _, result := range data.Search.Results {
		if result.Path == "" || result.Revision == "" || result.Line < 1 || result.Snippet == "" {
			t.Fatalf("missing search evidence: %#v", result)
		}
	}
	empty := executeTestCommand(t, command(), "", "content", "absent", "--scope", "Notes")
	if err := json.Unmarshal(empty.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.Search.TotalResults != 0 || len(data.Search.Results) != 0 {
		t.Fatalf("empty search is not distinguishable: %#v", data.Search)
	}
	truncated := executeTestCommand(t, command(), "", "content", "needle", "--max-files", "1")
	if err := json.Unmarshal(truncated.Data, &data); err != nil {
		t.Fatal(err)
	}
	if !data.Search.Truncated || data.Search.ScannedFiles != 1 {
		t.Fatalf("search limit not enforced: %#v", data.Search)
	}
	invalid, _, err := executeNoteCommandResult(
		t, registryFactory, serviceFactory, "", "get", "Notes/One",
	)
	if err == nil || invalid.Error == nil || invalid.Error.Code != protocol.InvalidFrontmatter ||
		invalid.Error.Details["path"] != "Notes/One.md" || invalid.Error.Details["revision"] == "" {
		t.Fatalf("invalid frontmatter evidence = %#v err=%v", invalid, err)
	}
}

func TestSearchRejectsUnboundedPage(t *testing.T) {
	registryFactory, serviceFactory, _ := noteCommandDependencies(t)
	response, _, err := executeTestCommandResult(
		newSearchCommand(registryFactory, serviceFactory), "",
		"content", "anything", "--page-size", "101",
	)
	if err == nil || response.Error == nil || response.Error.Code != protocol.InvalidArgument {
		t.Fatalf("unbounded search = %#v err=%v", response, err)
	}
}

func TestLinkBacklinksReturnsSourceRevision(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	files := map[string]string{
		"Knowledge/Target.md": "# Target\n",
		"Sources/Wiki.md":     "See [[Knowledge/Target#Heading|alias]].\n",
		"Sources/Markdown.md": "See [target](../Knowledge/Target.md).\n",
		"Sources/Other.md":    "See [[Elsewhere]].\n",
		"Outside.md":          "See [[Knowledge/Target]].\n",
	}
	for name, content := range files {
		full := filepath.Join(vaultRoot, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	response := executeTestCommand(t,
		newLinkCommand(registryFactory, serviceFactory), "",
		"backlinks", "Knowledge/Target", "--scope", "Sources", "--max-files", "20",
	)
	var data struct {
		Backlinks noteops.BacklinkReport `json:"backlinks"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatal(err)
	}
	if !data.Backlinks.TargetExists || len(data.Backlinks.Results) != 2 {
		t.Fatalf("backlinks = %#v", data.Backlinks)
	}
	for _, result := range data.Backlinks.Results {
		if result.Path == "" || result.Revision == "" || result.Line != 1 {
			t.Fatalf("missing backlink evidence: %#v", result)
		}
	}
	if err := os.Remove(filepath.Join(vaultRoot, "Knowledge", "Target.md")); err != nil {
		t.Fatal(err)
	}
	missing := executeTestCommand(t,
		newLinkCommand(registryFactory, serviceFactory), "",
		"backlinks", "Knowledge/Target", "--scope", "Sources", "--max-files", "20",
	)
	if err := json.Unmarshal(missing.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.Backlinks.TargetExists || len(data.Backlinks.Results) != 2 {
		t.Fatalf("missing-target backlinks = %#v", data.Backlinks)
	}
}
