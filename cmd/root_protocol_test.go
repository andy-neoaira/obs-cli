package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

func TestRootReturnsStableExitCodeForRenderedV2Failure(t *testing.T) {
	registry := obsidian.NewRegistry(config.NewStore(filepath.Join(t.TempDir(), "config-v2.json")))
	factory := func() (vaultRegistry, error) { return registry, nil }
	discover := func() ([]obsidian.DiscoveredVault, error) { return nil, nil }

	root := &cobra.Command{Use: "obs-cli", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(newVaultV2Command(factory, discover))
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
