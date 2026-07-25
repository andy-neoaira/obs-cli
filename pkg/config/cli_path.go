package config

import (
	"errors"
	"os"
	"path/filepath"
)

// UserConfigDirectory 返回 CLI 配置根目录。
//
// OBS_CLI_CONFIG_HOME 只覆盖 obs-cli 自有配置的根目录，主要用于隔离测试、
// 沙箱和多实例运行；未设置时仍遵循 os.UserConfigDir。使用变量而非直接
// 调用，方便测试替换为 Mock 实现。
var UserConfigDirectory = func() (string, error) {
	if override := os.Getenv("OBS_CLI_CONFIG_HOME"); override != "" {
		if !filepath.IsAbs(override) {
			return "", errors.New("OBS_CLI_CONFIG_HOME must be an absolute path")
		}
		return filepath.Clean(override), nil
	}
	return os.UserConfigDir()
}

// CliPath 返回 CLI 自身配置目录和配置文件的完整路径。
// 配置目录遵循操作系统标准：
//   - Linux/Mac: ~/.config/obs-cli/
//   - Windows: %APPDATA%\obs-cli\
func CliPath() (cliConfigDir string, cliConfigFile string, err error) {
	userConfigDir, err := UserConfigDirectory()
	if err != nil {
		return "", "", errors.New(UserConfigDirectoryNotFoundErrorMessage)
	}
	cliConfigDir = filepath.Join(userConfigDir, ObsCLIConfigDirectory)
	cliConfigFile = filepath.Join(cliConfigDir, ObsCLIConfigFile)
	return cliConfigDir, cliConfigFile, nil
}

// V2Path 返回 Agent-first V2 自有配置路径。
// 该文件与 Obsidian 官方 obsidian.json 完全分离。
func V2Path() (cliConfigDir string, cliConfigFile string, err error) {
	userConfigDir, err := UserConfigDirectory()
	if err != nil {
		return "", "", errors.New(UserConfigDirectoryNotFoundErrorMessage)
	}
	cliConfigDir = filepath.Join(userConfigDir, ObsCLIConfigDirectory)
	cliConfigFile = filepath.Join(cliConfigDir, ObsCLIV2ConfigFile)
	return cliConfigDir, cliConfigFile, nil
}
