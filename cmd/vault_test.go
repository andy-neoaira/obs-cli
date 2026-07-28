package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

func TestVaultCommandLifecycleJSON(t *testing.T) {
	registry := obsidian.NewRegistry(config.NewStore(filepath.Join(t.TempDir(), "config.json")))
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

func TestVaultDiscoverJSON(t *testing.T) {
	factory := func() (vaultRegistry, error) {
		return obsidian.NewRegistry(config.NewStore(filepath.Join(t.TempDir(), "config.json"))), nil
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

func TestVaultFailureEnvelopeAndDiagnosticRequestID(t *testing.T) {
	registry := obsidian.NewRegistry(config.NewStore(filepath.Join(t.TempDir(), "config.json")))
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

func TestVaultArgumentAndRequestIDErrorsAreJSON(t *testing.T) {
	factory := func() (vaultRegistry, error) {
		return obsidian.NewRegistry(config.NewStore(filepath.Join(t.TempDir(), "config.json"))), nil
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

func TestVaultSameDomainErrorUsesSameCodeAcrossCommands(t *testing.T) {
	registry := obsidian.NewRegistry(config.NewStore(filepath.Join(t.TempDir(), "config.json")))
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

func TestVaultAddDryRunDoesNotWriteConfigOrVault(t *testing.T) {
	root := t.TempDir()
	vaultPath := filepath.Join(root, "Notes")
	if err := os.Mkdir(vaultPath, 0o755); err != nil {
		t.Fatal(err)
	}
	notePath := filepath.Join(vaultPath, "existing.md")
	if err := os.WriteFile(notePath, []byte("# Existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "config", "config.json")
	registry := obsidian.NewRegistry(config.NewStore(configPath))
	factory := func() (vaultRegistry, error) { return registry, nil }
	discover := func() ([]obsidian.DiscoveredVault, error) { return nil, nil }
	before := directoryDigest(t, vaultPath)

	response := executeVaultCommand(t, factory, discover, "add", vaultPath, "--name", "Notes", "--dry-run")
	if !response.OK {
		t.Fatalf("unexpected response: %#v", response)
	}
	var data protocol.DryRunData
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatal(err)
	}
	if !data.DryRun || data.Applied || !data.Changed || len(data.Plan.Changes) != 1 {
		t.Fatalf("unexpected dry-run plan: %#v", data)
	}
	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry-run created config: %v", err)
	}
	if after := directoryDigest(t, vaultPath); after != before {
		t.Fatalf("vault digest changed: before=%s after=%s", before, after)
	}
}

func TestVaultRegistryMutationsDryRunDoNotWrite(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.json")
	registry := obsidian.NewRegistry(config.NewStore(configPath))
	vault, err := registry.Add(t.TempDir(), "Notes")
	if err != nil {
		t.Fatal(err)
	}
	factory := func() (vaultRegistry, error) { return registry, nil }
	discover := func() ([]obsidian.DiscoveredVault, error) { return nil, nil }
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{
		{"set-default", vault.ID, "--dry-run"},
		{"remove", vault.ID, "--dry-run"},
	} {
		response := executeVaultCommand(t, factory, discover, args...)
		if !response.OK {
			t.Fatalf("%v response: %#v", args, response)
		}
		after, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(after, before) {
			t.Fatalf("%v changed registry config", args)
		}
	}
}

func TestCommonFlagsRegistration(t *testing.T) {
	command := &cobra.Command{Use: "test"}
	var values commonFlags
	bindCommonFlags(command, &values, commonFlagSet{
		Output: true, RequestID: true, DryRun: true, IfMatch: true, Vault: true,
	}, false)
	command.SetArgs([]string{
		"--output", "json",
		"--request-id", "req-common",
		"--dry-run",
		"--if-match", "sha256:abc",
		"--vault", "Personal",
	})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if values.Output != "json" || values.RequestID != "req-common" || !values.DryRun ||
		values.IfMatch != "sha256:abc" || values.Vault != "Personal" {
		t.Fatalf("flags not bound: %#v", values)
	}
}

func directoryDigest(t *testing.T, root string) string {
	t.Helper()
	var entries []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			entries = append(entries, "d:"+relative)
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(content)
		entries = append(entries, fmt.Sprintf("f:%s:%x", relative, sum))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(entries)
	sum := sha256.Sum256([]byte(fmt.Sprint(entries)))
	return fmt.Sprintf("%x", sum)
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
	command := newVaultCommand(factory, discover)
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
	if response.ProtocolVersion != "obs-cli/v1" {
		t.Fatalf("protocol version = %q", response.ProtocolVersion)
	}
	return response, diagnostic.String(), executeErr
}
