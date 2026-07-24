package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/noteops"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

func noteCommandDependencies(t *testing.T) (vaultRegistryFactory, noteServiceFactory, string) {
	t.Helper()
	root := t.TempDir()
	vaultRoot := filepath.Join(root, "vault")
	if err := os.Mkdir(vaultRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	registry := obsidian.NewRegistry(config.NewStore(filepath.Join(root, "config-v2.json")))
	vault, err := registry.Add(vaultRoot, "Notes")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.SetDefault(vault.ID); err != nil {
		t.Fatal(err)
	}
	registryFactory := func() (vaultRegistry, error) { return registry, nil }
	lockRoot := filepath.Join(root, "runtime", "locks")
	serviceFactory := func(vaultPath string) (noteService, error) {
		return noteops.NewService(vaultPath, storage.NewStore(lockRoot))
	}
	return registryFactory, serviceFactory, vaultRoot
}

func TestNoteV2ReadAnalyzeConditionalUpdateLifecycle(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)

	created := executeNoteCommand(
		t, registryFactory, serviceFactory, "# Demo\nold\n",
		"create", "Projects/demo", "--content-file", "-", "--request-id", "req-create",
	)
	if !created.OK || created.Operation != "note.create" || created.RequestID != "req-create" {
		t.Fatalf("create response = %#v", created)
	}

	got := executeNoteCommand(t, registryFactory, serviceFactory, "", "get", "Projects/demo")
	var getData struct {
		Note noteops.Note `json:"note"`
	}
	if err := json.Unmarshal(got.Data, &getData); err != nil {
		t.Fatal(err)
	}
	if getData.Note.Content != "# Demo\nold\n" || getData.Note.Revision == "" {
		t.Fatalf("get data = %#v", getData)
	}

	matchFile := filepath.Join(t.TempDir(), "match.md")
	if err := os.WriteFile(matchFile, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	patched := executeNoteCommand(
		t, registryFactory, serviceFactory, "new",
		"patch", "Projects/demo",
		"--match-file", matchFile,
		"--content-file", "-",
		"--if-match", getData.Note.Revision,
	)
	if !patched.OK {
		t.Fatalf("patch response = %#v", patched)
	}
	content, err := os.ReadFile(filepath.Join(vaultRoot, "Projects", "demo.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "# Demo\nnew\n" {
		t.Fatalf("patched content = %q", content)
	}
}

func TestNoteV2DryRunAndInputSafety(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "base", "create", "demo", "--content-file", "-")
	repeated, _, repeatedErr := executeNoteCommandResult(
		t, registryFactory, serviceFactory, "base",
		"create", "demo", "--content-file", "-",
	)
	if repeatedErr == nil || repeated.Error == nil || repeated.Error.Code != protocol.AlreadyExists ||
		repeated.Error.Details["same_content"] != true {
		t.Fatalf("repeated create response=%#v err=%v", repeated, repeatedErr)
	}
	path := filepath.Join(vaultRoot, "demo.md")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	dry := executeNoteCommand(
		t, registryFactory, serviceFactory, "addition",
		"append", "demo", "--content-file", "-", "--dry-run",
	)
	var data struct {
		DryRun  bool          `json:"dry_run"`
		Applied bool          `json:"applied"`
		Plan    protocol.Plan `json:"plan"`
	}
	if err := json.Unmarshal(dry.Data, &data); err != nil {
		t.Fatal(err)
	}
	if !data.DryRun || data.Applied || len(data.Plan.Changes) != 1 {
		t.Fatalf("dry-run data = %#v", data)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("dry-run changed note: before=%q after=%q", before, after)
	}

	response, _, err := executeNoteCommandResult(
		t, registryFactory, serviceFactory, "unused",
		"patch", "demo", "--match-file", "-", "--content-file", "-", "--if-match", storage.Revision(before),
	)
	if err == nil || response.Error == nil || response.Error.Code != protocol.InvalidArgument {
		t.Fatalf("dual stdin response=%#v err=%v", response, err)
	}
}

func TestNoteV2PatchConflictAndDangerousReplaceGuard(t *testing.T) {
	registryFactory, serviceFactory, _ := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "same\nsame\n", "create", "demo", "--content-file", "-")
	current := executeNoteCommand(t, registryFactory, serviceFactory, "", "get", "demo")
	var getData struct {
		Note noteops.Note `json:"note"`
	}
	if err := json.Unmarshal(current.Data, &getData); err != nil {
		t.Fatal(err)
	}
	matchFile := filepath.Join(t.TempDir(), "match.md")
	if err := os.WriteFile(matchFile, []byte("same"), 0o600); err != nil {
		t.Fatal(err)
	}
	ambiguous, _, err := executeNoteCommandResult(
		t, registryFactory, serviceFactory, "new",
		"patch", "demo", "--match-file", matchFile, "--content-file", "-", "--if-match", getData.Note.Revision,
	)
	if err == nil || ambiguous.Error == nil || ambiguous.Error.Code != protocol.AmbiguousNote {
		t.Fatalf("ambiguous response=%#v err=%v", ambiguous, err)
	}

	guarded, _, err := executeNoteCommandResult(
		t, registryFactory, serviceFactory, "replace",
		"replace", "demo", "--content-file", "-",
	)
	if err == nil || guarded.Error == nil || guarded.Error.Code != protocol.InvalidArgument {
		t.Fatalf("replace guard response=%#v err=%v", guarded, err)
	}
}

