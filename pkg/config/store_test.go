package config_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

func TestStoreMissingReturnsEmptyConfig(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	cfg, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != config.CurrentConfigVersion || len(cfg.Vaults) != 0 {
		t.Fatalf("unexpected empty config: %#v", cfg)
	}
}

func TestStoreRejectsMalformedConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.NewStore(path).Load(); err == nil {
		t.Fatal("expected malformed config error")
	}
}

func TestStoreRejectsUnsupportedVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"version":2,"vaults":{}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.NewStore(path).Load(); !errors.Is(err, config.ErrUnsupportedVersion) {
		t.Fatalf("Load error = %v, want unsupported version", err)
	}
}

func TestDefaultStoreDoesNotReadOrMigrateHistoricalConfig(t *testing.T) {
	root := t.TempDir()
	t.Setenv("OBS_CLI_CONFIG_HOME", root)
	configDir := filepath.Join(root, "obs-cli")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	historicalPath := filepath.Join(configDir, "config-"+"v2.json")
	historical := `{"version":2,"vaults":{"old":{"id":"old","name":"Old","path":"/tmp/old"}}}`
	if err := os.WriteFile(historicalPath, []byte(historical), 0o600); err != nil {
		t.Fatal(err)
	}

	store, err := config.DefaultStore()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 1 || len(cfg.Vaults) != 0 {
		t.Fatalf("historical config was loaded: %#v", cfg)
	}
	if _, err := os.Stat(filepath.Join(configDir, "config.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Load created or migrated config: %v", err)
	}
}

func TestStoreRejectsDuplicateNamesAndPaths(t *testing.T) {
	base := t.TempDir()
	cfg := config.NewConfig()
	cfg.Vaults["one"] = config.VaultRecord{ID: "one", Name: "Notes", Path: filepath.Join(base, "one")}
	cfg.Vaults["two"] = config.VaultRecord{ID: "two", Name: "notes", Path: filepath.Join(base, "two")}
	if err := config.ValidateConfig(cfg); err == nil {
		t.Fatal("expected duplicate name error")
	}

	cfg.Vaults["two"] = config.VaultRecord{ID: "two", Name: "Other", Path: filepath.Join(base, "one")}
	if err := config.ValidateConfig(cfg); err == nil {
		t.Fatal("expected duplicate path error")
	}
}

func TestValidateConfigAndSortedVaults(t *testing.T) {
	root := t.TempDir()
	valid := config.NewConfig()
	valid.Vaults["z"] = config.VaultRecord{ID: "z", Name: "Zulu", Path: filepath.Join(root, "z")}
	valid.Vaults["a"] = config.VaultRecord{ID: "a", Name: "Alpha", Path: filepath.Join(root, "a")}
	valid.DefaultVaultID = "a"
	if err := config.ValidateConfig(valid); err != nil {
		t.Fatal(err)
	}
	sorted := config.SortedVaults(valid)
	if len(sorted) != 2 || sorted[0].ID != "a" || sorted[1].ID != "z" {
		t.Fatalf("SortedVaults = %#v", sorted)
	}

	tests := []struct {
		name   string
		mutate func(*config.Config)
	}{
		{"nil vault map", func(cfg *config.Config) { cfg.Vaults = nil }},
		{"mismatched id", func(cfg *config.Config) {
			cfg.Vaults["a"] = config.VaultRecord{ID: "b", Name: "A", Path: filepath.Join(root, "a")}
		}},
		{"blank name", func(cfg *config.Config) {
			cfg.Vaults["a"] = config.VaultRecord{ID: "a", Name: " ", Path: filepath.Join(root, "a")}
		}},
		{"relative path", func(cfg *config.Config) { cfg.Vaults["a"] = config.VaultRecord{ID: "a", Name: "A", Path: "relative"} }},
		{"unclean path", func(cfg *config.Config) {
			cfg.Vaults["a"] = config.VaultRecord{ID: "a", Name: "A", Path: filepath.Join(root, "folder", "..", "a") + string(filepath.Separator)}
		}},
		{"missing default", func(cfg *config.Config) { cfg.DefaultVaultID = "missing" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := config.NewConfig()
			test.mutate(&cfg)
			if err := config.ValidateConfig(cfg); err == nil {
				t.Fatalf("ValidateConfig(%s) succeeded", test.name)
			}
		})
	}
}

func TestStoreConcurrentUpdatesDoNotLoseData(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	const writers = 12

	var wg sync.WaitGroup
	errs := make(chan error, writers)
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			id := string(rune('a' + index))
			_, err := store.Update(func(cfg *config.Config) error {
				cfg.Vaults[id] = config.VaultRecord{
					ID:   id,
					Name: "Vault-" + id,
					Path: filepath.Join(t.TempDir(), id),
				}
				return nil
			})
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	cfg, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Vaults) != writers {
		t.Fatalf("got %d vaults, want %d", len(cfg.Vaults), writers)
	}

	data, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("written config is not valid JSON: %v", err)
	}
}

func TestStoreIgnoresAbandonedCoordinationFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "config.json")
	lockPath := path + ".lock"
	if err := os.WriteFile(lockPath, []byte("abandoned by crashed process\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	store := config.NewStore(path)
	cfg, err := store.Update(func(cfg *config.Config) error {
		cfg.Vaults["one"] = config.VaultRecord{
			ID: "one", Name: "One", Path: filepath.Join(root, "vault"),
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Update with abandoned lock file: %v", err)
	}
	if cfg.Vaults["one"].Name != "One" {
		t.Fatalf("config was not updated: %#v", cfg)
	}
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("coordination file should remain reusable: %v", err)
	}
}

func TestStoreUsesDurableAtomicReplace(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.json")
	injected := errors.New("injected directory sync")
	atomic := storage.NewStoreWithHooks(filepath.Join(root, "locks"), storage.Hooks{
		Checkpoint: func(stage string) error {
			if stage == "directory-sync-before" {
				return injected
			}
			return nil
		},
	})
	store := config.NewStoreWithAtomicStore(configPath, atomic)

	_, err := store.Update(func(cfg *config.Config) error {
		cfg.Vaults["one"] = config.VaultRecord{
			ID: "one", Name: "One", Path: filepath.Join(root, "vault"),
		}
		return nil
	})
	if !errors.Is(err, injected) {
		t.Fatalf("Update error = %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var decoded config.Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("post-commit config is incomplete: %v", err)
	}
	if decoded.Vaults["one"].Name != "One" {
		t.Fatalf("post-commit config = %#v", decoded)
	}
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config mode = %o, want 600", info.Mode().Perm())
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".obs-replace-") {
			t.Fatalf("temporary config artifact leaked: %s", entry.Name())
		}
	}
}
