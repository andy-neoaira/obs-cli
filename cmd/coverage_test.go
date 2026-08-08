package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

func TestDefaultFactoriesAndRootErrorPaths(t *testing.T) {
	t.Setenv("OBS_CLI_CONFIG_HOME", t.TempDir())
	registry, err := defaultVaultRegistryFactory()
	if err != nil || registry == nil {
		t.Fatalf("defaultVaultRegistryFactory = %v, %v", registry, err)
	}
	service, err := defaultNoteServiceFactory(t.TempDir())
	if err != nil || service == nil {
		t.Fatalf("defaultNoteServiceFactory = %v, %v", service, err)
	}

	plain := &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error { return errors.New("plain") }}
	plain.SetOut(io.Discard)
	plain.SetErr(io.Discard)
	if exit := executeRoot(plain); exit != 10 {
		t.Fatalf("plain root exit = %d", exit)
	}
	rendered := &cobra.Command{Use: "test", RunE: func(cmd *cobra.Command, _ []string) error {
		return renderEnvelope(cmd, "test", "req", func() (any, error) { return nil, protocol.NewError(protocol.RevisionConflict, "conflict", nil) })
	}}
	rendered.SetOut(io.Discard)
	rendered.SetErr(io.Discard)
	if exit := executeRoot(rendered); exit != 4 {
		t.Fatalf("rendered root exit = %d", exit)
	}
	errValue := &renderedCommandError{cause: errors.New("cause"), exit: 10}
	if errValue.Error() != "cause" || !errors.Is(errValue, errValue.cause) {
		t.Fatalf("rendered error methods failed: %v", errValue)
	}
}

func TestDoctorSkillAuditBranches(t *testing.T) {
	var checks []doctorCheck
	auditSkills(filepath.Join(t.TempDir(), "missing"), "v1.0.0", func(check doctorCheck) { checks = append(checks, check) })
	if len(checks) != 1 || checks[0].Status != "warning" {
		t.Fatalf("missing skills audit = %#v", checks)
	}

	root := t.TempDir()
	for index, name := range officialSkillNames {
		if index == 0 {
			continue
		}
		directory := filepath.Join(root, name)
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		content := []byte("skill " + name)
		if err := os.WriteFile(filepath.Join(directory, "SKILL.md"), content, 0o644); err != nil {
			t.Fatal(err)
		}
		if index == 1 {
			continue
		}
		metadataPath := filepath.Join(directory, ".obs-cli-managed.json")
		if index == 2 {
			if err := os.WriteFile(metadataPath, []byte("{"), 0o600); err != nil {
				t.Fatal(err)
			}
			continue
		}
		digest := sha256.Sum256(content)
		metadata := managedSkillMetadata{ManagedBy: "obs-cli", Version: "v1.0.0", Source: "test", Skill: name, SkillDigest: "sha256:" + hex.EncodeToString(digest[:])}
		switch index {
		case 3:
			metadata.ManagedBy = "other"
		case 4:
			metadata.SkillDigest = "sha256:changed"
		case 5:
			metadata.Version = "v0.9.0"
		}
		raw, err := json.Marshal(metadata)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(metadataPath, raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	checks = nil
	auditSkills(root, "v1.0.0", func(check doctorCheck) { checks = append(checks, check) })
	statuses := map[string]int{}
	for _, check := range checks {
		statuses[check.Status]++
	}
	if statuses["error"] < 2 || statuses["warning"] < 3 || statuses["ok"] < 1 {
		t.Fatalf("audit branch statuses = %#v; checks=%#v", statuses, checks)
	}
}

func TestResolveAgentDefaultPathsAndCommandHelpers(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CODEX_HOME", filepath.Join(root, "codex"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "xdg"))
	t.Setenv("KIMI_CODE_HOME", filepath.Join(root, "kimi"))
	for _, agent := range []string{"codex", "claude-code", "opencode", "cursor", "kimi-code"} {
		resolved, path, err := resolveAgentSkillsPath(agent, "")
		if err != nil || resolved != agent || !filepath.IsAbs(path) {
			t.Fatalf("resolveAgentSkillsPath(%s) = %q, %q, %v", agent, resolved, path, err)
		}
	}
	if _, _, err := resolveAgentSkillsPath("unknown", ""); err == nil {
		t.Fatal("unknown Agent should fail")
	}

	command := &cobra.Command{Use: "child"}
	common := commonFlags{RequestID: "bad id"}
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	if err := namespaceArgs(&common, "test.child", cobra.ExactArgs(1))(command, nil); err == nil {
		t.Fatal("invalid namespace arguments should fail")
	}
	if err := namespaceArgs(&common, "test.child", cobra.ExactArgs(1))(command, []string{"ok"}); err != nil {
		t.Fatalf("valid namespace arguments: %v", err)
	}

	if _, _, err := effectiveRevision(nil, "a.md", "revision", true); err == nil {
		t.Fatal("conflicting revision flags should fail")
	}
	if _, _, err := effectiveRevision(nil, "a.md", "", false); err == nil {
		t.Fatal("missing revision should fail")
	}
	if revision, warnings, err := effectiveRevision(nil, "a.md", "revision", false); err != nil || revision != "revision" || len(warnings) != 0 {
		t.Fatalf("explicit revision = %q, %#v, %v", revision, warnings, err)
	}
}

func TestReadNoteInputBranches(t *testing.T) {
	command := &cobra.Command{}
	if _, err := readNoteInput(command, "", "content-file"); err == nil {
		t.Fatal("missing input should fail")
	}
	command.SetIn(bytes.NewBufferString("stdin"))
	if data, err := readNoteInput(command, "-", "content-file"); err != nil || string(data) != "stdin" {
		t.Fatalf("stdin input = %q, %v", data, err)
	}
	path := filepath.Join(t.TempDir(), "input.md")
	if err := os.WriteFile(path, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if data, err := readNoteInput(command, path, "content-file"); err != nil || string(data) != "file" {
		t.Fatalf("file input = %q, %v", data, err)
	}
	if _, err := readNoteInput(command, filepath.Join(t.TempDir(), "missing"), "content-file"); err == nil {
		t.Fatal("missing file should fail")
	}
}
