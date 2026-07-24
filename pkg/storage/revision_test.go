package storage_test

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

func TestRevisionContractVectors(t *testing.T) {
	var fixture struct {
		Vectors []struct {
			Name          string `json:"name"`
			ContentBase64 string `json:"content_base64"`
			Revision      string `json:"revision"`
		} `json:"vectors"`
	}
	data, err := os.ReadFile("../../testdata/revision/revision-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, vector := range fixture.Vectors {
		t.Run(vector.Name, func(t *testing.T) {
			content, err := base64.StdEncoding.DecodeString(vector.ContentBase64)
			if err != nil {
				t.Fatal(err)
			}
			if actual := storage.Revision(content); actual != vector.Revision {
				t.Fatalf("Revision() = %s, want %s", actual, vector.Revision)
			}
		})
	}
}

func TestIsRevision(t *testing.T) {
	valid := storage.Revision([]byte("content"))
	if !storage.IsRevision(valid) {
		t.Fatalf("valid revision rejected: %s", valid)
	}
	for _, value := range []string{"", "sha256:stale", strings.ToUpper(valid), valid + "0"} {
		if storage.IsRevision(value) {
			t.Fatalf("invalid revision accepted: %q", value)
		}
	}
}
