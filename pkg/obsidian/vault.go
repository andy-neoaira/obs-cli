package obsidian

// CliConfig 是 CLI 自身配置文件的 JSON 结构。
// 保存在 ~/.config/obs-cli/preferences.json 中。
type CliConfig struct {
	DefaultVaultName string `json:"default_vault_name"`          // 默认 vault 名称
	DefaultOpenType  string `json:"default_open_type,omitempty"` // 默认打开方式：obsidian 或 editor
}

type ObsidianVaultEntry struct {
	Path string `json:"path"`
	Open bool   `json:"open,omitempty"`
}

// ObsidianVaultConfig 是 Obsidian 官方配置文件中 vault 注册表的只读结构。
// V2 只允许发现和导入，禁止通过该结构写入 obsidian.json。
type ObsidianVaultConfig struct {
	Vaults map[string]ObsidianVaultEntry `json:"vaults"`
}

// VaultManager 定义了与 Vault 交互的接口。
// 通过接口解耦，方便在测试中注入 Mock 对象。
type VaultManager interface {
	DefaultName() (string, error)     // 获取当前默认 vault 名称
	SetDefaultName(name string) error // 设置默认 vault 名称
	Path() (string, error)            // 获取当前 vault 的绝对路径
	DefaultOpenType() (string, error) // 获取默认打开方式
}

// Vault 是 VaultManager 的具体实现。
// Name 为空时，业务层会自动从配置中解析默认名称。
type Vault struct {
	Name string
}
