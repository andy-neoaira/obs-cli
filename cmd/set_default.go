package cmd

import (
	"fmt"
	"log"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

// runSetDefaultVault 是 set-default-vault 命令的执行逻辑。
// 它可以同时设置默认 vault 名称和默认打开方式（obsidian 或 editor）。
func runSetDefaultVault(cmd *cobra.Command, args []string) {
	openType, err := cmd.Flags().GetString("open-type")
	if err != nil {
		log.Fatalf("Failed to parse --open-type flag: %v", err)
	}

	// 参数校验：至少要提供 vault 名称或 --open-type 其中之一
	if len(args) == 0 && openType == "" {
		log.Fatal("Please provide a vault name or use --open-type to set the default open type")
	}

	// 设置默认 vault 名称
	if len(args) > 0 {
		registry, err := obsidian.DefaultRegistry()
		if err != nil {
			log.Fatal(err)
		}
		vault, err := registry.SetDefault(args[0])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Default vault set to:", vault.Name)
		fmt.Println("Default vault path set to:", vault.Path)
	}

	// 设置默认打开方式
	if openType != "" {
		if openType != "obsidian" && openType != "editor" {
			log.Fatalf("Invalid open type %q: must be 'obsidian' or 'editor'", openType)
		}
		registry, err := obsidian.DefaultRegistry()
		if err != nil {
			log.Fatal(err)
		}
		if err := registry.SetDefaultOpenType(openType); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Default open type set to:", openType)
	}
}

// setDefaultVaultCmd 是正式的 "set-default-vault" 命令。
var setDefaultVaultCmd = &cobra.Command{
	Use:   "set-default-vault",
	Short: "Sets default vault and/or open type",
	Args:  cobra.RangeArgs(0, 1), // 接收 0 或 1 个参数
	Run:   runSetDefaultVault,
}

// init 在包导入时自动执行，用于注册 set-default-vault 命令的 flag 并将其挂到根命令下。
func init() {
	setDefaultVaultCmd.Flags().String("open-type", "", "default open type: 'obsidian' (default) or 'editor'")
	rootCmd.AddCommand(setDefaultVaultCmd)
}
