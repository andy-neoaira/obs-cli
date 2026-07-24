package config

import (
	"errors"
	"os"
	"path/filepath"
)

// UserConfigDirectory 是一个包级变量，指向 os.UserConfigDir 函数。
// 使用变量而非直接调用，方便在测试中替换为 Mock 实现。
var UserConfigDirectory = os.UserConfigDir

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
