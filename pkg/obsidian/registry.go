package obsidian

import (
	"crypto/rand"
	"encoding/hex"
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

	_, err = r.store.Update(func(cfg *config.Config) error {
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
	_, err := r.store.Update(func(cfg *config.Config) error {
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
	_, err := r.store.Update(func(cfg *config.Config) error {
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

func validateVaultRegistration(cfg config.Config, plan VaultRegistrationPlan) error {
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

func resolveRegistryVault(cfg config.Config, reference string) (config.VaultRecord, error) {
	if record, ok := cfg.Vaults[reference]; ok {
		return record, nil
	}
	for _, record := range cfg.Vaults {
		if record.Name == reference {
			return record, nil
		}
	}
	if filepath.IsAbs(reference) {
		candidate := filepath.Clean(reference)
		if realPath, err := filepath.EvalSymlinks(candidate); err == nil {
			candidate = filepath.Clean(realPath)
		}
		for _, record := range cfg.Vaults {
			if sameVaultPath(record.Path, candidate) {
				return record, nil
			}
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

func generateVaultID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return "vlt_" + hex.EncodeToString(value), nil
}
