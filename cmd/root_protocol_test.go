package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"sort"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

func TestRootReturnsStableExitCodeForRenderedFailure(t *testing.T) {
	registry := obsidian.NewRegistry(config.NewStore(filepath.Join(t.TempDir(), "config.json")))
	factory := func() (vaultRegistry, error) { return registry, nil }
	discover := func() ([]obsidian.DiscoveredVault, error) { return nil, nil }

	root := &cobra.Command{Use: "obs-cli", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(newVaultCommand(factory, discover))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"vault", "get", "missing", "--request-id", "req-root"})

	if exit := executeRoot(root); exit != 3 {
		t.Fatalf("exit code = %d, want 3", exit)
	}
	var response struct {
		OK        bool                  `json:"ok"`
		RequestID string                `json:"request_id"`
		Error     *protocol.DomainError `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}
	if response.OK || response.Error == nil || response.Error.Code != protocol.VaultNotFound {
		t.Fatalf("unexpected response: %#v", response)
	}
	if response.RequestID != "req-root" || !bytes.Contains(stderr.Bytes(), []byte("req-root")) {
		t.Fatalf("request ID mismatch stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

func TestRootCommandTreeHasNoLegacyHumanFirstCommands(t *testing.T) {
	root := newRootCommand()
	names := make([]string, 0)
	for _, command := range root.Commands() {
		names = append(names, command.Name())
	}
	sort.Strings(names)
	want := []string{
		"batch", "capabilities", "daily", "doctor", "link", "metadata",
		"note", "search", "template", "vault",
	}
	if len(names) != len(want) {
		t.Fatalf("command tree = %v, want %v", names, want)
	}
	for index := range want {
		if names[index] != want[index] {
			t.Fatalf("command tree = %v, want %v", names, want)
		}
	}
	for _, legacy := range []string{
		"add-vault", "create", "delete", "frontmatter", "list", "list-vaults",
		"move", "open", "print", "remove-vault", "search-content", "set-default",
	} {
		if found, _, err := root.Find([]string{legacy}); err == nil && found.Name() == legacy {
			t.Fatalf("legacy command %q remains registered", legacy)
		}
	}
}

func TestReservedNamespaceReturnsCapabilityUnsupportedJSON(t *testing.T) {
	command := newReservedNamespaceCommand("search")
	var stdout bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&bytes.Buffer{})
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs([]string{"--request-id", "req-reserved"})
	err := command.Execute()
	if err == nil {
		t.Fatal("expected reserved namespace failure")
	}
	var response struct {
		OK        bool                  `json:"ok"`
		Operation string                `json:"operation"`
		Error     *protocol.DomainError `json:"error"`
	}
	if decodeErr := json.Unmarshal(stdout.Bytes(), &response); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if response.OK || response.Operation != "search.status" ||
		response.Error == nil || response.Error.Code != protocol.CapabilityUnsupported {
		t.Fatalf("unexpected response: %#v", response)
	}
}
