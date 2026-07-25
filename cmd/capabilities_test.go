package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/protocol"
)

func TestCapabilitiesGoldenJSON(t *testing.T) {
	response, _, err := executeCapabilities(t, "--output", "json", "--request-id", "req-capabilities")
	if err != nil {
		t.Fatal(err)
	}
	if !response.OK || response.Operation != "capabilities.get" || response.RequestID != "req-capabilities" {
		t.Fatalf("unexpected envelope: %#v", response)
	}
	var data capabilitiesData
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.CLIVersion == "" || len(data.ProtocolVersions) == 0 || len(data.Operations) == 0 {
		t.Fatalf("incomplete capabilities: %#v", data)
	}
	if !data.FeatureFlags["note_operations_v2"] || !data.FeatureFlags["daily_notes_v2"] ||
		!data.FeatureFlags["metadata_v2"] || !data.FeatureFlags["search_v2"] ||
		!data.FeatureFlags["link_inspection_v2"] || !data.FeatureFlags["dry_run_plans"] {
		t.Fatalf("capabilities feature flags are incorrect: %#v", data.FeatureFlags)
	}
	assertCapabilitiesSchema(t, data)
}

func assertCapabilitiesSchema(t *testing.T, data capabilitiesData) {
	t.Helper()
	raw, err := os.ReadFile("../docs/spec/schema/capabilities-v2.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	required, ok := schema["required"].([]any)
	if !ok || len(required) == 0 {
		t.Fatalf("capabilities schema has no required fields: %#v", schema)
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	for _, field := range required {
		if _, exists := object[field.(string)]; !exists {
			t.Fatalf("schema-required field %q missing from golden output", field)
		}
	}
}

func TestCapabilitiesUnsupportedRequirement(t *testing.T) {
	response, diagnostic, err := executeCapabilities(
		t,
		"--require",
		"attachment.import",
		"--request-id",
		"req-unsupported",
	)
	if err == nil {
		t.Fatal("expected unsupported capability error")
	}
	if response.OK || response.Error == nil || response.Error.Code != protocol.CapabilityUnsupported {
		t.Fatalf("unexpected envelope: %#v", response)
	}
	if response.RequestID != "req-unsupported" || !bytes.Contains([]byte(diagnostic), []byte("req-unsupported")) {
		t.Fatalf("request ID mismatch response=%q diagnostic=%q", response.RequestID, diagnostic)
	}
}

type capabilitiesEnvelope struct {
	OK        bool                  `json:"ok"`
	Operation string                `json:"operation"`
	RequestID string                `json:"request_id"`
	Data      json.RawMessage       `json:"data"`
	Error     *protocol.DomainError `json:"error"`
}

func executeCapabilities(t *testing.T, args ...string) (capabilitiesEnvelope, string, error) {
	t.Helper()
	command := newCapabilitiesCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs(args)
	executeErr := command.Execute()

	var response capabilitiesEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}
	return response, stderr.String(), executeErr
}
