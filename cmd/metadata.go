package cmd

import (
	"strings"

	"github.com/andy-neoaira/obs-cli/pkg/frontmatter"
	"github.com/andy-neoaira/obs-cli/pkg/noteops"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

func newMetadataCommand(registryFactory vaultRegistryFactory, serviceFactory noteServiceFactory) *cobra.Command {
	var common commonFlags
	command := newNamespace("metadata", "Read and safely update note frontmatter", &common)

	get := &cobra.Command{Use: "get <path>", Args: namespaceArgs(&common, "metadata.get", cobra.ExactArgs(1))}
	get.RunE = func(cmd *cobra.Command, values []string) error {
		return renderEnvelope(cmd, "metadata.get", common.RequestID, func() (any, error) {
			service, vault, err := resolveNoteContext(common, registryFactory, serviceFactory)
			if err != nil {
				return nil, err
			}
			note, err := service.Get(values[0])
			return map[string]any{"vault": vault, "note": map[string]any{
				"path": note.Path, "revision": note.Revision, "body_revision": note.BodyRevision,
				"frontmatter": note.Frontmatter,
			}}, err
		})
	}
	command.AddCommand(get)

	var key, value string
	var setFlags commonFlags
	set := &cobra.Command{Use: "set <path>", Args: namespaceArgs(&common, "metadata.set", cobra.ExactArgs(1))}
	set.RunE = func(cmd *cobra.Command, values []string) error {
		return renderEnvelope(cmd, "metadata.set", common.RequestID, func() (any, error) {
			if strings.TrimSpace(key) == "" {
				return nil, protocol.NewError(protocol.InvalidArgument, "--key is required", map[string]any{"field": "key"})
			}
			service, vault, err := resolveNoteContext(common, registryFactory, serviceFactory)
			if err != nil {
				return nil, err
			}
			note, err := service.Get(values[0])
			if err != nil {
				return nil, err
			}
			if setFlags.IfMatch == "" {
				return nil, noteops.ErrRevisionRequired
			}
			updated, err := frontmatter.SetKey(note.Content, key, value)
			if err != nil {
				return nil, noteops.ErrInvalidFrontmatter
			}
			var mutation noteops.Mutation
			if setFlags.DryRun {
				mutation, err = service.PlanReplace(values[0], []byte(updated), setFlags.IfMatch)
			} else {
				mutation, err = service.Replace(values[0], []byte(updated), setFlags.IfMatch)
			}
			return map[string]any{"key": key, "value": value, "body_revision": note.BodyRevision,
				"result": mutationResponse(vault, mutation, setFlags.DryRun, nil)}, err
		})
	}
	set.Flags().StringVar(&key, "key", "", "frontmatter key to set")
	set.Flags().StringVar(&value, "value", "", "frontmatter value (string, boolean, or list)")
	bindCommonFlags(set, &setFlags, commonFlagSet{DryRun: true, IfMatch: true}, false)
	command.AddCommand(set)
	return command
}
