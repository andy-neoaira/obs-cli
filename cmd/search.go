package cmd

import (
	"strings"

	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

func newSearchCommand(registryFactory vaultRegistryFactory, serviceFactory noteServiceFactory) *cobra.Command {
	var common commonFlags
	command := newNamespace("search", "Search Markdown notes with bounded revision evidence", &common)
	var scope string
	var page, pageSize, maxFiles int
	content := &cobra.Command{
		Use:  "content <query>",
		Args: namespaceArgs(&common, "search.content", cobra.ExactArgs(1)),
	}
	content.RunE = func(cmd *cobra.Command, values []string) error {
		return renderEnvelope(cmd, "search.content", common.RequestID, func() (any, error) {
			if strings.TrimSpace(values[0]) == "" {
				return nil, invalidSearchArgument("query", "query must not be empty")
			}
			if page < 1 {
				return nil, invalidSearchArgument("page", "page must be at least 1")
			}
			if pageSize < 1 || pageSize > 100 {
				return nil, invalidSearchArgument("page-size", "page-size must be between 1 and 100")
			}
			if maxFiles < 1 || maxFiles > 5000 {
				return nil, invalidSearchArgument("max-files", "max-files must be between 1 and 5000")
			}
			service, vault, err := resolveNoteContext(common, registryFactory, serviceFactory)
			if err != nil {
				return nil, err
			}
			result, err := service.Search(values[0], scope, page, pageSize, maxFiles)
			return map[string]any{"vault": vault, "search": result}, err
		})
	}
	content.Flags().StringVar(&scope, "scope", "", "optional Vault-relative directory")
	content.Flags().IntVar(&page, "page", 1, "one-based result page")
	content.Flags().IntVar(&pageSize, "page-size", 25, "results per page (max 100)")
	content.Flags().IntVar(&maxFiles, "max-files", 1000, "maximum Markdown files to read (max 5000)")
	command.AddCommand(content)
	return command
}

func invalidSearchArgument(field, message string) error {
	return protocol.NewError(protocol.InvalidArgument, message, map[string]any{"field": field})
}
