package actions

import (
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
)

// DeleteParams 定义了 delete 命令所需的业务参数。
type DeleteParams struct {
	NotePath string // 要删除的笔记路径
}

// DeleteNote 是 "delete" 命令的业务核心。
// 先校验路径是否在 vault 内，然后调用笔记管理器执行删除。
func DeleteNote(vault obsidian.VaultManager, note obsidian.NoteManager, params DeleteParams) error {
	_, err := vault.DefaultName()
	if err != nil {
		return err
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}

	// 校验目标路径是否在 vault 目录内部，防止误删或恶意路径
	notePath, err := obsidian.ValidateWritablePath(vaultPath, obsidian.AddMdSuffix(params.NotePath))
	if err != nil {
		return err
	}

	err = note.Delete(notePath)
	if err != nil {
		return err
	}
	return nil
}
