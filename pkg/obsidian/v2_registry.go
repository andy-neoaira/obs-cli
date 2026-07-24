package obsidian

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andy-neoaira/obs-cli/pkg/config"
)

var (
	ErrVaultNotFound      = errors.New("vault not found")
	ErrVaultAlreadyExists = errors.New("vault already registered")
	ErrVaultNameConflict  = errors.New("vault name already registered")
)

type Registry struct {
	store *config.Store
}

type VaultRegistrationPlan struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func NewRegistry(store *config.Store) *Registry {
	return &Registry{store: store}
}

func DefaultRegistry() (*Registry, error) {
	store, err := config.DefaultStore()
	if err != nil {
		return nil, err
	}
	return NewRegistry(store), nil
}

func (r *Registry) List() ([]config.VaultRecord, error) {
	cfg, err := r.store.Load()
	if err != nil {
		return nil, err
	}
	return config.SortedVaults(cfg), nil
}

func (r *Registry) Get(reference string) (config.VaultRecord, error) {
	cfg, err := r.store.Load()
	if err != nil {
		return config.VaultRecord{}, err
	}
	return resolveRegistryVault(cfg, reference)
}

func (r *Registry) Add(vaultPath, requestedName string) (config.VaultRecord, error) {
	plan, err := r.PlanAdd(vaultPath, requestedName)
	if err != nil {
		return config.VaultRecord{}, err
	}

	id, err := generateVaultID()
	if err != nil {
		return config.VaultRecord{}, fmt.Errorf("generate vault id: %w", err)
	}
	record := config.VaultRecord{ID: id, Name: plan.Name, Path: plan.Path}

	_, err = r.store.Update(func(cfg *config.V2Config) error {
		if err := validateVaultRegistration(*cfg, plan); err != nil {
			return err
		}
		cfg.Vaults[id] = record
		return nil
	})
	if err != nil {
		return config.VaultRecord{}, err
	}
	return record, nil
}

func (r *Registry) PlanAdd(vaultPath, requestedName string) (VaultRegistrationPlan, error) {
	path, err := canonicalVaultPath(vaultPath)
	if err != nil {
		return VaultRegistrationPlan{}, err
	}
	name := strings.TrimSpace(requestedName)
	if name == "" {
		name = vaultBaseName(path)
	}
	plan := VaultRegistrationPlan{Name: name, Path: path}
	cfg, err := r.store.Load()
	if err != nil {
		return VaultRegistrationPlan{}, err
	}
	if err := validateVaultRegistration(cfg, plan); err != nil {
		return VaultRegistrationPlan{}, err
	}
	return plan, nil
}

func (r *Registry) Remove(reference string) (config.VaultRecord, error) {
	var removed config.VaultRecord
	_, err := r.store.Update(func(cfg *config.V2Config) error {
		record, err := resolveRegistryVault(*cfg, reference)
		if err != nil {
			return err
		}
		removed = record
		delete(cfg.Vaults, record.ID)
		if cfg.DefaultVaultID == record.ID {
			cfg.DefaultVaultID = ""
		}
		return nil
	})
	return removed, err
}

func (r *Registry) SetDefault(reference string) (config.VaultRecord, error) {
	var selected config.VaultRecord
	_, err := r.store.Update(func(cfg *config.V2Config) error {
		record, err := resolveRegistryVault(*cfg, reference)
		if err != nil {
			return err
		}
		selected = record
		cfg.DefaultVaultID = record.ID
		return nil
	})
	return selected, err
}

func (r *Registry) Default() (config.VaultRecord, error) {
	cfg, err := r.store.Load()
	if err != nil {
		return config.VaultRecord{}, err
	}
	if cfg.DefaultVaultID == "" {
		return config.VaultRecord{}, fmt.Errorf("%w: no default vault configured", ErrVaultNotFound)
	}
	return resolveRegistryVault(cfg, cfg.DefaultVaultID)
}

func (r *Registry) SetDefaultOpenType(openType string) error {
	_, err := r.store.Update(func(cfg *config.V2Config) error {
		cfg.DefaultOpenType = openType
		return nil
	})
	return err
}

func (r *Registry) MigrateLegacy(legacyPath string, discovered []DiscoveredVault) (config.V2Config, error) {
	legacy, err := readLegacyConfig(legacyPath)
	if err != nil {
		return config.V2Config{}, err
	}

	return r.store.Update(func(cfg *config.V2Config) error {
		return applyLegacyMigration(cfg, legacy, discovered)
	})
}

