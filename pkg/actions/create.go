package actions

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
)

// CreateParams 定义了 create 命令所需的业务参数。
type CreateParams struct {
	NoteName        string // 笔记名称或路径
	ShouldAppend    bool   // 是否在已有笔记末尾追加内容
	ShouldOverwrite bool   // 是否覆盖已有笔记
	Content         string // 要写入的文本内容
	ShouldOpen      bool   // 创建后是否自动打开
	UseEditor       bool   // 打开时是否使用编辑器
}

// CreateNote 是 "create" 命令的业务核心。
// 流程：读取默认 vault → 应用默认文件夹 → 校验路径 → 创建目录 → 写入文件 →（可选）打开笔记。
func CreateNote(vault obsidian.VaultManager, uri obsidian.UriManager, params CreateParams) error {
	// 从 CLI 配置中读取默认 vault 名称（Path() 之前必须先有 DefaultName）
	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}

	// 如果笔记名没有显式路径（不含 "/"），自动加上 Obsidian 配置中的默认新建文件夹
	params.NoteName = obsidian.ApplyDefaultFolder(params.NoteName, vaultPath)

	// 校验最终路径是否仍在 vault 目录内部，防止路径遍历攻击
	notePath, err := obsidian.ValidatePath(vaultPath, obsidian.AddMdSuffix(params.NoteName))
	if err != nil {
		return err
	}

	// 如果笔记位于子目录中，自动创建所需的中间目录（类似 mkdir -p）
	if err := os.MkdirAll(filepath.Dir(notePath), 0755); err != nil {
		return fmt.Errorf("failed to create note directory: %w", err)
	}

	if err := WriteNoteFile(notePath, params.Content, params.ShouldAppend, params.ShouldOverwrite); err != nil {
		return err
	}

	// 如果用户没有要求打开，到此结束
	if !params.ShouldOpen {
		return nil
	}

	// 根据 UseEditor 决定用编辑器还是 Obsidian URI 打开
	if params.UseEditor {
		return obsidian.OpenInEditor(notePath)
	}

	obsidianUri := uri.Construct(ObsOpenUrl, map[string]string{
		"vault": vaultName,
		"file":  params.NoteName,
	})
	return uri.Execute(obsidianUri)
}

// WriteNoteFile 向 notePath 写入内容，根据 shouldAppend 和 shouldOverwrite 决定行为：
//   - 文件不存在：直接创建并写入
//   - 文件存在且 shouldAppend=true：追加内容
//   - 文件存在且 shouldOverwrite=true：覆盖内容
//   - 文件存在且两者都为 false：返回错误，要求用户明确选择 append 或 overwrite
func WriteNoteFile(notePath, content string, shouldAppend, shouldOverwrite bool) error {
	_, err := os.Stat(notePath)
	fileExists := err == nil

	if fileExists && shouldAppend {
		f, err := os.OpenFile(notePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open note for appending: %w", err)
		}
		if _, err = f.WriteString(content); err != nil {
			f.Close()
			return err
		}
		return f.Close()
	}

	if fileExists && !shouldOverwrite {
		return fmt.Errorf("note already exists: use --append or --overwrite: %s", notePath)
	}

	return os.WriteFile(notePath, []byte(content), 0644)
}
