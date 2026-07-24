package cmd

import (
	"github.com/andy-neoaira/obs-cli/pkg/actions"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

// shouldOpen 控制移动后是否自动打开新笔记（多个命令共享此变量）。
var shouldOpen bool

// moveCmd 定义了 "move" 子命令，用于移动或重命名笔记，
// 并自动更新 vault 中其他笔记指向该笔记的链接。
var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move or rename note in vault and updated corresponding links",
	Args:  cobra.ExactArgs(2), // 需要 2 个参数：原路径 和 新路径
	RunE: func(cmd *cobra.Command, args []string) error {
		currentName := args[0]
		newName := args[1]
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		linkRewriter := obsidian.LinkRewriter{}
		uri := obsidian.Uri{}
		params := actions.MoveParams{
			CurrentNoteName: currentName,
			NewNoteName:     newName,
			ShouldOpen:      shouldOpen,
			UseEditor:       resolveUseEditor(cmd, &vault),
		}
		return actions.MoveNote(&vault, &note, &linkRewriter, &uri, params)
	},
}

func init() {
	moveCmd.Flags().BoolVarP(&shouldOpen, "open", "o", false, "open new note")
	moveCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	moveCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian (requires --open flag)")
	rootCmd.AddCommand(moveCmd)
}
