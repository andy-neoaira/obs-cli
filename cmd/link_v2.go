package cmd

import (
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

func newLinkV2Command(registryFactory vaultRegistryFactory, serviceFactory noteServiceFactory) *cobra.Command {
	var common commonFlags
	command := newV2Namespace("link", "Inspect Obsidian links without modifying notes", &common)
	var scope string
	var maxFiles int
	backlinks := &cobra.Command{
		Use:  "backlinks <path>",
		Args: v2Args(&common, "link.backlinks", cobra.ExactArgs(1)),
	}
	backlinks.RunE = func(cmd *cobra.Command, values []string) error {
		return renderV2(cmd, "link.backlinks", common.RequestID, func() (any, error) {
			if maxFiles < 1 || maxFiles > 5000 {
				return nil, protocol.NewError(
					protocol.InvalidArgument, "max-files must be between 1 and 5000",
					map[string]any{"field": "max-files"},
				)
			}
			service, vault, err := resolveNoteContext(common, registryFactory, serviceFactory)
			if err != nil {
				return nil, err
			}
			report, err := service.Backlinks(values[0], scope, maxFiles)
			return map[string]any{"vault": vault, "backlinks": report}, err
		})
	}
	backlinks.Flags().StringVar(&scope, "scope", "", "optional Vault-relative source directory")
	backlinks.Flags().IntVar(&maxFiles, "max-files", 1000, "maximum Markdown files to read (max 5000)")
	command.AddCommand(backlinks)
	return command
}
