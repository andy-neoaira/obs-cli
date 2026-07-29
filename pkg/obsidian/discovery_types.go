package obsidian

type ObsidianVaultEntry struct {
	Path string `json:"path"`
	Open bool   `json:"open,omitempty"`
}

// ObsidianVaultConfig 是 Obsidian 官方配置文件中 vault 注册表的只读结构。
// CLI 只允许发现，禁止通过该结构写入 obsidian.json。
type ObsidianVaultConfig struct {
	Vaults map[string]ObsidianVaultEntry `json:"vaults"`
}
