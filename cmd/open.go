package cmd

import (
	"github.com/andy-neoaira/obs-cli/pkg/actions"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

// 命令行参数变量，用于接收 --vault、--section、--editor 等 flag 的值。
var vaultName string
var sectionName string

// OpenVaultCmd 定义了 "open" 子命令，用于在 Obsidian 或编辑器中打开指定笔记。
var OpenVaultCmd = &cobra.Command{
	Use:   "open",
	Short: "Opens note in vault by note name",
	Args:  cobra.ExactArgs(1), // 严格要求恰好 1 个参数：笔记名称
	RunE: func(cmd *cobra.Command, args []string) error {
		// 构造 Vault 对象（Name 为空时业务层会从配置中读取默认值）
		vault := obsidian.Vault{Name: vaultName}
		uri := obsidian.Uri{}
		noteName := args[0]

		// 组装参数并调用业务层函数 OpenNote
		params := actions.OpenParams{NoteName: noteName, Section: sectionName, UseEditor: resolveUseEditor(cmd, &vault)}
		return actions.OpenNote(&vault, &uri, params)
	},
}

// init 函数在包导入时自动执行，用于注册该子命令的 flag 并将其挂到根命令下。
func init() {
	OpenVaultCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	OpenVaultCmd.Flags().StringVarP(&sectionName, "section", "s", "", "heading text to open within the note (case-sensitive)")
	OpenVaultCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian")
	rootCmd.AddCommand(OpenVaultCmd)
}
