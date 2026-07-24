package protocol_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/pathpolicy"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

func TestJSONSuccessAndFailureEnvelope(t *testing.T) {
	for _, envelope := range []protocol.Envelope{
		protocol.Success("vault.list", "req-test", map[string]any{"vaults": []any{}}, nil),
		protocol.Failure("note.replace", "req-test", storage.ErrRevisionConflict, nil),
	} {
		var output bytes.Buffer
		if err := protocol.Render(&output, envelope); err != nil {
			t.Fatal(err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(output.Bytes(), &decoded); err != nil {
			t.Fatalf("stdout is not one JSON object: %v\n%s", err, output.String())
		}
		if decoded["protocol_version"] != protocol.Version || decoded["request_id"] != "req-test" {
			t.Fatalf("invalid envelope: %#v", decoded)
		}
		if warnings, ok := decoded["warnings"].([]any); !ok || len(warnings) != 0 {
			t.Fatalf("warnings must be an empty array: %#v", decoded["warnings"])
		}
		if envelope.OK {
			if _, ok := decoded["error"]; ok {
				t.Fatal("success envelope contains error")
			}
		} else if _, ok := decoded["data"]; ok {
			t.Fatal("failure envelope contains data")
		}
	}
}

func TestErrorCodeAndExitCodeMapping(t *testing.T) {
	cases := []struct {
		err  error
		code protocol.Code
		exit int
	}{
		{storage.ErrRevisionConflict, protocol.RevisionConflict, 4},
		{storage.ErrAlreadyExists, protocol.AlreadyExists, 4},
		{storage.ErrPartialFailure, protocol.PartialFailure, 7},
		{pathpolicy.ErrOutsideVault, protocol.PathOutsideVault, 5},
		{errors.New("unknown"), protocol.InternalError, 10},
	}
	for _, item := range cases {
		if actual := protocol.MapError(item.err).Code; actual != item.code {
			t.Fatalf("MapError(%v) = %s, want %s", item.err, actual, item.code)
		}
		if actual := protocol.ExitCodeFor(item.err); actual != item.exit {
			t.Fatalf("ExitCodeFor(%v) = %d, want %d", item.err, actual, item.exit)
		}
	}
}

func TestEveryStableErrorCodeHasExitCode(t *testing.T) {
	cases := []struct {
		code protocol.Code
		exit int
	}{
		{protocol.InvalidArgument, 2},
		{protocol.InvalidFrontmatter, 2},
		{protocol.VaultNotFound, 3},
		{protocol.NoteNotFound, 3},
		{protocol.AlreadyExists, 4},
		{protocol.AmbiguousNote, 4},
		{protocol.RevisionConflict, 4},
		{protocol.PathOutsideVault, 5},
		{protocol.PartialFailure, 7},
		{protocol.CapabilityUnsupported, 8},
		{protocol.InternalError, 10},
	}
	for _, item := range cases {
		err := protocol.NewError(item.code, "test", nil)
		if actual := protocol.ExitCodeFor(err); actual != item.exit {
			t.Fatalf("ExitCodeFor(%s) = %d, want %d", item.code, actual, item.exit)
		}
	}
}

func TestRequestIDValidation(t *testing.T) {
	valid := []string{"req-123", "agent.run:42", strings.Repeat("a", 128)}
	for _, input := range valid {
		actual, err := protocol.ResolveRequestID(input)
		if err != nil || actual != input {
			t.Fatalf("ResolveRequestID(%q) = %q, %v", input, actual, err)
		}
	}
	for _, input := range []string{"bad id", "换行", strings.Repeat("a", 129)} {
		if _, err := protocol.ResolveRequestID(input); err == nil {
			t.Fatalf("ResolveRequestID(%q) unexpectedly succeeded", input)
		}
	}
	generated, err := protocol.ResolveRequestID("")
	if err != nil || !strings.HasPrefix(generated, "req-") {
		t.Fatalf("generated request ID = %q, %v", generated, err)
	}
}
