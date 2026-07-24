package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
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

func TestVaultV2FailureEnvelopeAndDiagnosticRequestID(t *testing.T) {
	registry := obsidian.NewRegistry(config.NewStore(filepath.Join(t.TempDir(), "config-v2.json")))
	factory := func() (vaultRegistry, error) { return registry, nil }
	discover := func() ([]obsidian.DiscoveredVault, error) { return nil, nil }

	response, diagnostic, err := executeVaultCommandResult(
		t,
		factory,
		discover,
		"get",
		"missing",
		"--request-id",
		"req-failure",
	)
	if err == nil || !errors.Is(err, obsidian.ErrVaultNotFound) {
		t.Fatalf("execute error = %v", err)
	}
	if response.OK || response.Error == nil || response.Error.Code != protocol.VaultNotFound {
		t.Fatalf("unexpected failure envelope: %#v", response)
	}
	if response.RequestID != "req-failure" || !bytes.Contains([]byte(diagnostic), []byte("[req-failure]")) {
		t.Fatalf("request ID mismatch response=%q diagnostic=%q", response.RequestID, diagnostic)
	}
}

func TestVaultV2ArgumentAndRequestIDErrorsAreJSON(t *testing.T) {
	factory := func() (vaultRegistry, error) {
		return obsidian.NewRegistry(config.NewStore(filepath.Join(t.TempDir(), "config-v2.json"))), nil
	}
	discover := func() ([]obsidian.DiscoveredVault, error) { return nil, nil }

	missingArg, _, err := executeVaultCommandResult(t, factory, discover, "get")
	if err == nil || missingArg.Error == nil || missingArg.Error.Code != protocol.InvalidArgument {
		t.Fatalf("missing arg response=%#v error=%v", missingArg, err)
	}
	invalidID, _, err := executeVaultCommandResult(t, factory, discover, "list", "--request-id", "bad id")
	if err == nil || invalidID.Error == nil || invalidID.Error.Code != protocol.InvalidArgument {
		t.Fatalf("invalid ID response=%#v error=%v", invalidID, err)
	}
	if invalidID.RequestID == "bad id" || invalidID.RequestID == "" {
		t.Fatalf("invalid request ID leaked into envelope: %q", invalidID.RequestID)
	}
}

func TestVaultV2SameDomainErrorUsesSameCodeAcrossCommands(t *testing.T) {
	registry := obsidian.NewRegistry(config.NewStore(filepath.Join(t.TempDir(), "config-v2.json")))
	factory := func() (vaultRegistry, error) { return registry, nil }
	discover := func() ([]obsidian.DiscoveredVault, error) { return nil, nil }

	get, _, getErr := executeVaultCommandResult(t, factory, discover, "get", "missing")
	remove, _, removeErr := executeVaultCommandResult(t, factory, discover, "remove", "missing")
	if getErr == nil || removeErr == nil {
		t.Fatalf("expected both commands to fail: get=%v remove=%v", getErr, removeErr)
	}
	if get.Error == nil || remove.Error == nil || get.Error.Code != remove.Error.Code {
		t.Fatalf("inconsistent codes: get=%#v remove=%#v", get.Error, remove.Error)
	}
	if get.Error.Code != protocol.VaultNotFound {
		t.Fatalf("code = %s, want VAULT_NOT_FOUND", get.Error.Code)
	}
}

type vaultTestEnvelope struct {
	ProtocolVersion string                `json:"protocol_version"`
	OK              bool                  `json:"ok"`
	Operation       string                `json:"operation"`
	RequestID       string                `json:"request_id"`
	Data            json.RawMessage       `json:"data"`
	Error           *protocol.DomainError `json:"error"`
}

func executeVaultCommand(
	t *testing.T,
	factory vaultRegistryFactory,
	discover vaultDiscoverFunc,
	args ...string,
) vaultTestEnvelope {
	t.Helper()
	response, _, err := executeVaultCommandResult(t, factory, discover, args...)
	if err != nil {
		t.Fatalf("execute vault %v: %v", args, err)
	}
	return response
}

func executeVaultCommandResult(
	t *testing.T,
	factory vaultRegistryFactory,
	discover vaultDiscoverFunc,
	args ...string,
) (vaultTestEnvelope, string, error) {
	t.Helper()
	command := newVaultV2Command(factory, discover)
	var output bytes.Buffer
	var diagnostic bytes.Buffer
	command.SetOut(&output)
	command.SetErr(&diagnostic)
	command.SilenceUsage = true
	command.SilenceErrors = true
	command.SetArgs(args)
	executeErr := command.Execute()

	var response vaultTestEnvelope
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("decode response %q: %v", output.String(), err)
	}
	if response.ProtocolVersion != "obs-cli/v2" {
		t.Fatalf("protocol version = %q", response.ProtocolVersion)
	}
	return response, diagnostic.String(), executeErr
}
