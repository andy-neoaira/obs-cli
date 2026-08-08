package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/andy-neoaira/obs-cli/pkg/storage"
)

const CurrentConfigVersion = 1

var (
	ErrConfigLocked       = errors.New("obs-cli configuration is locked by another process")
	ErrUnsupportedVersion = errors.New("unsupported obs-cli configuration version")
)

type VaultRecord struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type Config struct {
	Version        int                    `json:"version"`
	DefaultVaultID string                 `json:"default_vault_id,omitempty"`
	Vaults         map[string]VaultRecord `json:"vaults"`
}

func NewConfig() Config {
	return Config{
		Version: CurrentConfigVersion,
		Vaults:  make(map[string]VaultRecord),
	}
}

type Store struct {
	path        string
	lockTimeout time.Duration
	atomic      *storage.Store
}

func NewStore(path string) *Store {
	return NewStoreWithAtomicStore(path, storage.DefaultStore())
}

func NewStoreWithAtomicStore(path string, atomic *storage.Store) *Store {
	if atomic == nil {
		atomic = storage.DefaultStore()
	}
	return &Store{path: path, lockTimeout: 2 * time.Second, atomic: atomic}
}

func DefaultStore() (*Store, error) {
	_, path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	return NewStore(path), nil
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Load() (Config, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return NewConfig(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read obs-cli config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse obs-cli config %s: %w", s.path, err)
	}
	if err := ValidateConfig(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (s *Store) Update(mutate func(*Config) error) (Config, error) {
	if mutate == nil {
		return Config{}, errors.New("config update callback is required")
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return Config{}, fmt.Errorf("create obs-cli config directory: %w", err)
	}

	unlock, err := s.lock()
	if err != nil {
		return Config{}, err
	}
	defer unlock()

	cfg, err := s.Load()
	if err != nil {
		return Config{}, err
	}
	if err := mutate(&cfg); err != nil {
		return Config{}, err
	}
	if err := ValidateConfig(cfg); err != nil {
		return Config{}, err
	}
	if err := s.writeAtomic(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func ValidateConfig(cfg Config) error {
	if cfg.Version != CurrentConfigVersion {
		return fmt.Errorf("%w: got %d, want %d", ErrUnsupportedVersion, cfg.Version, CurrentConfigVersion)
	}
	if cfg.Vaults == nil {
		return errors.New("obs-cli config vaults must not be null")
	}
	records := make([]VaultRecord, 0, len(cfg.Vaults))
	for key, vault := range cfg.Vaults {
		if key == "" || vault.ID == "" || key != vault.ID {
			return fmt.Errorf("vault key %q does not match non-empty vault id %q", key, vault.ID)
		}
		if strings.TrimSpace(vault.Name) == "" {
			return fmt.Errorf("vault %q has an empty name", key)
		}
		if !filepath.IsAbs(vault.Path) {
			return fmt.Errorf("vault %q path must be absolute: %s", key, vault.Path)
		}
		if filepath.Clean(vault.Path) != vault.Path {
			return fmt.Errorf("vault %q path is not normalized: %s", key, vault.Path)
		}
		records = append(records, vault)
	}
	for i := range records {
		for j := i + 1; j < len(records); j++ {
			if strings.EqualFold(records[i].Name, records[j].Name) {
				return fmt.Errorf(
					"duplicate vault name %q for %s and %s",
					records[j].Name,
					records[i].ID,
					records[j].ID,
				)
			}
			if strings.EqualFold(records[i].Path, records[j].Path) {
				return fmt.Errorf(
					"duplicate vault path %q for %s and %s",
					records[j].Path,
					records[i].ID,
					records[j].ID,
				)
			}
		}
	}
	if cfg.DefaultVaultID != "" {
		if _, ok := cfg.Vaults[cfg.DefaultVaultID]; !ok {
			return fmt.Errorf("default vault id %q is not registered", cfg.DefaultVaultID)
		}
	}
	return nil
}

func SortedVaults(cfg Config) []VaultRecord {
	result := make([]VaultRecord, 0, len(cfg.Vaults))
	for _, vault := range cfg.Vaults {
		result = append(result, vault)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		return result[i].Path < result[j].Path
	})
	return result
}

func (s *Store) writeAtomic(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode obs-cli config: %w", err)
	}
	data = append(data, '\n')

	if err := s.atomic.ReplaceAtomic(
		s.path,
		data,
		storage.WriteOptions{Mode: 0o600},
	); err != nil {
		return fmt.Errorf("replace obs-cli config: %w", err)
	}
	return nil
}

func (s *Store) lock() (func(), error) {
	lockPath := s.path + ".lock"
	deadline := time.Now().Add(s.lockTimeout)
	for {
		unlock, acquired, err := tryConfigLock(lockPath)
		if err != nil {
			return nil, fmt.Errorf("acquire config lock: %w", err)
		}
		if acquired {
			return unlock, nil
		}
		if time.Now().After(deadline) {
			return nil, ErrConfigLocked
		}
		time.Sleep(10 * time.Millisecond)
	}
}
