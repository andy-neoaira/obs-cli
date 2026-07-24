package cmd

import (
	"errors"
	"fmt"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

type vaultRegistry interface {
	List() ([]config.VaultRecord, error)
	Get(string) (config.VaultRecord, error)
	Add(string, string) (config.VaultRecord, error)
	Remove(string) (config.VaultRecord, error)
	SetDefault(string) (config.VaultRecord, error)
	Default() (config.VaultRecord, error)
	MigrateLegacy(string, []obsidian.DiscoveredVault) (config.V2Config, error)
}

type vaultRegistryFactory func() (vaultRegistry, error)
type vaultDiscoverFunc func() ([]obsidian.DiscoveredVault, error)

func defaultVaultRegistryFactory() (vaultRegistry, error) {
	return obsidian.DefaultRegistry()
}

func newVaultV2Command(factory vaultRegistryFactory, discover vaultDiscoverFunc) *cobra.Command {
	var output string
	var requestID string

	command := &cobra.Command{
		Use:   "vault",
		Short: "Manage obs-cli V2 vault registry",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			operation := "vault." + cmd.Name()
			if output != "json" {
				err := protocol.NewError(
					protocol.InvalidArgument,
					fmt.Sprintf("unsupported output format %q: only json is supported", output),
					map[string]any{"field": "output"},
				)
				resolved, resolveErr := protocol.ResolveRequestID(requestID)
				if resolveErr != nil {
					resolved, _ = protocol.ResolveRequestID("")
				}
				requestID = resolved
				return renderV2(cmd, operation, requestID, func() (any, error) { return nil, err })
			}
			resolved, err := protocol.ResolveRequestID(requestID)
			if err != nil {
				requestID, _ = protocol.ResolveRequestID("")
				return renderV2(cmd, operation, requestID, func() (any, error) { return nil, err })
			}
			requestID = resolved
			return nil
		},
	}
	command.PersistentFlags().StringVar(&output, "output", "json", "output format (json)")
	command.PersistentFlags().StringVar(&requestID, "request-id", "", "caller-provided request identifier")

	execute := func(cmd *cobra.Command, operation string, run func() (any, error)) error {
		return renderV2(cmd, operation, requestID, run)
	}
	args := func(operation string, validate cobra.PositionalArgs) cobra.PositionalArgs {
		return func(cmd *cobra.Command, values []string) error {
			if err := validate(cmd, values); err != nil {
				resolved, resolveErr := protocol.ResolveRequestID(requestID)
				if resolveErr != nil {
					resolved, _ = protocol.ResolveRequestID("")
				}
				requestID = resolved
				domain := protocol.Wrap(
					protocol.InvalidArgument,
					"invalid command arguments",
					err,
					map[string]any{"command": cmd.CommandPath()},
				)
				return renderV2(cmd, operation, requestID, func() (any, error) {
					return nil, domain
				})
			}
			return nil
		}
	}
	registry := func() (vaultRegistry, error) {
		return factory()
	}

	command.AddCommand(&cobra.Command{
		Use:   "discover",
		Short: "Read vaults from Obsidian configuration without modifying it",
		Args:  args("vault.discover", cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return execute(cmd, "vault.discover", func() (any, error) {
				vaults, err := discover()
				return map[string]any{"vaults": vaults}, err
			})
		},
	})

	command.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List vaults registered in obs-cli V2",
		Args:  args("vault.list", cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return execute(cmd, "vault.list", func() (any, error) {
				instance, err := registry()
				if err != nil {
					return nil, err
				}
				vaults, err := instance.List()
				if err != nil {
					return nil, err
				}
				var defaultID string
				if selected, defaultErr := instance.Default(); defaultErr == nil {
					defaultID = selected.ID
				} else if !errors.Is(defaultErr, obsidian.ErrVaultNotFound) {
					return nil, defaultErr
				}
				return map[string]any{
					"vaults":           vaults,
					"default_vault_id": defaultID,
				}, nil
			})
		},
	})

	command.AddCommand(&cobra.Command{
		Use:   "get <id-or-name>",
		Short: "Get one obs-cli V2 vault",
		Args:  args("vault.get", cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(cmd, "vault.get", func() (any, error) {
				instance, err := registry()
				if err != nil {
					return nil, err
				}
				vault, err := instance.Get(args[0])
				return map[string]any{"vault": vault}, err
			})
		},
	})

	var addName string
	var addDefault bool
	add := &cobra.Command{
		Use:   "add <path>",
		Short: "Register a directory in obs-cli V2",
		Args:  args("vault.add", cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(cmd, "vault.add", func() (any, error) {
				instance, err := registry()
				if err != nil {
					return nil, err
				}
				vault, err := instance.Add(args[0], addName)
				if err != nil {
					return nil, err
				}
				if addDefault {
					vault, err = instance.SetDefault(vault.ID)
					if err != nil {
						return nil, err
					}
				}
				return map[string]any{"vault": vault, "set_default": addDefault}, nil
			})
		},
	}
	add.Flags().StringVar(&addName, "name", "", "vault display name (defaults to directory basename)")
	add.Flags().BoolVar(&addDefault, "set-default", false, "set the added vault as default")
	command.AddCommand(add)

	command.AddCommand(&cobra.Command{
		Use:   "remove <id-or-name>",
		Short: "Remove a vault from obs-cli V2 without deleting files",
		Args:  args("vault.remove", cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(cmd, "vault.remove", func() (any, error) {
				instance, err := registry()
				if err != nil {
					return nil, err
				}
				vault, err := instance.Remove(args[0])
				return map[string]any{"vault": vault, "files_deleted": false}, err
			})
		},
	})

	command.AddCommand(&cobra.Command{
		Use:   "set-default <id-or-name>",
		Short: "Set the default obs-cli V2 vault",
		Args:  args("vault.set-default", cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(cmd, "vault.set-default", func() (any, error) {
				instance, err := registry()
				if err != nil {
					return nil, err
				}
				vault, err := instance.SetDefault(args[0])
				return map[string]any{"vault": vault}, err
			})
		},
	})

	command.AddCommand(&cobra.Command{
		Use:   "migrate",
		Short: "Import legacy obs-cli preferences and discovered Obsidian vaults once",
		Args:  args("vault.migrate", cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return execute(cmd, "vault.migrate", func() (any, error) {
				instance, err := registry()
				if err != nil {
					return nil, err
				}
				vaults, err := discover()
				if err != nil {
					return nil, err
				}
				_, legacyPath, err := config.CliPath()
				if err != nil {
					return nil, err
				}
				cfg, err := instance.MigrateLegacy(legacyPath, vaults)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"config_version":    cfg.Version,
					"vaults":            config.SortedVaults(cfg),
					"default_vault_id":  cfg.DefaultVaultID,
					"migration_warning": cfg.MigrationWarning,
				}, nil
			})
		},
	})

	return command
}

func init() {
	rootCmd.AddCommand(newVaultV2Command(defaultVaultRegistryFactory, obsidian.DiscoverObsidianVaults))
}
