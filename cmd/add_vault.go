package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

// addVaultCmd 定义了 "add-vault" 子命令，用于将本地目录注册为 Obsidian 笔记库。
// 如果目录不存在或不是文件夹，会返回错误。
// 通过 --set-default  flag 可以同时将新添加的 vault 设为默认。
var addVaultCmd = &cobra.Command{
	Use:   "add-vault <path>",
	Short: "Register a vault directory",
	Long:  "Registers a directory as an Obsidian vault. Creates the Obsidian config file if it does not exist.",
	Args:  cobra.ExactArgs(1), // 必须提供 1 个参数：vault 的本地路径
	Run: func(cmd *cobra.Command, args []string) {
		absPath, err := obsidian.AddVault(args[0])
		if err != nil {
			log.Fatal(err)
		}

		// 使用路径的最后一个部分作为 vault 名称
		name := filepath.Base(absPath)
		fmt.Printf("Vault %q registered at: %s\n", name, absPath)

		// 如果用户传了 --set-default，则同时设为默认 vault
		setDefault, _ := cmd.Flags().GetBool("set-default")
		if setDefault {
			v := obsidian.Vault{Name: name}
			if err := v.SetDefaultName(name); err != nil {
				log.Fatal(err)
			}
			fmt.Println("Default vault set to:", name)
		}
	},
}

func init() {
	addVaultCmd.Flags().Bool("set-default", false, "set the added vault as the default")
	rootCmd.AddCommand(addVaultCmd)
}
