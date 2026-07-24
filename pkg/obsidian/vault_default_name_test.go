package obsidian_test

import (
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
)

func TestVaultUsesV2DefaultConfiguration(t *testing.T) {
	original := config.UserConfigDirectory
	configRoot := t.TempDir()
	config.UserConfigDirectory = func() (string, error) { return configRoot, nil }
	t.Cleanup(func() { config.UserConfigDirectory = original })

	registry, err := obsidian.DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	added, err := registry.Add(t.TempDir(), "Personal")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.SetDefault(added.ID); err != nil {
		t.Fatal(err)
	}

	vault := &obsidian.Vault{}
	name, err := vault.DefaultName()
	if err != nil || name != "Personal" {
		t.Fatalf("default name = %q, %v", name, err)
	}
	path, err := vault.Path()
	if err != nil || path != added.Path {
		t.Fatalf("path = %q, %v", path, err)
	}

	if err := vault.SetDefaultOpenType("editor"); err != nil {
		t.Fatal(err)
	}
	openType, err := vault.DefaultOpenType()
	if err != nil || openType != "editor" {
		t.Fatalf("open type = %q, %v", openType, err)
	}
}

func TestVaultExplicitNameUsesV2Registry(t *testing.T) {
	original := config.UserConfigDirectory
	configRoot := t.TempDir()
	config.UserConfigDirectory = func() (string, error) { return configRoot, nil }
	t.Cleanup(func() { config.UserConfigDirectory = original })

	registry, err := obsidian.DefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	added, err := registry.Add(t.TempDir(), "Work")
	if err != nil {
		t.Fatal(err)
	}

	vault := &obsidian.Vault{Name: "Work"}
	path, err := vault.Path()
	if err != nil || path != added.Path {
		t.Fatalf("path = %q, %v", path, err)
	}
}
