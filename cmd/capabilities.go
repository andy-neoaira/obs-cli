package cmd

import (
	"runtime"
	"sort"

	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

type capabilityOperation struct {
	Name        string   `json:"name"`
	Version     int      `json:"version"`
	Mutating    bool     `json:"mutating"`
	CommonFlags []string `json:"common_flags"`
}

type capabilitiesData struct {
	CLIVersion       string                `json:"cli_version"`
	ProtocolVersions []string              `json:"protocol_versions"`
	WriteProtocols   []string              `json:"write_protocols"`
	VaultContract    map[string]any        `json:"vault_contract"`
	Operations       []capabilityOperation `json:"operations"`
	FeatureFlags     map[string]bool       `json:"feature_flags"`
	Platform         map[string]any        `json:"platform"`
}

func currentCapabilities() capabilitiesData {
	return capabilitiesData{
		CLIVersion:       resolveVersion(),
		ProtocolVersions: []string{protocol.Version},
		WriteProtocols:   []string{"obs-write/v1"},
		VaultContract: map[string]any{
			"target":      "vault-contract/v1",
			"implemented": nil,
		},
		Operations: []capabilityOperation{
			{Name: "capabilities.get", Version: 1, CommonFlags: []string{"output", "request-id"}},
			{Name: "daily.append", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "if-match", "output", "request-id", "vault"}},
			{Name: "daily.create", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "output", "request-id", "vault"}},
			{Name: "daily.get", Version: 1, CommonFlags: []string{"output", "request-id", "vault"}},
			{Name: "link.backlinks", Version: 1, CommonFlags: []string{"output", "request-id", "vault"}},
			{Name: "metadata.get", Version: 1, CommonFlags: []string{"output", "request-id", "vault"}},
			{Name: "metadata.set", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "if-match", "output", "request-id", "vault"}},
			{Name: "note.append", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "if-match", "output", "request-id", "vault"}},
			{Name: "note.create", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "output", "request-id", "vault"}},
			{Name: "note.delete", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "if-match", "output", "request-id", "vault"}},
			{Name: "note.get", Version: 1, CommonFlags: []string{"output", "request-id", "vault"}},
			{Name: "note.list", Version: 1, CommonFlags: []string{"output", "request-id", "vault"}},
			{Name: "note.move", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "if-match", "output", "request-id", "vault"}},
			{Name: "note.patch", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "if-match", "output", "request-id", "vault"}},
			{Name: "note.replace", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "if-match", "output", "request-id", "vault"}},
			{Name: "search.content", Version: 1, CommonFlags: []string{"output", "request-id", "vault"}},
			{Name: "vault.add", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "output", "request-id"}},
			{Name: "vault.discover", Version: 1, CommonFlags: []string{"output", "request-id"}},
			{Name: "vault.get", Version: 1, CommonFlags: []string{"output", "request-id"}},
			{Name: "vault.list", Version: 1, CommonFlags: []string{"output", "request-id"}},
			{Name: "vault.migrate", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "output", "request-id"}},
			{Name: "vault.remove", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "output", "request-id"}},
			{Name: "vault.set-default", Version: 1, Mutating: true, CommonFlags: []string{"dry-run", "output", "request-id"}},
		},
		FeatureFlags: map[string]bool{
			"atomic_writes":             true,
			"daily_notes_v2":            true,
			"dry_run_plans":             true,
			"json_error_envelopes":      true,
			"link_inspection_v2":        true,
			"metadata_v2":               true,
			"multi_file_transactions":   true,
			"note_operations_v2":        true,
			"revision_preconditions":    true,
			"search_v2":                 true,
			"vault_discovery_read_only": true,
			"vault_path_policy":         true,
		},
		Platform: map[string]any{
			"os":   runtime.GOOS,
			"arch": runtime.GOARCH,
			"limitations": []string{
				"external applications do not participate in obs-cli cooperative locks",
			},
		},
	}
}

func newCapabilitiesCommand() *cobra.Command {
	var common commonFlags
	var required []string
	command := &cobra.Command{
		Use:   "capabilities",
		Short: "Describe the machine-readable obs-cli V2 capability surface",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			resolved, err := protocol.ResolveRequestID(common.RequestID)
			if err != nil {
				resolved, _ = protocol.ResolveRequestID("")
				return renderV2(cmd, "capabilities.get", resolved, func() (any, error) {
					return nil, err
				})
			}
			common.RequestID = resolved
			if common.Output != "json" {
				domain := protocol.NewError(
					protocol.InvalidArgument,
					"capabilities only supports JSON output",
					map[string]any{"field": "output"},
				)
				return renderV2(cmd, "capabilities.get", common.RequestID, func() (any, error) {
					return nil, domain
				})
			}
			data := currentCapabilities()
			available := make(map[string]struct{}, len(data.Operations))
			for _, operation := range data.Operations {
				available[operation.Name] = struct{}{}
			}
			missing := make([]string, 0)
			for _, name := range required {
				if _, ok := available[name]; !ok {
					missing = append(missing, name)
				}
			}
			if len(missing) != 0 {
				sort.Strings(missing)
				domain := protocol.NewError(
					protocol.CapabilityUnsupported,
					"one or more required capabilities are unavailable",
					map[string]any{"required": missing, "protocol_version": protocol.Version},
				)
				return renderV2(cmd, "capabilities.get", common.RequestID, func() (any, error) {
					return nil, domain
				})
			}
			return renderV2(cmd, "capabilities.get", common.RequestID, func() (any, error) {
				return data, nil
			})
		},
	}
	bindCommonFlags(command, &common, commonFlagSet{Output: true, RequestID: true}, false)
	command.Flags().StringSliceVar(&required, "require", nil, "required operation name (repeatable or comma-separated)")
	return command
}
