package obsidian

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ObsidianAppConfig 对应 vault 中 .obsidian/app.json 的结构。
// 主要读取新建笔记的默认位置和用户配置的排除规则。
type ObsidianAppConfig struct {
	NewFileLocation   string   `json:"newFileLocation"`   // 新笔记存放方式：root 或 folder
	NewFileFolderPath string   `json:"newFileFolderPath"` // 当方式为 folder 时的目标文件夹
	UserIgnoreFilters []string `json:"userIgnoreFilters"` // 用户配置的排除路径/通配符列表
}

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

func LoadAppConfig(vaultPath string) (ObsidianAppConfig, bool, error) {
	var config ObsidianAppConfig
	found, err := loadObsidianConfig(vaultPath, "app.json", &config)
	return config, found, err
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

// ExcludedPaths 读取 vault 中 .obsidian/app.json 的 userIgnoreFilters，
// 返回需要排除的路径模式列表。如果配置文件不存在或读取失败，返回 nil。
func ExcludedPaths(vaultPath string) []string {
	config, _, err := LoadAppConfig(vaultPath)
	if err != nil {
		return nil
	}
	return config.UserIgnoreFilters
}

// DefaultNoteFolder 读取 vault 中配置的新笔记默认存放文件夹。
// 如果未配置或读取失败，返回空字符串（调用方应回退到 vault 根目录）。
func DefaultNoteFolder(vaultPath string) string {
	config, _, err := LoadAppConfig(vaultPath)
	if err != nil {
		return ""
	}
	if config.NewFileLocation == "folder" && config.NewFileFolderPath != "" {
		return config.NewFileFolderPath
	}

	return ""
}

// ReadDailyNotesConfig 读取 vault 中 Daily Notes 插件的配置。
// 如果配置文件不存在或读取失败，返回零值结构体。
func ReadDailyNotesConfig(vaultPath string) DailyNotesConfig {
	config, _, err := LoadDailyNotesConfig(vaultPath)
	if err != nil {
		return DailyNotesConfig{}
	}
	return config
}

// ApplyDefaultFolder 在 noteName 没有显式路径（不含 "/"）时，
// 自动在前面加上 Obsidian 配置的新笔记默认文件夹。
// 如果 noteName 已包含 "/" 或未配置默认文件夹，则原样返回。
func ApplyDefaultFolder(noteName, vaultPath string) string {
	if strings.Contains(noteName, "/") {
		return noteName
	}
	if folder := DefaultNoteFolder(vaultPath); folder != "" {
		return folder + "/" + noteName
	}
	return noteName
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
