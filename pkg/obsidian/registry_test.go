package obsidian_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
)

func TestRegistryLifecycle(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
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
	byPath, err := registry.Get(added.Path)
	if err != nil || byPath != added {
		t.Fatalf("get by registered path = %#v, %v", byPath, err)
	}
	unregisteredPath := t.TempDir()
	if _, err := registry.Get(unregisteredPath); !errors.Is(err, obsidian.ErrVaultNotFound) {
		t.Fatalf("get by unregistered path error = %v", err)
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

func TestRegistryRejectsDuplicateCanonicalPathAndName(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
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

func TestRegistryCanonicalizesSymlink(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
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
	byLink, err := registry.Get(link)
	if err != nil || byLink != added {
		t.Fatalf("get by registered symlink alias = %#v, %v", byLink, err)
	}
}

func TestRegistryListAndDefaultFactory(t *testing.T) {
	root := t.TempDir()
	t.Setenv("OBS_CLI_CONFIG_HOME", root)
	registry, err := obsidian.DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if records, err := registry.List(); err != nil || len(records) != 0 {
		t.Fatalf("empty List = %#v, %v", records, err)
	}
	if _, err := registry.Add(t.TempDir(), "Zulu"); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Add(t.TempDir(), "Alpha"); err != nil {
		t.Fatal(err)
	}
	records, err := registry.List()
	if err != nil || len(records) != 2 || records[0].Name != "Alpha" || records[1].Name != "Zulu" {
		t.Fatalf("sorted List = %#v, %v", records, err)
	}
}
