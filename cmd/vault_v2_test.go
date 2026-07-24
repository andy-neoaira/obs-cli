package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
)

func TestVaultV2CommandLifecycleJSON(t *testing.T) {
	registry := obsidian.NewRegistry(config.NewStore(filepath.Join(t.TempDir(), "config-v2.json")))
	factory := func() (vaultRegistry, error) { return registry, nil }
	discover := func() ([]obsidian.DiscoveredVault, error) { return nil, nil }
	vaultPath := t.TempDir()

	add := executeVaultCommand(t, factory, discover, "add", vaultPath, "--name", "Personal", "--set-default", "--request-id", "test-add")
	if add.Operation != "vault.add" || !add.OK || add.RequestID != "test-add" {
		t.Fatalf("unexpected add response: %#v", add)
	}

	list := executeVaultCommand(t, factory, discover, "list", "--request-id", "test-list")
	if list.Operation != "vault.list" || !list.OK {
		t.Fatalf("unexpected list response: %#v", list)
	}

	get := executeVaultCommand(t, factory, discover, "get", "Personal")
	if get.Operation != "vault.get" || !get.OK {
		t.Fatalf("unexpected get response: %#v", get)
	}

	selected := executeVaultCommand(t, factory, discover, "set-default", "Personal")
	if selected.Operation != "vault.set-default" || !selected.OK {
		t.Fatalf("unexpected set-default response: %#v", selected)
	}

	removed := executeVaultCommand(t, factory, discover, "remove", "Personal")
	if removed.Operation != "vault.remove" || !removed.OK {
		t.Fatalf("unexpected remove response: %#v", removed)
	}
}

func TestVaultV2DiscoverJSON(t *testing.T) {
	factory := func() (vaultRegistry, error) {
		return obsidian.NewRegistry(config.NewStore(filepath.Join(t.TempDir(), "config-v2.json"))), nil
	}
	discover := func() ([]obsidian.DiscoveredVault, error) {
		return []obsidian.DiscoveredVault{
			{SourceID: "official-id", Name: "Notes", Path: "/vault/Notes", Open: true, Available: true},
		}, nil
	}

	response := executeVaultCommand(t, factory, discover, "discover", "--output", "json")
	if response.Operation != "vault.discover" || !response.OK {
		t.Fatalf("unexpected response: %#v", response)
	}
}

type vaultTestEnvelope struct {
	ProtocolVersion string          `json:"protocol_version"`
	OK              bool            `json:"ok"`
	Operation       string          `json:"operation"`
	RequestID       string          `json:"request_id"`
	Data            json.RawMessage `json:"data"`
}

func executeVaultCommand(
	t *testing.T,
	factory vaultRegistryFactory,
	discover vaultDiscoverFunc,
	args ...string,
) vaultTestEnvelope {
	t.Helper()
	command := newVaultV2Command(factory, discover)
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetErr(&output)
	command.SilenceUsage = true
	command.SilenceErrors = true
	command.SetArgs(args)
	if err := command.Execute(); err != nil {
		t.Fatalf("execute vault %v: %v", args, err)
	}

	var response vaultTestEnvelope
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("decode response %q: %v", output.String(), err)
	}
	if response.ProtocolVersion != "obs-cli/v2" {
		t.Fatalf("protocol version = %q", response.ProtocolVersion)
	}
	return response
}
