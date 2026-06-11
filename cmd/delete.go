package cmd

import (
	"github.com/andy-neoaira/obs-cli/pkg/actions"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"log"

	"github.com/spf13/cobra"
)

// deleteCmd 定义了 "delete" 子命令，用于删除笔记库中的指定笔记。
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete note in vault",
	Args:  cobra.ExactArgs(1), // 必须提供 1 个参数：要删除的笔记路径
	Run: func(cmd *cobra.Command, args []string) {
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		notePath := args[0]
		params := actions.DeleteParams{NotePath: notePath}
		err := actions.DeleteNote(&vault, &note, params)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	rootCmd.AddCommand(deleteCmd)
}
