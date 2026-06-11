package obsidian

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// VaultInfo 保存单个 vault 的名称和路径信息。
type VaultInfo struct {
	Name string `json:"name"` // vault 名称（目录 basename）
	Path string `json:"path"` // vault 的绝对路径
}

// ListVaults 返回所有已注册 Obsidian vault 的列表。
func ListVaults() ([]VaultInfo, error) {
	obsidianConfigFile, err := ObsidianConfigFile()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(obsidianConfigFile)
	if err != nil {
		return nil, errors.New(ObsidianConfigReadError)
	}

	vaultsContent := ObsidianVaultConfig{}
	if json.Unmarshal(content, &vaultsContent) != nil {
		return nil, errors.New(ObsidianConfigParseError)
	}

	vaults := make([]VaultInfo, 0, len(vaultsContent.Vaults))
	for _, element := range vaultsContent.Vaults {
		path := element.Path
		if RunningInWSL() {
			path = adjustForWslMount(path)
		}
		vaults = append(vaults, VaultInfo{
			Name: vaultBaseName(path),
			Path: path,
		})
	}

	return vaults, nil
}

// ResolveVaultName 将用户输入的名称解析为已注册的 vault 名称。
// 如果匹配到多个同名 vault，会返回歧义错误，要求用户先清理重复注册。
func ResolveVaultName(input string) (string, error) {
	vaults, err := ListVaults()
	if err != nil {
		return "", err
	}

	if len(vaults) == 0 {
		return "", errors.New("no vaults registered in Obsidian. Please create a vault in Obsidian first")
	}

	// 先按名称精确匹配
	var nameMatches []VaultInfo
	for _, v := range vaults {
		if v.Name == input {
			nameMatches = append(nameMatches, v)
		}
	}
	if len(nameMatches) == 1 {
		return nameMatches[0].Name, nil
	}
	if len(nameMatches) > 1 {
		var paths []string
		for _, m := range nameMatches {
			paths = append(paths, fmt.Sprintf("  %s", m.Path))
		}
		return "", fmt.Errorf("multiple vaults named %q found. Remove duplicate registrations first:\n%s", input, strings.Join(paths, "\n"))
	}

	// 未找到，返回友好的错误信息并列出所有可用 vault
	var available []string
	for _, v := range vaults {
		available = append(available, fmt.Sprintf("  %s\t(%s)", v.Name, v.Path))
	}

	return "", fmt.Errorf("vault %q not found in Obsidian.\nAvailable vaults:\n%s", input, strings.Join(available, "\n"))
}
