package cmd

import (
	"fmt"
	"log"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

// removeVaultCmd 定义了 "remove-vault" 子命令，用于从 Obsidian 配置中注销一个 vault。
// 注意：此操作不会删除磁盘上的任何文件，只是移除了注册信息。
// 如果该 vault 恰好是当前默认 vault，还会自动清除默认设置。
var removeVaultCmd = &cobra.Command{
	Use:   "remove-vault <name>",
	Short: "Unregister a vault",
	Long:  "Removes a vault from obs-cli V2. Does not modify Obsidian config or delete files.",
	Args:  cobra.ExactArgs(1), // 必须提供 1 个参数：vault 名称
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]

		registry, err := obsidian.DefaultRegistry()
		if err != nil {
			log.Fatal(err)
		}
		vault, err := registry.Remove(input)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Vault %q removed\n", vault.Name)
	},
}

func init() {
	rootCmd.AddCommand(removeVaultCmd)
}
