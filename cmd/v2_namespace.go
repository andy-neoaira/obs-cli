package cmd

import (
	"fmt"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

func newV2Namespace(use, short string, common *commonFlags) *cobra.Command {
	command := &cobra.Command{
		Use: use, Short: short,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			operation := use + "." + cmd.Name()
			resolved, err := protocol.ResolveRequestID(common.RequestID)
			if err != nil {
				common.RequestID, _ = protocol.ResolveRequestID("")
				return renderV2(cmd, operation, common.RequestID, func() (any, error) { return nil, err })
			}
			common.RequestID = resolved
			if common.Output != "json" {
				domain := protocol.NewError(protocol.InvalidArgument,
					fmt.Sprintf("unsupported output format %q: only json is supported", common.Output),
					map[string]any{"field": "output"})
				return renderV2(cmd, operation, common.RequestID, func() (any, error) { return nil, domain })
			}
			return nil
		},
	}
	bindCommonFlags(command, common, commonFlagSet{Output: true, RequestID: true, Vault: true}, true)
	return command
}

func v2Args(common *commonFlags, operation string, validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, values []string) error {
		if err := validate(cmd, values); err != nil {
			resolved, resolveErr := protocol.ResolveRequestID(common.RequestID)
			if resolveErr != nil {
				resolved, _ = protocol.ResolveRequestID("")
			}
			common.RequestID = resolved
			domain := protocol.Wrap(protocol.InvalidArgument, "invalid command arguments", err,
				map[string]any{"command": cmd.CommandPath()})
			return renderV2(cmd, operation, common.RequestID, func() (any, error) { return nil, domain })
		}
		return nil
	}
}

func resolveNoteContext(common commonFlags, registryFactory vaultRegistryFactory, serviceFactory noteServiceFactory) (noteService, config.VaultRecord, error) {
	registry, err := registryFactory()
	if err != nil {
		return nil, config.VaultRecord{}, err
	}
	var vault config.VaultRecord
	if common.Vault == "" {
		vault, err = registry.Default()
	} else {
		vault, err = registry.Get(common.Vault)
	}
	if err != nil {
		return nil, config.VaultRecord{}, err
	}
	service, err := serviceFactory(vault.Path)
	return service, vault, err
}
