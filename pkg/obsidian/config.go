package obsidian

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// MomentToGoFormat 将 Moment.js 日期格式字符串转换为 Go 的 time.Layout 格式。
// 采用两遍替换策略：先用唯一占位符替换 Moment 标记，避免级联替换错误
// （例如将 "January" 中的 "a" 错误替换为 "pm"）。
//
// 注意：Moment.js 的 "dd" 标记（两位星期缩写，如 "Mo"）在 Go 中没有对应物，不支持。
func MomentToGoFormat(momentFmt string) string {
	if result, err := ParseMomentToGoFormat(momentFmt); err == nil {
		return result
	}

	// 顺序很重要：必须先替换长标记，再替换短标记，否则会发生冲突
	replacements := []struct {
		moment string
		goFmt  string
	}{
		{"YYYY", "2006"},
		{"YY", "06"},
		{"MMMM", "January"},
		{"MMM", "Jan"},
		{"MM", "01"},
		{"M", "1"},
		{"DD", "02"},
		{"D", "2"},
		{"dddd", "Monday"},
		{"ddd", "Mon"},
		{"HH", "15"},
		{"hh", "03"},
		{"h", "3"},
		{"mm", "04"},
		{"ss", "05"},
		{"A", "PM"},
		{"a", "pm"},
	}

	// 第一遍：将所有 Moment 标记替换为唯一占位符（使用不可打印字符避免冲突）
	result := momentFmt
	for i, r := range replacements {
		placeholder := fmt.Sprintf("\x00%d\x00", i)
		result = strings.ReplaceAll(result, r.moment, placeholder)
	}

	// 第二遍：将占位符替换为 Go 格式字符串
	for i, r := range replacements {
		placeholder := fmt.Sprintf("\x00%d\x00", i)
		result = strings.ReplaceAll(result, placeholder, r.goFmt)
	}

	return result
}
