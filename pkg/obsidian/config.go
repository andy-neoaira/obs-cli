package obsidian

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DailyNotesConfig 对应 vault 中 .obsidian/daily-notes.json 的结构。
// 保存 Daily Notes 插件的文件夹、日期格式和模板配置。
type DailyNotesConfig struct {
	Folder   string `json:"folder"`   // 日记存放文件夹
	Format   string `json:"format"`   // 日期格式（Moment.js 风格）
	Template string `json:"template"` // 模板笔记路径
}

type ConfigFileError struct {
	File string
	Kind string
	Err  error
}

func (e *ConfigFileError) Error() string {
	return fmt.Sprintf("%s Obsidian config %s: %v", e.Kind, e.File, e.Err)
}

func (e *ConfigFileError) Unwrap() error {
	return e.Err
}

func LoadDailyNotesConfig(vaultPath string) (DailyNotesConfig, bool, error) {
	var config DailyNotesConfig
	found, err := loadObsidianConfig(vaultPath, "daily-notes.json", &config)
	return config, found, err
}

func loadObsidianConfig(vaultPath, name string, target any) (bool, error) {
	relative := filepath.ToSlash(filepath.Join(".obsidian", name))
	data, err := os.ReadFile(filepath.Join(vaultPath, filepath.FromSlash(relative)))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, &ConfigFileError{File: relative, Kind: "read", Err: err}
	}
	if err := json.Unmarshal(data, target); err != nil {
		return true, &ConfigFileError{File: relative, Kind: "parse", Err: err}
	}
	return true, nil
}
