package protocol_test

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

func TestGoldenEnvelopesAgainstResponseSchemaContract(t *testing.T) {
	var schema map[string]any
	data, err := os.ReadFile("../../docs/spec/schema/response-v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}

	envelopes := []protocol.Envelope{
		protocol.Success("vault.list", "req-schema-success", map[string]any{"vaults": []any{}}, nil),
		protocol.Failure("note.replace", "req-schema-failure", storage.ErrRevisionConflict, nil),
	}
	for _, envelope := range envelopes {
		raw, err := json.Marshal(envelope)
		if err != nil {
			t.Fatal(err)
		}
		var object map[string]any
		if err := json.Unmarshal(raw, &object); err != nil {
			t.Fatal(err)
		}
		validateEnvelopeSchema(t, schema, object)
	}
}

func validateEnvelopeSchema(t *testing.T, schema, object map[string]any) {
	t.Helper()
	for _, required := range stringList(t, schema["required"]) {
		if _, ok := object[required]; !ok {
			t.Fatalf("schema-required property %q missing: %#v", required, object)
		}
	}
	properties := objectMap(t, schema["properties"])
	if object["protocol_version"] != objectMap(t, properties["protocol_version"])["const"] {
		t.Fatalf("protocol_version violates schema: %#v", object)
	}
	for _, field := range []string{"operation", "request_id"} {
		rule := objectMap(t, properties[field])
		pattern := regexp.MustCompile(rule["pattern"].(string))
		if !pattern.MatchString(object[field].(string)) {
			t.Fatalf("%s violates schema pattern: %q", field, object[field])
		}
	}
	ok := object["ok"].(bool)
	if ok {
		if _, exists := object["data"]; !exists {
			t.Fatal("success schema requires data")
		}
		if _, exists := object["error"]; exists {
			t.Fatal("success schema forbids error")
		}
	} else {
		if _, exists := object["data"]; exists {
			t.Fatal("failure schema forbids data")
		}
		errorObject := objectMap(t, object["error"])
		errorSchema := objectMap(t, objectMap(t, schema["$defs"])["error"])
		for _, required := range stringList(t, errorSchema["required"]) {
			if _, exists := errorObject[required]; !exists {
				t.Fatalf("schema-required error property %q missing", required)
			}
		}
		codeSchema := objectMap(t, objectMap(t, errorSchema["properties"])["code"])
		allowed := stringList(t, codeSchema["enum"])
		if !containsString(allowed, errorObject["code"].(string)) {
			t.Fatalf("error code not present in schema enum: %q", errorObject["code"])
		}
	}
	if _, ok := object["warnings"].([]any); !ok {
		t.Fatalf("warnings violates array schema: %#v", object["warnings"])
	}
}

func objectMap(t *testing.T, value any) map[string]any {
	t.Helper()
	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("schema value is not object: %#v", value)
	}
	return result
}

func stringList(t *testing.T, value any) []string {
	t.Helper()
	values, ok := value.([]any)
	if !ok {
		t.Fatalf("schema value is not array: %#v", value)
	}
	result := make([]string, 0, len(values))
	for _, item := range values {
		result = append(result, item.(string))
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
