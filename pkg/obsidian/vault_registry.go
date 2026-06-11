package obsidian

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andy-neoaira/obs-cli/pkg/config"
)

// AddVault 将本地目录注册为 Obsidian vault。
// 如果 Obsidian 的配置文件不存在，会自动创建。
// 返回解析后的绝对路径。
func AddVault(vaultPath string) (string, error) {
	// 将输入路径解析为绝对路径，消除 . 和 .. 等相对路径元素
	absPath, err := filepath.Abs(vaultPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// 校验路径存在且是目录
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("path does not exist: %s", absPath)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", absPath)
	}

	// 读取或创建 Obsidian 的 vault 注册表配置
	obsidianConfigFile, vaultsConfig, err := readOrCreateObsidianConfig()
	if err != nil {
		return "", err
	}

	// 检查该路径是否已被注册，避免重复
	for _, v := range vaultsConfig.Vaults {
		if filepath.Clean(v.Path) == filepath.Clean(absPath) {
			return "", fmt.Errorf("vault already registered: %s", absPath)
		}
	}

	// 生成唯一的随机 vault ID（Obsidian 官方也使用这种 32 位 hex ID）
	id, err := generateVaultID()
	if err != nil {
		return "", fmt.Errorf("failed to generate vault ID: %w", err)
	}

	vaultsConfig.Vaults[id] = struct {
		Path string `json:"path"`
	}{Path: absPath}

	return absPath, writeObsidianConfig(obsidianConfigFile, vaultsConfig)
}

// RemoveVault 从 Obsidian 配置中注销一个 vault。
// 只支持通过 vault 名称指定；如果匹配到多个 vault，会报错提示用户先清理重复注册。
// 返回被移除 vault 的名称（目录 basename），方便调用方做后续清理（如清除默认设置）。
func RemoveVault(input string) (string, error) {
	obsidianConfigFile, err := ObsidianConfigFile()
	if err != nil {
		return "", errors.New(ObsidianConfigReadError)
	}

	content, err := os.ReadFile(obsidianConfigFile)
	if err != nil {
		return "", errors.New(ObsidianConfigReadError)
	}

	vaultsConfig := ObsidianVaultConfig{}
	if json.Unmarshal(content, &vaultsConfig) != nil {
		return "", errors.New(ObsidianConfigParseError)
	}

	// 按名称（目录 basename）匹配，收集所有匹配项以检测歧义
	type match struct {
		id   string
		path string
	}
	var matches []match
	for id, v := range vaultsConfig.Vaults {
		if vaultBaseName(v.Path) == input {
			matches = append(matches, match{id: id, path: v.Path})
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("vault %q not found", input)
	}
	if len(matches) > 1 {
		var paths []string
		for _, m := range matches {
			paths = append(paths, fmt.Sprintf("  %s", m.path))
		}
		return "", fmt.Errorf("multiple vaults named %q found. Remove duplicate registrations first:\n%s", input, strings.Join(paths, "\n"))
	}

	delete(vaultsConfig.Vaults, matches[0].id)
	return input, writeObsidianConfig(obsidianConfigFile, vaultsConfig)
}

// ClearDefaultIfMatch 如果被移除的 vault 名称恰好是当前默认 vault，则清除默认设置。
func ClearDefaultIfMatch(name string) error {
	_, cliConfigFile, err := CliConfigPath()
	if err != nil {
		return nil //nolint:nilerr // 没有配置目录意味着无需清理
	}

	content, err := os.ReadFile(cliConfigFile)
	if err != nil {
		return nil //nolint:nilerr // 没有配置文件意味着无需清理
	}

	cliConfig := CliConfig{}
	if err := json.Unmarshal(content, &cliConfig); err != nil {
		return nil //nolint:nilerr // 配置解析失败也意味着无需清理
	}

	// 只有匹配时才清除
	if cliConfig.DefaultVaultName != name {
		return nil
	}

	v := &Vault{}
	return v.SetDefaultName("")
}

// readOrCreateObsidianConfig 读取 Obsidian 的 vault 注册表配置。
// 如果配置文件不存在，则创建新的空配置并返回。
func readOrCreateObsidianConfig() (string, ObsidianVaultConfig, error) {
	empty := ObsidianVaultConfig{
		Vaults: make(map[string]struct {
			Path string `json:"path"`
		}),
	}

	// 尝试读取已有配置
	obsidianConfigFile, err := ObsidianConfigFile()
	if err == nil {
		content, readErr := os.ReadFile(obsidianConfigFile)
		if readErr == nil {
			vaultsConfig := ObsidianVaultConfig{}
			if err := json.Unmarshal(content, &vaultsConfig); err != nil {
				return "", empty, fmt.Errorf("corrupt obsidian config at %s: %w", obsidianConfigFile, err)
			}
			if vaultsConfig.Vaults == nil {
				vaultsConfig.Vaults = make(map[string]struct {
					Path string `json:"path"`
				})
			}
			return obsidianConfigFile, vaultsConfig, nil
		}
	}

	// 配置不存在：创建新的空配置
	userConfigDir, err := config.UserConfigDirectory()
	if err != nil {
		return "", empty, fmt.Errorf("failed to determine config directory: %w", err)
	}

	configDir := filepath.Join(userConfigDir, config.ObsidianConfigDirectory)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return "", empty, fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := filepath.Join(configDir, config.ObsidianConfigFile)
	return configFile, empty, nil
}

// writeObsidianConfig 将 vault 注册表配置序列化为 JSON 并写入文件。
func writeObsidianConfig(path string, cfg ObsidianVaultConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// generateVaultID 生成 16 字节随机数并编码为 32 位 hex 字符串，作为 vault 的唯一 ID。
func generateVaultID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
