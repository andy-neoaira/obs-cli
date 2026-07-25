package cmd

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

func TestAgentResultSchemaEnforcesRecoveryContract(t *testing.T) {
	raw, err := os.ReadFile("../docs/spec/schema/agent-result-v1.schema.json")
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
	golden := map[string]any{
		"schema_version": "miniobsidian.agent-result/v1",
		"request_id":     "handoff-update-1",
		"status":         "success",
		"summary":        "updated current note",
		"changes": []any{map[string]any{
			"path":            "Notes/Status.md",
			"revision_before": "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"revision_after":  "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			"summary":         "updated status",
			"before_content":  "# Status\n\nBefore\n",
		}},
		"errors": []any{},
	}
	if err := resolved.Validate(golden); err != nil {
		t.Fatalf("golden Agent result rejected: %v", err)
	}

	cancelled := cloneJSONMap(t, golden)
	cancelled["status"] = "cancelled"
	if err := resolved.Validate(cancelled); err == nil {
		t.Fatal("cancelled result accepted changed files")
	}

	partial := cloneJSONMap(t, golden)
	partial["status"] = "partial"
	if err := resolved.Validate(partial); err == nil {
		t.Fatal("partial result accepted without recovery errors")
	}

	outside := cloneJSONMap(t, golden)
	outside["changes"].([]any)[0].(map[string]any)["path"] = "outside.txt"
	if err := resolved.Validate(outside); err == nil {
		t.Fatal("Agent result accepted a non-Markdown changed file")
	}
}

func cloneJSONMap(t *testing.T, input map[string]any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var output map[string]any
	if err := json.Unmarshal(raw, &output); err != nil {
		t.Fatal(err)
	}
	return output
}
