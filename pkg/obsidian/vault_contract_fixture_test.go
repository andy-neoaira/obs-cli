package obsidian_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVaultContractFixtureManifest(t *testing.T) {
	fixtureRoot := "../../testdata/vault-contract/v1"
	data, err := os.ReadFile(filepath.Join(fixtureRoot, "manifest.json"))
	if err != nil {
		t.Fatalf("read fixture manifest: %v", err)
	}

	var manifest struct {
		Contract             string   `json:"contract"`
		FixtureVersion       int      `json:"fixture_version"`
		ContainsPersonalData bool     `json:"contains_personal_data"`
		Cases                []string `json:"cases"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode fixture manifest: %v", err)
	}
	if manifest.Contract != "vault-contract/v1" {
		t.Fatalf("unexpected contract: %s", manifest.Contract)
	}
	if manifest.FixtureVersion != 1 {
		t.Fatalf("unexpected fixture version: %d", manifest.FixtureVersion)
	}
	if manifest.ContainsPersonalData {
		t.Fatal("contract fixture must not contain personal data")
	}
	if len(manifest.Cases) == 0 {
		t.Fatal("contract fixture has no cases")
	}

	seen := make(map[string]bool, len(manifest.Cases))
	for _, name := range manifest.Cases {
		if seen[name] {
			t.Fatalf("duplicate fixture case: %s", name)
		}
		seen[name] = true

		expectedPath := filepath.Join(fixtureRoot, name, "expected.json")
		expectedData, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("read %s: %v", expectedPath, err)
		}
		var expected struct {
			Contract string `json:"contract"`
			Case     string `json:"case"`
		}
		if err := json.Unmarshal(expectedData, &expected); err != nil {
			t.Fatalf("decode %s: %v", expectedPath, err)
		}
		if expected.Contract != manifest.Contract {
			t.Errorf("%s contract = %q, want %q", name, expected.Contract, manifest.Contract)
		}
		if expected.Case != name {
			t.Errorf("%s case = %q", name, expected.Case)
		}
	}
}
