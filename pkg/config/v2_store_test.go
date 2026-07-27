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

func TestV2StoreMissingReturnsEmptyConfig(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "config-v2.json"))
	cfg, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != config.CurrentConfigVersion || len(cfg.Vaults) != 0 {
		t.Fatalf("unexpected empty config: %#v", cfg)
	}
}

func TestV2StoreRejectsMalformedConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config-v2.json")
	if err := os.WriteFile(path, []byte("{bad"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.NewStore(path).Load(); err == nil {
		t.Fatal("expected malformed config error")
	}
}

func TestV2StoreRejectsDuplicateNamesAndPaths(t *testing.T) {
	base := t.TempDir()
	cfg := config.NewV2Config()
	cfg.Vaults["one"] = config.VaultRecord{ID: "one", Name: "Notes", Path: filepath.Join(base, "one")}
	cfg.Vaults["two"] = config.VaultRecord{ID: "two", Name: "notes", Path: filepath.Join(base, "two")}
	if err := config.ValidateV2Config(cfg); err == nil {
		t.Fatal("expected duplicate name error")
	}

	cfg.Vaults["two"] = config.VaultRecord{ID: "two", Name: "Other", Path: filepath.Join(base, "one")}
	if err := config.ValidateV2Config(cfg); err == nil {
		t.Fatal("expected duplicate path error")
	}
}

func TestV2StoreConcurrentUpdatesDoNotLoseData(t *testing.T) {
	store := config.NewStore(filepath.Join(t.TempDir(), "config-v2.json"))
	const writers = 12

	var wg sync.WaitGroup
	errs := make(chan error, writers)
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			id := string(rune('a' + index))
			_, err := store.Update(func(cfg *config.V2Config) error {
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

func TestV2StoreUsesDurableAtomicReplace(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config-v2.json")
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

	_, err := store.Update(func(cfg *config.V2Config) error {
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
	var decoded config.V2Config
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
		if entry.Name() == "config-v2.json.lock" || strings.HasPrefix(entry.Name(), ".obs-replace-") {
			t.Fatalf("temporary config artifact leaked: %s", entry.Name())
		}
	}
}
