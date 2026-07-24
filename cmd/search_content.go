package cmd

import (
	"os"

	"github.com/andy-neoaira/obs-cli/pkg/actions"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var searchContentCmd = &cobra.Command{
	Use:   "search-content [search term]",
	Short: "Search note content for search term",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}
		uri := obsidian.Uri{}
		fuzzyFinder := obsidian.FuzzyFinder{}

		searchTerm := args[0]
		options, err := buildSearchContentOptions(cmd, &vault, isInteractiveTerminal())
		if err != nil {
			return err
		}

		return actions.SearchNotesContentWithOptions(&vault, &note, &uri, &fuzzyFinder, searchTerm, options)
	},
}

func buildSearchContentOptions(cmd *cobra.Command, vault obsidian.VaultManager, interactiveTerminal bool) (actions.SearchContentOptions, error) {
	noInteractive, err := cmd.Flags().GetBool("no-interactive")
	if err != nil {
		return actions.SearchContentOptions{}, err
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return actions.SearchContentOptions{}, err
	}

	page, err := cmd.Flags().GetInt("page")
	if err != nil {
		return actions.SearchContentOptions{}, err
	}

	pageSize, err := cmd.Flags().GetInt("page-size")
	if err != nil {
		return actions.SearchContentOptions{}, err
	}

	useEditor := resolveUseEditor(cmd, vault)

	return actions.SearchContentOptions{
		UseEditor:           useEditor,
		EditorFlagExplicit:  cmd.Flags().Changed("editor"),
		NoInteractive:       noInteractive,
		Format:              format,
		InteractiveTerminal: interactiveTerminal,
		Output:              os.Stdout,
		Page:                page,
		PageSize:            pageSize,
	}, nil
}

func isInteractiveTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

func init() {
	searchContentCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	searchContentCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian")
	searchContentCmd.Flags().Bool("no-interactive", false, "disable interactive selection and print results to stdout")
	searchContentCmd.Flags().String("format", "text", "output format for non-interactive mode: text|json")
	searchContentCmd.Flags().Int("page", 0, "page number for paginated results (enables pagination)")
	searchContentCmd.Flags().Int("page-size", 0, "results per page, max 100 (default 25 when pagination is enabled)")
}