func TestNoteV2DeleteRequiresMatchingRevision(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "content", "create", "demo", "--content-file", "-")
	stale, _, err := executeNoteCommandResult(
		t, registryFactory, serviceFactory, "",
		"delete", "demo", "--if-match", storage.Revision([]byte("stale")),
	)
	if err == nil || stale.Error == nil || stale.Error.Code != protocol.RevisionConflict {
		t.Fatalf("stale delete response=%#v err=%v", stale, err)
	}
	current := executeNoteCommand(t, registryFactory, serviceFactory, "", "get", "demo")
	var getData struct {
		Note noteops.Note `json:"note"`
	}
	if err := json.Unmarshal(current.Data, &getData); err != nil {
		t.Fatal(err)
	}
	deleted := executeNoteCommand(
		t, registryFactory, serviceFactory, "",
		"delete", "demo", "--if-match", getData.Note.Revision,
	)
	if !deleted.OK {
		t.Fatalf("delete response = %#v", deleted)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "demo.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("note still exists: %v", err)
	}
}

func TestNoteV2MoveDryRunAndApply(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	executeNoteCommand(t, registryFactory, serviceFactory, "source", "create", "Old", "--content-file", "-")
	executeNoteCommand(t, registryFactory, serviceFactory, "[[Old|Alias]]", "create", "links", "--content-file", "-")
	current := executeNoteCommand(t, registryFactory, serviceFactory, "", "get", "Old")
	var getData struct {
		Note noteops.Note `json:"note"`
	}
	if err := json.Unmarshal(current.Data, &getData); err != nil {
		t.Fatal(err)
	}
	dry := executeNoteCommand(
		t, registryFactory, serviceFactory, "",
		"move", "Old", "Archive/New", "--if-match", getData.Note.Revision, "--dry-run",
	)
	var dryData struct {
		DryRun bool          `json:"dry_run"`
		Plan   protocol.Plan `json:"plan"`
	}
	if err := json.Unmarshal(dry.Data, &dryData); err != nil {
		t.Fatal(err)
	}
	if !dryData.DryRun || len(dryData.Plan.Changes) != 2 {
		t.Fatalf("move dry-run = %#v", dryData)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "Archive", "New.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry-run created target: %v", err)
	}
	applied := executeNoteCommand(
		t, registryFactory, serviceFactory, "",
		"move", "Old", "Archive/New", "--if-match", getData.Note.Revision,
	)
	if !applied.OK {
		t.Fatalf("move response = %#v", applied)
	}
	assertFileContent(t, filepath.Join(vaultRoot, "Archive", "New.md"), "source")
	assertFileContent(t, filepath.Join(vaultRoot, "links.md"), "[[New|Alias]]")
}

