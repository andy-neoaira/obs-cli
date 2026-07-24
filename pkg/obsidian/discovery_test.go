package obsidian_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
)

func TestDiscoverObsidianVaultsFromIsStableAndReadOnly(t *testing.T) {
	root := t.TempDir()
	alpha := filepath.Join(root, "Alpha")
	beta := filepath.Join(root, "Beta")
	closed := filepath.Join(root, "Closed")
	for _, path := range []string{alpha, beta, closed} {
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	configPath := filepath.Join(root, "obsidian.json")
	original := []byte(`{
  "vaults": {
    "z": {"path": "` + closed + `", "open": false},
    "b": {"path": "` + beta + `", "open": true},
    "a": {"path": "` + alpha + `", "open": true},
    "missing": {"path": "` + filepath.Join(root, "Missing") + `", "open": true}
  }
}`)
	if err := os.WriteFile(configPath, original, 0o600); err != nil {
		t.Fatal(err)
	}

	first, err := obsidian.DiscoverObsidianVaultsFrom(configPath)
	if err != nil {
		t.Fatal(err)
	}
	second, err := obsidian.DiscoverObsidianVaultsFrom(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("discovery order is unstable:\n%#v\n%#v", first, second)
	}

	gotIDs := make([]string, 0, len(first))
	for _, vault := range first {
		gotIDs = append(gotIDs, vault.SourceID)
	}
	wantIDs := []string{"a", "b", "missing", "z"}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("source IDs = %v, want %v", gotIDs, wantIDs)
	}
	if first[2].Available {
		t.Fatal("missing vault must be marked unavailable")
	}

	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, original) {
		t.Fatal("discovery modified Obsidian config")
	}
}

func TestDiscoverObsidianVaultsFromRejectsMalformedConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "obsidian.json")
	if err := os.WriteFile(path, []byte("{bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := obsidian.DiscoverObsidianVaultsFrom(path); err == nil {
		t.Fatal("expected parse error")
	}
}
