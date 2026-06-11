package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

// removeVaultCmd 定义了 "remove-vault" 子命令，用于从 Obsidian 配置中注销一个 vault。
// 注意：此操作不会删除磁盘上的任何文件，只是移除了注册信息。
// 如果该 vault 恰好是当前默认 vault，还会自动清除默认设置。
var removeVaultCmd = &cobra.Command{
	Use:   "remove-vault <name>",
	Short: "Unregister a vault",
	Long:  "Removes a vault from the Obsidian config. Does not delete any files on disk.",
	Args:  cobra.ExactArgs(1), // 必须提供 1 个参数：vault 名称
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		name, err := obsidian.RemoveVault(input)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Vault %q removed\n", name)

		// 如果被移除的 vault 是当前默认 vault，则清除默认设置，避免后续命令找不到 vault
		if err := obsidian.ClearDefaultIfMatch(name); err != nil {
			fmt.Fprintln(os.Stderr, "Warning: could not clear default vault:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(removeVaultCmd)
}
