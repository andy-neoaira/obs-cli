package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

func TestDailyV2UsesObsidianConfigTemplateAndRevision(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	if err := os.MkdirAll(filepath.Join(vaultRoot, ".obsidian"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(vaultRoot, "Templates"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vaultRoot, ".obsidian", "daily-notes.json"),
		[]byte(`{"folder":"Journal","format":"YYYY/MM/DD","template":"Templates/Daily"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vaultRoot, "Templates", "Daily.md"),
		[]byte("# {{title}}\n\n## Log\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	now := func() time.Time { return time.Date(2026, 7, 25, 9, 30, 0, 0, time.Local) }

	get := executeV2TestCommand(t, newDailyV2Command(registryFactory, serviceFactory, now), "", "get")
	var getData struct {
		Exists bool        `json:"exists"`
		Target dailyTarget `json:"target"`
	}
	if err := json.Unmarshal(get.Data, &getData); err != nil {
		t.Fatal(err)
	}
	if getData.Exists || getData.Target.Path != "Journal/2026/07/25.md" {
		t.Fatalf("daily get = %#v", getData)
	}

	dry := executeV2TestCommand(t, newDailyV2Command(registryFactory, serviceFactory, now), "", "create", "--dry-run")
	if !dry.OK {
		t.Fatalf("daily dry-run = %#v", dry)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "Journal", "2026", "07", "25.md")); !os.IsNotExist(err) {
		t.Fatalf("daily dry-run wrote file: %v", err)
	}
	created := executeV2TestCommand(t, newDailyV2Command(registryFactory, serviceFactory, now), "", "create")
	if !created.OK {
		t.Fatalf("daily create = %#v", created)
	}
	service, err := serviceFactory(vaultRoot)
	if err != nil {
		t.Fatal(err)
	}
	note, err := service.Get("Journal/2026/07/25")
	if err != nil {
		t.Fatal(err)
	}
	if note.Content != "# 2026/07/25\n\n## Log\n" {
		t.Fatalf("rendered daily = %q", note.Content)
	}

	appended := executeV2TestCommand(t, newDailyV2Command(registryFactory, serviceFactory, now), "- shipped",
		"append", "--content-file", "-", "--section", "Log", "--if-match", note.Revision)
	if !appended.OK {
		t.Fatalf("daily append = %#v", appended)
	}
	updated, err := service.Get("Journal/2026/07/25")
	if err != nil || !strings.Contains(updated.Content, "## Log\n- shipped") {
		t.Fatalf("updated daily = %#v err=%v", updated, err)
	}
	if err := os.Remove(filepath.Join(vaultRoot, "Templates", "Daily.md")); err != nil {
		t.Fatal(err)
	}
	withoutTemplate := executeV2TestCommand(t, newDailyV2Command(registryFactory, serviceFactory, now), "", "get")
	if !withoutTemplate.OK {
		t.Fatalf("existing daily depends on deleted template: %#v", withoutTemplate)
	}
	conflict, _, err := executeV2TestCommandResult(
		newDailyV2Command(registryFactory, serviceFactory, now), "again",
		"append", "--content-file", "-", "--if-match", note.Revision,
	)
	if err == nil || conflict.Error == nil || conflict.Error.Code != protocol.RevisionConflict {
		t.Fatalf("stale daily append = %#v err=%v", conflict, err)
	}
}

func TestMetadataV2PreservesUnknownFieldsAndBody(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	service, err := serviceFactory(vaultRoot)
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.Create("Projects/Demo", []byte("---\nowner: andy\ncustom: keep\n---\n# Demo\nBody\n"))
	if err != nil {
		t.Fatal(err)
	}
	command := func() *cobra.Command { return newMetadataV2Command(registryFactory, serviceFactory) }

	dry := executeV2TestCommand(t, command(), "", "set", "Projects/Demo",
		"--key", "status", "--value", "active", "--if-match", created.RevisionAfter, "--dry-run")
	if !dry.OK {
		t.Fatalf("metadata dry-run = %#v", dry)
	}
	before, err := service.Get("Projects/Demo")
	if err != nil || before.Frontmatter["status"] != nil {
		t.Fatalf("dry-run changed metadata: %#v err=%v", before, err)
	}
	applied := executeV2TestCommand(t, command(), "", "set", "Projects/Demo",
		"--key", "status", "--value", "active", "--if-match", before.Revision)
	if !applied.OK {
		t.Fatalf("metadata set = %#v", applied)
	}
	after, err := service.Get("Projects/Demo")
	if err != nil {
		t.Fatal(err)
	}
	if after.Frontmatter["owner"] != "andy" || after.Frontmatter["custom"] != "keep" ||
		after.Frontmatter["status"] != "active" || !strings.HasSuffix(after.Content, "# Demo\nBody\n") ||
		after.BodyRevision != before.BodyRevision {
		t.Fatalf("metadata update damaged note: %#v", after)
	}
	conflict, _, err := executeV2TestCommandResult(command(), "",
		"set", "Projects/Demo", "--key", "status", "--value", "done", "--if-match", before.Revision)
	if err == nil || conflict.Error == nil || conflict.Error.Code != protocol.RevisionConflict {
		t.Fatalf("stale metadata set = %#v err=%v", conflict, err)
	}
}

func TestDailyV2DoesNotOverwriteDailyCreatedByAnotherEntry(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	now := func() time.Time { return time.Date(2026, 7, 25, 12, 0, 0, 0, time.Local) }
	get := executeV2TestCommand(t, newDailyV2Command(registryFactory, serviceFactory, now), "", "get")
	if !get.OK {
		t.Fatalf("initial daily get = %#v", get)
	}
	external := []byte("# Created on mobile\n")
	if err := os.WriteFile(filepath.Join(vaultRoot, "2026-07-25.md"), external, 0o644); err != nil {
		t.Fatal(err)
	}
	conflict, _, err := executeV2TestCommandResult(
		newDailyV2Command(registryFactory, serviceFactory, now), "", "create",
	)
	if err == nil || conflict.Error == nil || conflict.Error.Code != protocol.AlreadyExists {
		t.Fatalf("cross-entry daily create = %#v err=%v", conflict, err)
	}
	data, readErr := os.ReadFile(filepath.Join(vaultRoot, "2026-07-25.md"))
	if readErr != nil || !bytes.Equal(data, external) {
		t.Fatalf("external daily overwritten: %q err=%v", data, readErr)
	}
}

func TestDailyV2RejectsMalformedConfigWithoutWriting(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	configDir := filepath.Join(vaultRoot, ".obsidian")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "daily-notes.json"), []byte("{bad"), 0o644); err != nil {
		t.Fatal(err)
	}
	now := func() time.Time { return time.Date(2026, 7, 28, 12, 0, 0, 0, time.Local) }

	for _, operation := range []string{"get", "create"} {
		response, _, err := executeV2TestCommandResult(
			newDailyV2Command(registryFactory, serviceFactory, now), "", operation,
		)
		if err == nil || response.Error == nil || response.Error.Code != protocol.InvalidArgument {
			t.Fatalf("%s malformed config response=%#v err=%v", operation, response, err)
		}
		if response.Error.Details["config_file"] != ".obsidian/daily-notes.json" ||
			response.Error.Details["failure_kind"] != "parse" {
			t.Fatalf("%s malformed config details=%#v", operation, response.Error.Details)
		}
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "2026-07-28.md")); !os.IsNotExist(err) {
		t.Fatalf("malformed config created default daily: %v", err)
	}
}

func executeV2TestCommand(t *testing.T, command *cobra.Command, stdin string, args ...string) noteTestEnvelope {
	t.Helper()
	response, output, err := executeV2TestCommandResult(command, stdin, args...)
	if err != nil {
		t.Fatalf("execute %v: %v\n%s", args, err, output)
	}
	return response
}

func executeV2TestCommandResult(command *cobra.Command, stdin string, args ...string) (noteTestEnvelope, string, error) {
	var stdout, stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetIn(strings.NewReader(stdin))
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(args)
	err := command.Execute()
	var response noteTestEnvelope
	if decodeErr := json.Unmarshal(stdout.Bytes(), &response); decodeErr != nil {
		return response, stdout.String(), decodeErr
	}
	return response, stderr.String(), err
}
