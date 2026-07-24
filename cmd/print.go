package cmd

import (
	"fmt"
	"github.com/andy-neoaira/obs-cli/pkg/actions"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

// includeMentions 控制是否在打印笔记内容后附加显示反向链接（mentions）。
var includeMentions bool

// printCmd 定义了 "print" 子命令，用于在终端中打印笔记的完整内容。
var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Print contents of note",
	Args:  cobra.ExactArgs(1), // 必须提供 1 个参数：笔记名称或路径
	RunE: func(cmd *cobra.Command, args []string) error {
		noteName := args[0]
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		params := actions.PrintParams{
			NoteName:        noteName,
			IncludeMentions: includeMentions,
		}
		contents, err := actions.PrintNote(&vault, &note, params)
		if err != nil {
			return err
		}
		fmt.Println(contents)
		return nil
	},
}

func init() {
	printCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	printCmd.Flags().BoolVarP(&includeMentions, "mentions", "m", false, "include linked mentions at the end")
	rootCmd.AddCommand(printCmd)
}