func (r *Registry) PlanMigrate(legacyPath string, discovered []DiscoveredVault) (config.V2Config, error) {
	legacy, err := readLegacyConfig(legacyPath)
	if err != nil {
		return config.V2Config{}, err
	}
	cfg, err := r.store.Load()
	if err != nil {
		return config.V2Config{}, err
	}
	planned := cloneV2Config(cfg)
	if err := applyLegacyMigration(&planned, legacy, discovered); err != nil {
		return config.V2Config{}, err
	}
	if err := config.ValidateV2Config(planned); err != nil {
		return config.V2Config{}, err
	}
	return planned, nil
}

func validateVaultRegistration(cfg config.V2Config, plan VaultRegistrationPlan) error {
	for _, current := range cfg.Vaults {
		if sameVaultPath(current.Path, plan.Path) {
			return fmt.Errorf("%w: %s", ErrVaultAlreadyExists, plan.Path)
		}
		if strings.EqualFold(current.Name, plan.Name) {
			return fmt.Errorf("%w: %s", ErrVaultNameConflict, plan.Name)
		}
	}
	return nil
}

func readLegacyConfig(legacyPath string) (CliConfig, error) {
	legacy := CliConfig{}
	data, err := os.ReadFile(legacyPath)
	if err == nil {
		if err := json.Unmarshal(data, &legacy); err != nil {
			return CliConfig{}, fmt.Errorf("parse legacy obs-cli config %s: %w", legacyPath, err)
		}
		return legacy, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return CliConfig{}, fmt.Errorf("read legacy obs-cli config %s: %w", legacyPath, err)
	}
	return legacy, nil
}

func applyLegacyMigration(cfg *config.V2Config, legacy CliConfig, discovered []DiscoveredVault) error {
	if len(cfg.Vaults) != 0 {
		return errors.New("V2 config already contains vaults; migration refused")
	}

	defaultMatches := make([]string, 0, 1)
	for _, item := range discovered {
		if !item.Available {
			continue
		}
		path, err := canonicalVaultPath(item.Path)
		if err != nil {
			return err
		}
		plan := VaultRegistrationPlan{Name: item.Name, Path: path}
		if err := validateVaultRegistration(*cfg, plan); err != nil {
			return fmt.Errorf("%w during migration", err)
		}

		id := deterministicVaultID(path)
		cfg.Vaults[id] = config.VaultRecord{ID: id, Name: item.Name, Path: path}
		if item.Name == legacy.DefaultVaultName {
			defaultMatches = append(defaultMatches, id)
		}
	}
	if len(defaultMatches) == 1 {
		cfg.DefaultVaultID = defaultMatches[0]
	} else if legacy.DefaultVaultName != "" {
		cfg.MigrationWarning = fmt.Sprintf("legacy default vault %q could not be resolved uniquely", legacy.DefaultVaultName)
	}
	cfg.DefaultOpenType = legacy.DefaultOpenType
	cfg.MigratedFrom = "preferences.json+obsidian.json"
	return nil
}

func cloneV2Config(cfg config.V2Config) config.V2Config {
	cloned := cfg
	cloned.Vaults = make(map[string]config.VaultRecord, len(cfg.Vaults))
	for id, record := range cfg.Vaults {
		cloned.Vaults[id] = record
	}
	return cloned
}

func resolveRegistryVault(cfg config.V2Config, reference string) (config.VaultRecord, error) {
	if record, ok := cfg.Vaults[reference]; ok {
		return record, nil
	}
	for _, record := range cfg.Vaults {
		if record.Name == reference {
			return record, nil
		}
	}
	return config.VaultRecord{}, fmt.Errorf("%w: %s", ErrVaultNotFound, reference)
}

func canonicalVaultPath(input string) (string, error) {
	absolute, err := filepath.Abs(input)
	if err != nil {
		return "", fmt.Errorf("resolve vault path: %w", err)
	}
	realPath, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", fmt.Errorf("resolve vault symlinks: %w", err)
	}
	info, err := os.Stat(realPath)
	if err != nil {
		return "", fmt.Errorf("stat vault path: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("vault path is not a directory: %s", realPath)
	}
	return filepath.Clean(realPath), nil
}

func sameVaultPath(left, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if left == right {
		return true
	}
	// Conservative portability rule: paths differing only by case conflict.
	return strings.EqualFold(left, right)
}

func deterministicVaultID(path string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(path)))
	return "vlt_" + hex.EncodeToString(sum[:16])
}

func generateVaultID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return "vlt_" + hex.EncodeToString(value), nil
}
