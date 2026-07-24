package obsidian_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
)

func TestV2RegistryLifecycle(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "config-v2.json"))
	registry := obsidian.NewRegistry(store)
	vaultPath := t.TempDir()

	added, err := registry.Add(vaultPath, "Personal")
	if err != nil {
		t.Fatal(err)
	}
	if added.ID == "" || added.Name != "Personal" {
		t.Fatalf("unexpected record: %#v", added)
	}

	byID, err := registry.Get(added.ID)
	if err != nil || byID != added {
		t.Fatalf("get by ID = %#v, %v", byID, err)
	}
	byName, err := registry.Get("Personal")
	if err != nil || byName != added {
		t.Fatalf("get by name = %#v, %v", byName, err)
	}

	selected, err := registry.SetDefault(added.ID)
	if err != nil || selected != added {
		t.Fatalf("set default = %#v, %v", selected, err)
	}
	defaultVault, err := registry.Default()
	if err != nil || defaultVault != added {
		t.Fatalf("default = %#v, %v", defaultVault, err)
	}

	removed, err := registry.Remove("Personal")
	if err != nil || removed != added {
		t.Fatalf("remove = %#v, %v", removed, err)
	}
	if _, err := registry.Default(); !errors.Is(err, obsidian.ErrVaultNotFound) {
		t.Fatalf("default after remove error = %v", err)
	}
}

func TestV2RegistryRejectsDuplicateCanonicalPathAndName(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "config-v2.json"))
	registry := obsidian.NewRegistry(store)
	firstPath := t.TempDir()
	secondPath := t.TempDir()

	if _, err := registry.Add(firstPath, "Notes"); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Add(firstPath, "Other"); !errors.Is(err, obsidian.ErrVaultAlreadyExists) {
		t.Fatalf("duplicate path error = %v", err)
	}
	if _, err := registry.Add(secondPath, "notes"); !errors.Is(err, obsidian.ErrVaultNameConflict) {
		t.Fatalf("duplicate name error = %v", err)
	}
}

func TestV2RegistryCanonicalizesSymlink(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "config-v2.json"))
	registry := obsidian.NewRegistry(store)
	root := t.TempDir()
	target := filepath.Join(root, "target")
	link := filepath.Join(root, "link")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	added, err := registry.Add(link, "Linked")
	if err != nil {
		t.Fatal(err)
	}
	canonicalTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		t.Fatal(err)
	}
	if added.Path != canonicalTarget {
		t.Fatalf("canonical path = %q, want %q", added.Path, canonicalTarget)
	}
}

func TestV2RegistryMigratesLegacyConfigAndDiscovery(t *testing.T) {
	root := t.TempDir()
	personal := filepath.Join(root, "Personal")
	work := filepath.Join(root, "Work")
	for _, path := range []string{personal, work} {
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	legacyPath := filepath.Join(root, "preferences.json")
	if err := os.WriteFile(legacyPath, []byte(`{"default_vault_name":"Work","default_open_type":"editor"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	store := config.NewStore(filepath.Join(root, "config-v2.json"))
	registry := obsidian.NewRegistry(store)
	cfg, err := registry.MigrateLegacy(legacyPath, []obsidian.DiscoveredVault{
		{SourceID: "a", Name: "Personal", Path: personal, Available: true},
		{SourceID: "b", Name: "Work", Path: work, Available: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Vaults) != 2 || cfg.DefaultOpenType != "editor" {
		t.Fatalf("unexpected migrated config: %#v", cfg)
	}
	if cfg.DefaultVaultID == "" || cfg.Vaults[cfg.DefaultVaultID].Name != "Work" {
		t.Fatalf("default vault was not migrated: %#v", cfg)
	}

	if _, err := registry.MigrateLegacy(legacyPath, nil); err == nil {
		t.Fatal("repeat migration must be refused")
	}
}

func TestV2RegistryRejectsMalformedLegacyConfig(t *testing.T) {
	root := t.TempDir()
	legacyPath := filepath.Join(root, "preferences.json")
	if err := os.WriteFile(legacyPath, []byte("{bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	registry := obsidian.NewRegistry(config.NewStore(filepath.Join(root, "config-v2.json")))
	if _, err := registry.MigrateLegacy(legacyPath, nil); err == nil {
		t.Fatal("expected malformed legacy config error")
	}
}
