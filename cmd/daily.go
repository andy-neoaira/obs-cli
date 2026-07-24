package cmd

import (
	"github.com/andy-neoaira/obs-cli/pkg/actions"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var DailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Creates or opens daily note in vault",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		uri := obsidian.Uri{}
		resolvedContent, err := resolveContentInput(dailyContent, dailyContentFile)
		if err != nil {
			return err
		}

		return actions.DailyNote(&vault, &uri, actions.DailyParams{
			Content:   resolvedContent,
			UseEditor: resolveUseEditor(cmd, &vault),
		})
	},
}

var dailyContent string
var dailyContentFile string

func init() {
	DailyCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")
	DailyCmd.Flags().StringVarP(&dailyContent, "content", "c", "", "text to add to daily note (appends if note exists)")
	DailyCmd.Flags().StringVar(&dailyContentFile, "content-file", "", "read daily note content from a file, or '-' for stdin")
	DailyCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian")
}
