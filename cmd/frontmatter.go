package cmd

import (
	"fmt"
	"log"

	"github.com/andy-neoaira/obs-cli/pkg/actions"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

// frontmatter 相关命令行参数变量。
var fmPrint bool
var fmEdit bool
var fmDelete bool
var fmKey string
var fmValue string

// frontmatterCmd 定义了 "frontmatter" 子命令，用于查看或修改笔记的 YAML frontmatter。
var frontmatterCmd = &cobra.Command{
	Use:   "frontmatter <note>",
	Short: "View or modify note frontmatter",
	Long: `View or modify YAML frontmatter in a note.

Use --print to display frontmatter, --edit to modify a key,
or --delete to remove a key.

Examples:
  obs-cli frontmatter "My Note" --print
  obs-cli frontmatter "My Note" --edit --key "status" --value "done"
  obs-cli frontmatter "My Note" --delete --key "draft"`,
	Args: cobra.ExactArgs(1), // 必须提供 1 个参数：笔记名称
	Run: func(cmd *cobra.Command, args []string) {
		noteName := args[0]
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}

		params := actions.FrontmatterParams{
			NoteName: noteName,
			Print:    fmPrint,
			Edit:     fmEdit,
			Delete:   fmDelete,
			Key:      fmKey,
			Value:    fmValue,
		}

		output, err := actions.Frontmatter(&vault, &note, params)
		if err != nil {
			log.Fatal(err)
		}

		if output != "" {
			fmt.Print(output)
		}
	},
}

func init() {
	frontmatterCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	frontmatterCmd.Flags().BoolVarP(&fmPrint, "print", "p", false, "print frontmatter")
	frontmatterCmd.Flags().BoolVarP(&fmEdit, "edit", "e", false, "edit a frontmatter key")
	frontmatterCmd.Flags().BoolVarP(&fmDelete, "delete", "d", false, "delete a frontmatter key")
	frontmatterCmd.Flags().StringVarP(&fmKey, "key", "k", "", "key to edit or delete")
	frontmatterCmd.Flags().StringVar(&fmValue, "value", "", "value to set (required for --edit)")
	rootCmd.AddCommand(frontmatterCmd)
}
