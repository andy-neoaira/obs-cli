package cmd

import (
	"fmt"
	"log"

	"github.com/andy-neoaira/obs-cli/pkg/actions"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

// listCmd 定义了 "list" 子命令，用于列出笔记库中的文件和文件夹。
var listCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List files and folders in vault",
	Args:  cobra.MaximumNArgs(1), // 最多接收 1 个参数：可选的目标路径
	Run: func(cmd *cobra.Command, args []string) {
		// 如果用户没有提供路径参数，则默认列出 vault 根目录
		var targetPath string
		if len(args) > 0 {
			targetPath = args[0]
		}

		vault := obsidian.Vault{Name: vaultName}
		entries, err := actions.ListEntries(&vault, actions.ListParams{Path: targetPath})
		if err != nil {
			log.Fatal(err)
		}

		// 遍历并打印每个条目，前面加上圆点符号美化输出
		for _, entry := range entries {
			fmt.Printf("• %s\n", entry)
		}
	},
}

func init() {
	listCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	rootCmd.AddCommand(listCmd)
}
