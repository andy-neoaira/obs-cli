package obsidian

import (
	"encoding/json"
	"errors"
	"github.com/andy-neoaira/obs-cli/pkg/config"
	"path"
	"strings"
)

// ObsidianConfigFile 和 RunningInWSL 是包级变量，分别指向 config 包中的对应函数。
// 使用变量而非直接调用函数，方便在测试中替换为 Mock 实现。
var ObsidianConfigFile = config.ObsidianFile
var RunningInWSL = config.RunningInWSL

// Path 返回当前 vault 的绝对路径。
// Vault.Name 必须是已注册 vault 的名称；运行命令不再接受路径作为隐式 vault 标识。
func (v *Vault) Path() (string, error) {
	registry, err := DefaultRegistry()
	if err != nil {
		return "", err
	}
	var record config.VaultRecord
	if v.Name == "" {
		record, err = registry.Default()
	} else {
		record, err = registry.Get(v.Name)
	}
	if err != nil {
		return "", err
	}
	return record.Path, nil
}

// adjustForWslMount 将 Windows 绝对路径转换为 WSL 的 /mnt/ 挂载路径。
// 例如 "C:\Users\name" 会转换为 "/mnt/c/Users/name"。
func adjustForWslMount(dir string) string {
	// 检测 Windows 盘符模式（如 C:, D:, E:）
	if len(dir) >= 2 && dir[1] == ':' && ((dir[0] >= 'A' && dir[0] <= 'Z') || (dir[0] >= 'a' && dir[0] <= 'z')) {
		driveLetter := strings.ToLower(string(dir[0]))
		mnted := "/mnt/" + driveLetter + dir[2:]
		return strings.ReplaceAll(mnted, "\\", "/")
	}

	return dir
}

// vaultBaseName 返回 vault 路径最后一段，兼容 Obsidian 配置中的 Unix 和 Windows 路径。
func vaultBaseName(vaultPath string) string {
	normalized := strings.ReplaceAll(vaultPath, "\\", "/")
	return path.Base(normalized)
}

// getPathForVault 从 Obsidian 配置内容中查找指定名称的 vault 路径。
func getPathForVault(content []byte, name string) (string, error) {
	vaultsContent := ObsidianVaultConfig{}
	if json.Unmarshal(content, &vaultsContent) != nil {
		return "", errors.New(ObsidianConfigParseError)
	}

	for _, element := range vaultsContent.Vaults {
		if vaultBaseName(element.Path) == name {
			return element.Path, nil
		}
	}

	return "", errors.New(ObsidianConfigVaultNotFoundError)
}
