package obsidian

// DefaultName 返回 registry 中的默认 Vault 名称。
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

// SetDefaultName 通过 ID 或名称设置 registry 中的默认 Vault。
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
