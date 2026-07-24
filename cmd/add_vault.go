package cmd

import (
	"fmt"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

// addVaultCmd 定义了 "add-vault" 子命令，用于将本地目录注册为 Obsidian 笔记库。
// 如果目录不存在或不是文件夹，会返回错误。
// 通过 --set-default  flag 可以同时将新添加的 vault 设为默认。
var addVaultCmd = &cobra.Command{
	Use:   "add-vault <path>",
	Short: "Register a vault directory",
	Long:  "Registers a directory in obs-cli V2. It never modifies Obsidian's obsidian.json.",
	Args:  cobra.ExactArgs(1), // 必须提供 1 个参数：vault 的本地路径
	RunE: func(cmd *cobra.Command, args []string) error {
		registry, err := obsidian.DefaultRegistry()
		if err != nil {
			return err
		}
		vault, err := registry.Add(args[0], "")
		if err != nil {
			return err
		}
		fmt.Printf("Vault %q registered at: %s\n", vault.Name, vault.Path)

		// 如果用户传了 --set-default，则同时设为默认 vault
		setDefault, _ := cmd.Flags().GetBool("set-default")
		if setDefault {
			if _, err := registry.SetDefault(vault.ID); err != nil {
				return err
			}
			fmt.Println("Default vault set to:", vault.Name)
		}
		return nil
	},
}

func init() {
	addVaultCmd.Flags().Bool("set-default", false, "set the added vault as the default")
	rootCmd.AddCommand(addVaultCmd)
}
