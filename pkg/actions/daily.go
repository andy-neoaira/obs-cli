package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
)

// DailyParams 定义了 daily 命令所需的业务参数。
type DailyParams struct {
	Content   string // 要追加到日记的内容
	UseEditor bool   // 是否使用编辑器打开
	Now       time.Time
}

// DailyNote 是 "daily" 命令的业务核心。
// 流程：读取 Obsidian Daily Notes 插件配置 → 按格式生成今日日期作为文件名 →
// 如有模板则读取模板 → 写入/追加内容 → 打开笔记。
func DailyNote(vault obsidian.VaultManager, uri obsidian.UriManager, params DailyParams) error {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}

	// 读取 Obsidian 的 Daily Notes 插件配置（文件夹、日期格式、模板）
	config := obsidian.ReadDailyNotesConfig(vaultPath)

	// 使用配置的 Moment.js 日期格式生成今日日期字符串
	format := config.Format
	if format == "" {
		format = "YYYY-MM-DD" // 默认格式，与 Obsidian 官方插件一致
	}
	now := params.Now
	if now.IsZero() {
		now = time.Now()
	}
	layout, err := obsidian.ParseMomentToGoFormat(format)
	if err != nil {
		return err
	}
	noteName := now.Format(layout)

	// 如果配置了专门的日记文件夹，将文件名放在该文件夹下
	if config.Folder != "" {
		noteName = config.Folder + "/" + noteName
	}

	// 校验并解析最终路径
	notePath, err := obsidian.ValidateWritablePath(vaultPath, obsidian.AddMdSuffix(noteName))
	if err != nil {
		return err
	}

	// 创建日记所在目录（如果不存在）
	if err := os.MkdirAll(filepath.Dir(notePath), 0755); err != nil {
		return fmt.Errorf("failed to create daily note directory: %w", err)
	}

	// 已存在的 Daily Note 在纯打开场景下不再依赖历史模板。
	_, statErr := os.Stat(notePath)
	fileExists := statErr == nil

	templateContent := ""
	if !fileExists && config.Template != "" {
		// 模板路径来自 vault 内的 .obsidian 配置文件，同样必须经过 ValidatePath。
		// 这样即使配置被误写成 "../../secret"，也不会读取 vault 外部文件。
		templatePath, err := obsidian.ResolveNoteIDPath(vaultPath, config.Template)
		if err != nil {
			return err
		}
		data, readErr := os.ReadFile(templatePath)
		if readErr != nil {
			return fmt.Errorf("failed to read configured daily template: %w", readErr)
		}
		templateContent, _, err = obsidian.RenderDailyTemplate(string(data), now, format, filepath.Base(noteName))
		if err != nil {
			return fmt.Errorf("failed to render configured daily template: %w", err)
		}
	}

	if fileExists && params.Content != "" {
		// 日记已存在且有新内容：追加到末尾
		if err := WriteNoteFile(notePath, params.Content, true, false); err != nil {
			return err
		}
	} else if !fileExists {
		// 日记不存在：创建新文件，内容为模板 + 用户输入
		newContent := templateContent + params.Content
		if err := WriteNoteFile(notePath, newContent, false, false); err != nil {
			return err
		}
	}

	// 打开日记（编辑器或 Obsidian）
	if params.UseEditor {
		return obsidian.OpenInEditor(notePath)
	}

	obsidianUri := uri.Construct(ObsOpenUrl, map[string]string{
		"vault": vaultName,
		"file":  noteName,
	})
	return uri.Execute(obsidianUri)
}
