package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
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
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if output != "json" {
				return fmt.Errorf("unsupported output format %q: only json is supported", output)
			}
			if requestID == "" {
				requestID = newVaultRequestID()
			}
			return nil
		},
	}
	command.PersistentFlags().StringVar(&output, "output", "json", "output format (json)")
	command.PersistentFlags().StringVar(&requestID, "request-id", "", "caller-provided request identifier")

	write := func(cmd *cobra.Command, operation string, data any) error {
		return writeVaultEnvelope(cmd.OutOrStdout(), operation, requestID, data)
	}
	registry := func() (vaultRegistry, error) {
		return factory()
	}

	command.AddCommand(&cobra.Command{
		Use:   "discover",
		Short: "Read vaults from Obsidian configuration without modifying it",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			vaults, err := discover()
			if err != nil {
				return err
			}
			return write(cmd, "vault.discover", map[string]any{"vaults": vaults})
		},
	})

	command.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List vaults registered in obs-cli V2",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			instance, err := registry()
			if err != nil {
				return err
			}
			vaults, err := instance.List()
			if err != nil {
				return err
			}
			var defaultID string
			if selected, defaultErr := instance.Default(); defaultErr == nil {
				defaultID = selected.ID
			} else if !errors.Is(defaultErr, obsidian.ErrVaultNotFound) {
				return defaultErr
			}
			return write(cmd, "vault.list", map[string]any{
				"vaults":           vaults,
				"default_vault_id": defaultID,
			})
		},
	})

	command.AddCommand(&cobra.Command{
		Use:   "get <id-or-name>",
		Short: "Get one obs-cli V2 vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instance, err := registry()
			if err != nil {
				return err
			}
			vault, err := instance.Get(args[0])
			if err != nil {
				return err
			}
			return write(cmd, "vault.get", map[string]any{"vault": vault})
		},
	})

	var addName string
	var addDefault bool
	add := &cobra.Command{
		Use:   "add <path>",
		Short: "Register a directory in obs-cli V2",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instance, err := registry()
			if err != nil {
				return err
			}
			vault, err := instance.Add(args[0], addName)
			if err != nil {
				return err
			}
			if addDefault {
				vault, err = instance.SetDefault(vault.ID)
				if err != nil {
					return err
				}
			}
			return write(cmd, "vault.add", map[string]any{
				"vault":       vault,
				"set_default": addDefault,
			})
		},
	}
	add.Flags().StringVar(&addName, "name", "", "vault display name (defaults to directory basename)")
	add.Flags().BoolVar(&addDefault, "set-default", false, "set the added vault as default")
	command.AddCommand(add)

	command.AddCommand(&cobra.Command{
		Use:   "remove <id-or-name>",
		Short: "Remove a vault from obs-cli V2 without deleting files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instance, err := registry()
			if err != nil {
				return err
			}
			vault, err := instance.Remove(args[0])
			if err != nil {
				return err
			}
			return write(cmd, "vault.remove", map[string]any{
				"vault":         vault,
				"files_deleted": false,
			})
		},
	})

	command.AddCommand(&cobra.Command{
		Use:   "set-default <id-or-name>",
		Short: "Set the default obs-cli V2 vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instance, err := registry()
			if err != nil {
				return err
			}
			vault, err := instance.SetDefault(args[0])
			if err != nil {
				return err
			}
			return write(cmd, "vault.set-default", map[string]any{"vault": vault})
		},
	})

	command.AddCommand(&cobra.Command{
		Use:   "migrate",
		Short: "Import legacy obs-cli preferences and discovered Obsidian vaults once",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			instance, err := registry()
			if err != nil {
				return err
			}
			vaults, err := discover()
			if err != nil {
				return err
			}
			_, legacyPath, err := config.CliPath()
			if err != nil {
				return err
			}
			cfg, err := instance.MigrateLegacy(legacyPath, vaults)
			if err != nil {
				return err
			}
			return write(cmd, "vault.migrate", map[string]any{
				"config_version":    cfg.Version,
				"vaults":            config.SortedVaults(cfg),
				"default_vault_id":  cfg.DefaultVaultID,
				"migration_warning": cfg.MigrationWarning,
			})
		},
	})

	return command
}

func writeVaultEnvelope(writer io.Writer, operation, requestID string, data any) error {
	response := map[string]any{
		"protocol_version": "obs-cli/v2",
		"ok":               true,
		"operation":        operation,
		"request_id":       requestID,
		"data":             data,
		"warnings":         []any{},
	}
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(response)
}

func newVaultRequestID() string {
	value := make([]byte, 12)
	if _, err := rand.Read(value); err != nil {
		return "req-unavailable"
	}
	return "req-" + hex.EncodeToString(value)
}

func init() {
	rootCmd.AddCommand(newVaultV2Command(defaultVaultRegistryFactory, obsidian.DiscoverObsidianVaults))
}
