package obsidian

import (
	"encoding/json"
	"errors"
	"github.com/andy-neoaira/obs-cli/pkg/config"
	"os"
)

// CliConfigPath 和 JsonMarshal 是包级变量，指向 config 包和 json 包的函数。
// 使用变量方便测试中替换为 Mock 实现。
var CliConfigPath = config.CliPath
var JsonMarshal = json.Marshal

// DefaultName 返回当前 Vault 的名称。
// 如果 Vault.Name 已设置，直接返回；否则从 CLI 配置文件中读取默认 vault 名称。
func (v *Vault) DefaultName() (string, error) {
	if v.Name != "" {
		return v.Name, nil
	}

	// 获取 CLI 配置文件路径
	_, cliConfigFile, err := CliConfigPath()
	if err != nil {
		return "", err
	}

	// 读取配置文件
	content, err := os.ReadFile(cliConfigFile)
	if err != nil {
		return "", errors.New(ObsidianCLIConfigReadError)
	}

	// 解析 JSON
	cliConfig := CliConfig{}
	err = json.Unmarshal(content, &cliConfig)

	if err != nil {
		return "", errors.New(ObsidianCLIConfigParseError)
	}

	if cliConfig.DefaultVaultName == "" {
		return "", errors.New(ObsidianCLIConfigParseError)
	}

	// 缓存到 Vault 对象中，避免后续重复读取文件
	v.Name = cliConfig.DefaultVaultName
	return cliConfig.DefaultVaultName, nil
}

// SetDefaultName 设置 CLI 配置中的默认 vault 名称。
// 会保留配置文件中已有的其他字段（如 DefaultOpenType）。
func (v *Vault) SetDefaultName(name string) error {
	obsConfigDir, obsConfigFile, err := CliConfigPath()
	if err != nil {
		return err
	}

	// 先读取已有配置，保留其他字段
	cliConfig := CliConfig{}
	if content, readErr := os.ReadFile(obsConfigFile); readErr == nil {
		if err := json.Unmarshal(content, &cliConfig); err != nil {
			return errors.New(ObsidianCLIConfigParseError)
		}
	}

	cliConfig.DefaultVaultName = name

	// 序列化为 JSON
	jsonContent, err := JsonMarshal(cliConfig)
	if err != nil {
		return errors.New(ObsidianCLIConfigGenerateJSONError)
	}

	// 确保配置目录存在
	err = os.MkdirAll(obsConfigDir, os.ModePerm)
	if err != nil {
		return errors.New(ObsidianCLIConfigDirWriteEror)
	}

	// 写入配置文件
	err = os.WriteFile(obsConfigFile, jsonContent, 0644)
	if err != nil {
		return errors.New(ObsidianCLIConfigWriteError)
	}

	v.Name = name

	return nil
}

// DefaultOpenType 返回 CLI 配置中的默认打开方式。
// 配置文件不存在时默认使用 "obsidian"；配置存在但内容损坏时返回错误。
func (v *Vault) DefaultOpenType() (string, error) {
	_, cliConfigFile, err := CliConfigPath()
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(cliConfigFile)
	if err != nil {
		return "obsidian", nil //nolint:nilerr // 配置文件不存在时使用默认值
	}

	cliConfig := CliConfig{}
	if err := json.Unmarshal(content, &cliConfig); err != nil {
		return "", errors.New(ObsidianCLIConfigParseError)
	}

	if cliConfig.DefaultOpenType == "" {
		return "obsidian", nil
	}

	return cliConfig.DefaultOpenType, nil
}

// SetDefaultOpenType 设置 CLI 配置中的默认打开方式（obsidian 或 editor）。
// 同样会保留配置文件中已有的其他字段。
func (v *Vault) SetDefaultOpenType(openType string) error {
	obsConfigDir, obsConfigFile, err := CliConfigPath()
	if err != nil {
		return err
	}

	// 先读取已有配置，保留其他字段
	cliConfig := CliConfig{}
	if content, readErr := os.ReadFile(obsConfigFile); readErr == nil {
		if err := json.Unmarshal(content, &cliConfig); err != nil {
			return errors.New(ObsidianCLIConfigParseError)
		}
	}

	cliConfig.DefaultOpenType = openType

	jsonContent, err := JsonMarshal(cliConfig)
	if err != nil {
		return errors.New(ObsidianCLIConfigGenerateJSONError)
	}

	if err := os.MkdirAll(obsConfigDir, os.ModePerm); err != nil {
		return errors.New(ObsidianCLIConfigDirWriteEror)
	}

	if err := os.WriteFile(obsConfigFile, jsonContent, 0644); err != nil {
		return errors.New(ObsidianCLIConfigWriteError)
	}

	return nil
}
