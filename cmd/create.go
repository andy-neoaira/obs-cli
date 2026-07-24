package cmd

import (
	"github.com/andy-neoaira/obs-cli/pkg/actions"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var shouldAppend bool
var shouldOverwrite bool
var content string
var contentFile string
var createNoteCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates note in vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		uri := obsidian.Uri{}
		noteName := args[0]
		resolvedContent, err := resolveContentInput(content, contentFile)
		if err != nil {
			return err
		}

		params := actions.CreateParams{
			NoteName:        noteName,
			Content:         resolvedContent,
			ShouldAppend:    shouldAppend,
			ShouldOverwrite: shouldOverwrite,
			ShouldOpen:      shouldOpen,
			UseEditor:       resolveUseEditor(cmd, &vault),
		}
		return actions.CreateNote(&vault, &uri, params)
	},
}

func init() {
	createNoteCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	createNoteCmd.Flags().BoolVarP(&shouldOpen, "open", "", false, "open created note")
	createNoteCmd.Flags().StringVarP(&content, "content", "c", "", "text to add to note")
	createNoteCmd.Flags().StringVar(&contentFile, "content-file", "", "read note content from a file, or '-' for stdin")
	createNoteCmd.Flags().BoolVarP(&shouldAppend, "append", "a", false, "append to note")
	createNoteCmd.Flags().BoolVarP(&shouldOverwrite, "overwrite", "o", false, "overwrite note")
	createNoteCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian (requires --open flag)")
	createNoteCmd.MarkFlagsMutuallyExclusive("append", "overwrite")
}
