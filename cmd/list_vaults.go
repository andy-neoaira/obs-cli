package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

// listVaults 相关命令行参数变量。
var listVaultsJSON bool
var listVaultsPathOnly bool
var listVaultsDefault bool

// listVaultsCmd 定义了 "list-vaults" 子命令，用于列出所有已注册的 Obsidian vault。
// 支持多种输出格式：表格（默认）、JSON、仅路径、仅默认 vault。
var listVaultsCmd = &cobra.Command{
	Use:   "list-vaults",
	Short: "lists all registered Obsidian vaults",
	Args:  cobra.ExactArgs(0), // 不接受任何参数
	Run: func(cmd *cobra.Command, args []string) {
		registry, err := obsidian.DefaultRegistry()
		if err != nil {
			log.Fatal(err)
		}
		records, err := registry.List()
		if err != nil {
			log.Fatal(err)
		}
		vaults := make([]obsidian.VaultInfo, 0, len(records))
		for _, record := range records {
			vaults = append(vaults, obsidian.VaultInfo{Name: record.Name, Path: record.Path})
		}
		defaultName := ""
		if selected, defaultErr := registry.Default(); defaultErr == nil {
			defaultName = selected.Name
		}

		// 如果用户指定了 --default，只显示默认 vault 的信息
		if listVaultsDefault {
			runListVaultsDefault(vaults, defaultName)
			return
		}

		if len(vaults) == 0 {
			fmt.Println("No vaults registered. Use add-vault to register one.")
			return
		}

		// 按 vault 名称字母顺序排序，保证输出稳定
		sort.Slice(vaults, func(i, j int) bool {
			return vaults[i].Name < vaults[j].Name
		})

		// 根据用户选择的格式输出
		if listVaultsJSON {
			output, err := json.MarshalIndent(vaults, "", "  ")
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(output))
			return
		}

		if listVaultsPathOnly {
			for _, v := range vaults {
				fmt.Println(v.Path)
			}
		} else {
			formatVaultsTable(os.Stdout, vaults, defaultName)
		}
	},
}

// runListVaultsDefault 处理 --default flag 的逻辑，仅输出当前默认 vault 的信息。
func runListVaultsDefault(vaults []obsidian.VaultInfo, defaultName string) {
	if defaultName == "" {
		fmt.Println("No default vault set. Use set-default-vault to set one.")
		return
	}

	// 在已注册列表中找到默认 vault
	var defaultVault *obsidian.VaultInfo
	for _, v := range vaults {
		if v.Name == defaultName {
			defaultVault = &v
			break
		}
	}

	if defaultVault == nil {
		fmt.Printf("Default vault %q is set but not found in registered vaults.\n", defaultName)
		return
	}

	if listVaultsJSON {
		output, err := json.MarshalIndent(defaultVault, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(output))
		return
	}

	if listVaultsPathOnly {
		fmt.Println(defaultVault.Path)
		return
	}

	vault := obsidian.Vault{Name: defaultName}
	openType, _ := vault.DefaultOpenType()

	fmt.Println("Default vault name:", defaultVault.Name)
	fmt.Println("Default vault path:", defaultVault.Path)
	fmt.Println("Default open type:", openType)
}

// formatVaultsTable 使用 tabwriter 将 vault 列表格式化为对齐的表格输出。
// 默认 vault 会在行尾标注 (default)。
//
// 示例输出：
//
//	Notes          /home/user/Notes  (default)
//	LongVaultName  /home/user/LongVaultName
//	Work           /home/user/Work
func formatVaultsTable(w io.Writer, vaults []obsidian.VaultInfo, defaultName string) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, v := range vaults {
		if v.Name == defaultName {
			_, _ = fmt.Fprintf(tw, "%s\t%s\t(default)\n", v.Name, v.Path)
		} else {
			_, _ = fmt.Fprintf(tw, "%s\t%s\n", v.Name, v.Path)
		}
	}
	_ = tw.Flush()
}

// resolveDefaultVaultName 读取 CLI 配置，返回当前默认 vault 的名称。
// 如果没有设置默认 vault，返回空字符串。
func resolveDefaultVaultName() string {
	registry, err := obsidian.DefaultRegistry()
	if err != nil {
		return ""
	}
	vault, err := registry.Default()
	if err != nil {
		return ""
	}
	return vault.Name
}

func init() {
	listVaultsCmd.Flags().BoolVar(&listVaultsJSON, "json", false, "output as JSON array")
	listVaultsCmd.Flags().BoolVar(&listVaultsPathOnly, "path-only", false, "output one path per line")
	listVaultsCmd.Flags().BoolVar(&listVaultsDefault, "default", false, "show only the default vault")
	listVaultsCmd.MarkFlagsMutuallyExclusive("json", "path-only")
	rootCmd.AddCommand(listVaultsCmd)
}
