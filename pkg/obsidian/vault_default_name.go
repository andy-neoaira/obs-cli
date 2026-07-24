package obsidian

import (
	"github.com/andy-neoaira/obs-cli/pkg/config"
)

// DefaultName 返回 V2 registry 中的默认 Vault 名称。
func (v *Vault) DefaultName() (string, error) {
	if v.Name != "" {
		return v.Name, nil
	}
	registry, err := DefaultRegistry()
	if err != nil {
		return "", err
	}
	record, err := registry.Default()
	if err != nil {
		return "", err
	}
	v.Name = record.Name
	return record.Name, nil
}

// SetDefaultName 通过 ID 或名称设置 V2 registry 的默认 Vault。
func (v *Vault) SetDefaultName(reference string) error {
	registry, err := DefaultRegistry()
	if err != nil {
		return err
	}
	record, err := registry.SetDefault(reference)
	if err != nil {
		return err
	}
	v.Name = record.Name
	return nil
}

func (v *Vault) DefaultOpenType() (string, error) {
	store, err := config.DefaultStore()
	if err != nil {
		return "", err
	}
	cfg, err := store.Load()
	if err != nil {
		return "", err
	}
	if cfg.DefaultOpenType == "" {
		return "obsidian", nil
	}
	return cfg.DefaultOpenType, nil
}

func (v *Vault) SetDefaultOpenType(openType string) error {
	registry, err := DefaultRegistry()
	if err != nil {
		return err
	}
	return registry.SetDefaultOpenType(openType)
}
