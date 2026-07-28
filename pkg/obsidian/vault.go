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

// VaultManager 定义了与 Vault 交互的接口。
// 通过接口解耦，方便在测试中注入 Mock 对象。
type VaultManager interface {
	DefaultName() (string, error)     // 获取当前默认 vault 名称
	SetDefaultName(name string) error // 设置默认 vault 名称
	Path() (string, error)            // 获取当前 vault 的绝对路径
}

// Vault 是 VaultManager 的具体实现。
// Name 为空时，业务层会自动从配置中解析默认名称。
type Vault struct {
	Name string
}
