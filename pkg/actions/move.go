package actions

import (
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
)

// MoveParams 定义了 move 命令所需的业务参数。
type MoveParams struct {
	CurrentNoteName string // 原笔记路径
	NewNoteName     string // 新笔记路径
	ShouldOpen      bool   // 移动后是否自动打开
	UseEditor       bool   // 打开时是否使用编辑器
}

// MoveNote 是 "move" 命令的业务核心。
// 流程：校验原路径和新路径 → 移动文件 → 更新 vault 中所有指向该笔记的链接 →（可选）打开新笔记。
//
// note 只负责文件移动，linkRewriter 只负责内容中的链接重写。
// 这两个依赖分开传入，是为了避免 NoteManager 继续承担链接解析策略，保持职责边界清晰。
func MoveNote(vault obsidian.VaultManager, note obsidian.NoteManager, linkRewriter obsidian.LinkRewriteManager, uri obsidian.UriManager, params MoveParams) error {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}
	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}

	// 分别校验原路径和新路径，确保都在 vault 目录内
	currentPath, err := obsidian.ValidateWritablePath(vaultPath, obsidian.AddMdSuffix(params.CurrentNoteName))
	if err != nil {
		return err
	}
	newPath, err := obsidian.ValidateWritablePath(vaultPath, obsidian.AddMdSuffix(params.NewNoteName))
	if err != nil {
		return err
	}

	// 执行文件移动/重命名
	err = note.Move(currentPath, newPath)
	if err != nil {
		return err
	}

	// 遍历 vault 中所有笔记，将旧链接替换为新链接，保持笔记间引用关系
	err = linkRewriter.UpdateLinks(vaultPath, params.CurrentNoteName, params.NewNoteName)
	if err != nil {
		return err
	}

	// 如果用户要求打开新笔记
	if params.ShouldOpen {
		if params.UseEditor {
			filePathWithExt, err := obsidian.ValidatePath(vaultPath, obsidian.AddMdSuffix(params.NewNoteName))
			if err != nil {
				return err
			}
			return obsidian.OpenInEditor(filePathWithExt)
		}

		obsidianUri := uri.Construct(ObsOpenUrl, map[string]string{
			"file":  params.NewNoteName,
			"vault": vaultName,
		})

		err := uri.Execute(obsidianUri)
		if err != nil {
			return err
		}
	}

	return nil
}
