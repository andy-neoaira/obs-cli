package obsidian_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestRevisionContractVectors(t *testing.T) {
	type vector struct {
		Name          string `json:"name"`
		ContentBase64 string `json:"content_base64"`
		Revision      string `json:"revision"`
	}
	var fixture struct {
		Vectors []vector `json:"vectors"`
	}

	data, err := os.ReadFile("../../testdata/revision/revision-v1.json")
	if err != nil {
		t.Fatalf("read revision fixture: %v", err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode revision fixture: %v", err)
	}

	for _, item := range fixture.Vectors {
		t.Run(item.Name, func(t *testing.T) {
			content, err := base64.StdEncoding.DecodeString(item.ContentBase64)
			if err != nil {
				t.Fatalf("decode content: %v", err)
			}

			sum := sha256.Sum256(content)
			actual := "sha256:" + hex.EncodeToString(sum[:])
			if actual != item.Revision {
				t.Fatalf("revision mismatch: got %s, want %s", actual, item.Revision)
			}
			if strings.ToLower(actual) != actual {
				t.Fatalf("revision must use lowercase hexadecimal: %s", actual)
			}
		})
	}
}
