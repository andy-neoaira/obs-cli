package cmd

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

func TestAgentHandoffSchemaEnforcesPermissionBoundary(t *testing.T) {
	raw, err := os.ReadFile("../docs/spec/schema/agent-handoff-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema jsonschema.Schema
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	resolved, err := schema.Resolve(nil)
	if err != nil {
		t.Fatal(err)
	}
	base := func(mode string) map[string]any {
		writePaths := []any{}
		skill := "obsidian-knowledge-synthesis"
		capabilities := []any{"note.get"}
		if mode == "update" {
			writePaths = []any{"Notes/Source.md"}
			skill = "obsidian-safe-note-update"
			capabilities = []any{"note.get", "note.patch"}
		}
		return map[string]any{
			"schema_version": "miniobsidian.agent-handoff/v1",
			"request_id":     "handoff-" + mode + "-1",
			"mode":           mode,
			"intent":         "review current note",
			"vault":          map[string]any{"id": "vault-test"},
			"source": map[string]any{
				"path":            "Notes/Source.md",
				"revision":        "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"buffer_modified": false,
			},
			"context": map[string]any{
				"scope": "none", "start_line": nil, "end_line": nil, "text": nil, "line_count": 0,
			},
			"permissions": map[string]any{
				"allow_vault_scan": false,
				"read_paths":       []any{"Notes/Source.md"},
				"write_paths":      writePaths,
			},
			"agent": map[string]any{
				"skill": skill, "required_capabilities": capabilities,
			},
		}
	}

	for _, mode := range []string{"analyze", "update"} {
		if err := resolved.Validate(base(mode)); err != nil {
			t.Fatalf("%s golden handoff rejected: %v", mode, err)
		}
	}

	readonlyWrite := base("analyze")
	readonlyWrite["permissions"].(map[string]any)["write_paths"] = []any{"Notes/Source.md"}
	if err := resolved.Validate(readonlyWrite); err == nil {
		t.Fatal("analyze handoff accepted write permission")
	}

	dirtyUpdate := base("update")
	dirtyUpdate["source"].(map[string]any)["buffer_modified"] = true
	if err := resolved.Validate(dirtyUpdate); err == nil {
		t.Fatal("update handoff accepted a modified buffer")
	}

	wholeVault := base("analyze")
	wholeVault["permissions"].(map[string]any)["allow_vault_scan"] = true
	if err := resolved.Validate(wholeVault); err == nil {
		t.Fatal("handoff accepted whole-Vault scan permission")
	}
}
