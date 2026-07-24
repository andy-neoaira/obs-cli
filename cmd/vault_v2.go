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
	PlanAdd(string, string) (obsidian.VaultRegistrationPlan, error)
	Add(string, string) (config.VaultRecord, error)
	Remove(string) (config.VaultRecord, error)
	SetDefault(string) (config.VaultRecord, error)
	Default() (config.VaultRecord, error)
	PlanMigrate(string, []obsidian.DiscoveredVault) (config.V2Config, error)
	MigrateLegacy(string, []obsidian.DiscoveredVault) (config.V2Config, error)
}

type vaultRegistryFactory func() (vaultRegistry, error)
type vaultDiscoverFunc func() ([]obsidian.DiscoveredVault, error)

func defaultVaultRegistryFactory() (vaultRegistry, error) {
	return obsidian.DefaultRegistry()
}

func newVaultV2Command(factory vaultRegistryFactory, discover vaultDiscoverFunc) *cobra.Command {
	var common commonFlags

	command := &cobra.Command{
		Use:   "vault",
		Short: "Manage obs-cli V2 vault registry",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			operation := "vault." + cmd.Name()
			if common.Output != "json" {
				err := protocol.NewError(
					protocol.InvalidArgument,
					fmt.Sprintf("unsupported output format %q: only json is supported", common.Output),
					map[string]any{"field": "output"},
				)
				resolved, resolveErr := protocol.ResolveRequestID(common.RequestID)
				if resolveErr != nil {
					resolved, _ = protocol.ResolveRequestID("")
				}
				common.RequestID = resolved
				return renderV2(cmd, operation, common.RequestID, func() (any, error) { return nil, err })
			}
			resolved, err := protocol.ResolveRequestID(common.RequestID)
			if err != nil {
				common.RequestID, _ = protocol.ResolveRequestID("")
				return renderV2(cmd, operation, common.RequestID, func() (any, error) { return nil, err })
			}
			common.RequestID = resolved
			return nil
		},
	}
	bindCommonFlags(command, &common, commonFlagSet{Output: true, RequestID: true}, true)

	execute := func(cmd *cobra.Command, operation string, run func() (any, error)) error {
		return renderV2(cmd, operation, common.RequestID, run)
	}
	args := func(operation string, validate cobra.PositionalArgs) cobra.PositionalArgs {
		return func(cmd *cobra.Command, values []string) error {
			if err := validate(cmd, values); err != nil {
				resolved, resolveErr := protocol.ResolveRequestID(common.RequestID)
				if resolveErr != nil {
					resolved, _ = protocol.ResolveRequestID("")
				}
				common.RequestID = resolved
				domain := protocol.Wrap(
					protocol.InvalidArgument,
					"invalid command arguments",
					err,
					map[string]any{"command": cmd.CommandPath()},
				)
				return renderV2(cmd, operation, common.RequestID, func() (any, error) {
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
	var addCommon commonFlags
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
				if addCommon.DryRun {
					plan, err := instance.PlanAdd(args[0], addName)
					if err != nil {
						return nil, err
					}
					changes := []protocol.PlanChange{{
						Action:   "create",
						Resource: "vault-registry",
						Target:   plan.Name,
						Details:  map[string]any{"canonical_path": plan.Path},
					}}
					if addDefault {
						changes = append(changes, protocol.PlanChange{
							Action:   "update",
							Resource: "vault-registry.default",
							Target:   plan.Name,
						})
					}
					return protocol.NewDryRunData(
						changes,
						[]string{},
						[]string{"target path exists and is a directory", "vault name and path remain unique"},
					), nil
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
	bindCommonFlags(add, &addCommon, commonFlagSet{DryRun: true}, false)
	command.AddCommand(add)

	var removeCommon commonFlags
	remove := &cobra.Command{
		Use:   "remove <id-or-name>",
		Short: "Remove a vault from obs-cli V2 without deleting files",
		Args:  args("vault.remove", cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(cmd, "vault.remove", func() (any, error) {
				instance, err := registry()
				if err != nil {
					return nil, err
				}
				if removeCommon.DryRun {
					vault, err := instance.Get(args[0])
					if err != nil {
						return nil, err
					}
					return protocol.NewDryRunData(
						[]protocol.PlanChange{{
							Action:   "delete",
							Resource: "vault-registry",
							Target:   vault.ID,
							Details: map[string]any{
								"name":          vault.Name,
								"files_deleted": false,
							},
						}},
						[]string{"the default vault selection is cleared when this vault is currently selected"},
						[]string{"vault remains registered"},
					), nil
				}
				vault, err := instance.Remove(args[0])
				return map[string]any{"vault": vault, "files_deleted": false}, err
			})
		},
	}
	bindCommonFlags(remove, &removeCommon, commonFlagSet{DryRun: true}, false)
	command.AddCommand(remove)

	var defaultCommon commonFlags
	setDefault := &cobra.Command{
		Use:   "set-default <id-or-name>",
		Short: "Set the default obs-cli V2 vault",
		Args:  args("vault.set-default", cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			return execute(cmd, "vault.set-default", func() (any, error) {
				instance, err := registry()
				if err != nil {
					return nil, err
				}
				if defaultCommon.DryRun {
					vault, err := instance.Get(args[0])
					if err != nil {
						return nil, err
					}
					return protocol.NewDryRunData(
						[]protocol.PlanChange{{
							Action:   "update",
							Resource: "vault-registry.default",
							Target:   vault.ID,
							Details:  map[string]any{"name": vault.Name},
						}},
						[]string{},
						[]string{"vault remains registered"},
					), nil
				}
				vault, err := instance.SetDefault(args[0])
				return map[string]any{"vault": vault}, err
			})
		},
	}
	bindCommonFlags(setDefault, &defaultCommon, commonFlagSet{DryRun: true}, false)
	command.AddCommand(setDefault)

	var migrateCommon commonFlags
	migrate := &cobra.Command{
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
				var cfg config.V2Config
				if migrateCommon.DryRun {
					cfg, err = instance.PlanMigrate(legacyPath, vaults)
				} else {
					cfg, err = instance.MigrateLegacy(legacyPath, vaults)
				}
				if err != nil {
					return nil, err
				}
				if migrateCommon.DryRun {
					changes := make([]protocol.PlanChange, 0, len(cfg.Vaults))
					for _, vault := range config.SortedVaults(cfg) {
						changes = append(changes, protocol.PlanChange{
							Action:   "create",
							Resource: "vault-registry",
							Target:   vault.ID,
							Details: map[string]any{
								"name":           vault.Name,
								"canonical_path": vault.Path,
							},
						})
					}
					return protocol.NewDryRunData(
						changes,
						[]string{"migration is refused when the V2 registry already contains vaults"},
						[]string{"discovered vault paths remain available"},
					), nil
				}
				return map[string]any{
					"config_version":    cfg.Version,
					"vaults":            config.SortedVaults(cfg),
					"default_vault_id":  cfg.DefaultVaultID,
					"migration_warning": cfg.MigrationWarning,
				}, nil
			})
		},
	}
	bindCommonFlags(migrate, &migrateCommon, commonFlagSet{DryRun: true}, false)
	command.AddCommand(migrate)

	return command
}

func init() {
	rootCmd.AddCommand(newVaultV2Command(defaultVaultRegistryFactory, obsidian.DiscoverObsidianVaults))
}