func TestV2RCSmoke(t *testing.T) {
	registryFactory, serviceFactory, vaultRoot := noteCommandDependencies(t)
	created := executeNoteCommand(
		t, registryFactory, serviceFactory, "# RC\n",
		"create", "RC/demo", "--content-file", "-", "--request-id", "rc-create",
	)
	if !created.OK {
		t.Fatalf("create failed: %#v", created)
	}
	duplicate, _, err := executeNoteCommandResult(
		t, registryFactory, serviceFactory, "# RC\n",
		"create", "RC/demo", "--content-file", "-",
	)
	if err == nil || duplicate.Error == nil || duplicate.Error.Code != protocol.AlreadyExists {
		t.Fatalf("duplicate create=%#v err=%v", duplicate, err)
	}

	current := executeNoteCommand(t, registryFactory, serviceFactory, "", "get", "RC/demo")
	var data struct {
		Note noteops.Note `json:"note"`
	}
	if err := json.Unmarshal(current.Data, &data); err != nil {
		t.Fatal(err)
	}
	dry := executeNoteCommand(
		t, registryFactory, serviceFactory, "item",
		"append", "RC/demo", "--content-file", "-", "--if-match", data.Note.Revision, "--dry-run",
	)
	if !dry.OK {
		t.Fatalf("append dry-run failed: %#v", dry)
	}
	appended := executeNoteCommand(
		t, registryFactory, serviceFactory, "item",
		"append", "RC/demo", "--content-file", "-", "--if-match", data.Note.Revision,
	)
	if !appended.OK {
		t.Fatalf("append failed: %#v", appended)
	}
	conflict, _, err := executeNoteCommandResult(
		t, registryFactory, serviceFactory, "stale",
		"replace", "RC/demo", "--content-file", "-", "--if-match", data.Note.Revision,
	)
	if err == nil || conflict.Error == nil || conflict.Error.Code != protocol.RevisionConflict {
		t.Fatalf("replace conflict=%#v err=%v", conflict, err)
	}

	current = executeNoteCommand(t, registryFactory, serviceFactory, "", "get", "RC/demo")
	if err := json.Unmarshal(current.Data, &data); err != nil {
		t.Fatal(err)
	}
	replaced := executeNoteCommand(
		t, registryFactory, serviceFactory, "final",
		"replace", "RC/demo", "--content-file", "-", "--if-match", data.Note.Revision,
	)
	if !replaced.OK {
		t.Fatalf("replace failed: %#v", replaced)
	}
	current = executeNoteCommand(t, registryFactory, serviceFactory, "", "get", "RC/demo")
	if err := json.Unmarshal(current.Data, &data); err != nil {
		t.Fatal(err)
	}
	deleted := executeNoteCommand(
		t, registryFactory, serviceFactory, "",
		"delete", "RC/demo", "--if-match", data.Note.Revision,
	)
	if !deleted.OK {
		t.Fatalf("delete failed: %#v", deleted)
	}
	if _, err := os.Stat(filepath.Join(vaultRoot, "RC", "demo.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("delete verification failed: %v", err)
	}
}

type noteTestEnvelope struct {
	OK        bool                  `json:"ok"`
	Operation string                `json:"operation"`
	RequestID string                `json:"request_id"`
	Data      json.RawMessage       `json:"data"`
	Error     *protocol.DomainError `json:"error"`
}

func executeNoteCommand(
	t *testing.T,
	registryFactory vaultRegistryFactory,
	serviceFactory noteServiceFactory,
	stdin string,
	args ...string,
) noteTestEnvelope {
	t.Helper()
	response, _, err := executeNoteCommandResult(t, registryFactory, serviceFactory, stdin, args...)
	if err != nil {
		t.Fatalf("execute note %v: %v", args, err)
	}
	return response
}

func executeNoteCommandResult(
	t *testing.T,
	registryFactory vaultRegistryFactory,
	serviceFactory noteServiceFactory,
	stdin string,
	args ...string,
) (noteTestEnvelope, string, error) {
	t.Helper()
	command := newNoteV2Command(registryFactory, serviceFactory)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetIn(strings.NewReader(stdin))
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(args)
	executeErr := command.Execute()

	var response noteTestEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}
	return response, stderr.String(), executeErr
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want {
		t.Fatalf("content = %q, want %q", data, want)
	}
}
