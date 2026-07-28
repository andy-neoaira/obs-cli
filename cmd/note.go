package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/andy-neoaira/obs-cli/pkg/config"
	"github.com/andy-neoaira/obs-cli/pkg/noteops"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/andy-neoaira/obs-cli/pkg/storage"
	"github.com/spf13/cobra"
)

type noteService interface {
	List() ([]string, error)
	Get(string) (noteops.Note, error)
	PlanCreate(string, []byte) (noteops.Mutation, error)
	Create(string, []byte) (noteops.Mutation, error)
	PlanAppend(string, []byte, string, string) (noteops.Mutation, error)
	Append(string, []byte, string, string) (noteops.Mutation, error)
	PlanPatch(string, []byte, []byte, string) (noteops.Mutation, error)
	Patch(string, []byte, []byte, string) (noteops.Mutation, error)
	PlanReplace(string, []byte, string) (noteops.Mutation, error)
	Replace(string, []byte, string) (noteops.Mutation, error)
	PlanDelete(string, string) (noteops.Mutation, error)
	Delete(string, string) (noteops.Mutation, error)
	PlanMove(string, string, string) (noteops.MovePlan, error)
	ApplyMovePlan(noteops.MovePlan) (noteops.MoveResult, error)
	Search(string, string, int, int, int) (noteops.SearchPage, error)
	Backlinks(string, string, int) (noteops.BacklinkReport, error)
}

type noteServiceFactory func(string) (noteService, error)

func defaultNoteServiceFactory(vaultRoot string) (noteService, error) {
	return noteops.NewService(vaultRoot, storage.DefaultStore())
}

func newNoteCommand(registryFactory vaultRegistryFactory, serviceFactory noteServiceFactory) *cobra.Command {
	var common commonFlags
	command := &cobra.Command{
		Use:   "note",
		Short: "Read and safely mutate notes through the obs-cli protocol",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			operation := "note." + cmd.Name()
			resolved, err := protocol.ResolveRequestID(common.RequestID)
			if err != nil {
				common.RequestID, _ = protocol.ResolveRequestID("")
				return renderEnvelope(cmd, operation, common.RequestID, func() (any, error) { return nil, err })
			}
			common.RequestID = resolved
			if common.Output != "json" {
				domain := protocol.NewError(
					protocol.InvalidArgument,
					fmt.Sprintf("unsupported output format %q: only json is supported", common.Output),
					map[string]any{"field": "output"},
				)
				return renderEnvelope(cmd, operation, common.RequestID, func() (any, error) { return nil, domain })
			}
			return nil
		},
	}
	bindCommonFlags(command, &common, commonFlagSet{Output: true, RequestID: true, Vault: true}, true)

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
				return renderEnvelope(cmd, operation, common.RequestID, func() (any, error) { return nil, domain })
			}
			return nil
		}
	}
	execute := func(cmd *cobra.Command, operation string, run func() (any, error)) error {
		return renderEnvelope(cmd, operation, common.RequestID, run)
	}
	resolveService := func() (noteService, config.VaultRecord, error) {
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

	command.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List Markdown notes in the selected Vault",
		Args:  args("note.list", cobra.NoArgs),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return execute(cmd, "note.list", func() (any, error) {
				service, vault, err := resolveService()
				if err != nil {
					return nil, err
				}
				notes, err := service.List()
				return map[string]any{"vault": vault, "notes": notes}, err
			})
		},
	})

	command.AddCommand(&cobra.Command{
		Use:   "get <path>",
		Short: "Read a stable note snapshot with revision and frontmatter",
		Args:  args("note.get", cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, values []string) error {
			return execute(cmd, "note.get", func() (any, error) {
				service, vault, err := resolveService()
				if err != nil {
					return nil, err
				}
				note, err := service.Get(values[0])
				return map[string]any{"vault": vault, "note": note}, err
			})
		},
	})

	var createInput string
	var createFlags commonFlags
	create := &cobra.Command{
		Use:   "create <path>",
		Short: "Create a note without overwriting an existing target",
		Args:  args("note.create", cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, values []string) error {
			return execute(cmd, "note.create", func() (any, error) {
				content, err := readNoteInput(cmd, createInput, "content-file")
				if err != nil {
					return nil, err
				}
				service, vault, err := resolveService()
				if err != nil {
					return nil, err
				}
				var mutation noteops.Mutation
				if createFlags.DryRun {
					mutation, err = service.PlanCreate(values[0], content)
				} else {
					mutation, err = service.Create(values[0], content)
				}
				return mutationResponse(vault, mutation, createFlags.DryRun, nil), err
			})
		},
	}
	create.Flags().StringVar(&createInput, "content-file", "", "Markdown input file, or - for stdin")
	bindCommonFlags(create, &createFlags, commonFlagSet{DryRun: true}, false)
	command.AddCommand(create)

	var appendInput string
	var appendSection string
	var appendFlags commonFlags
	appendCommand := &cobra.Command{
		Use:   "append <path>",
		Short: "Append Markdown using an atomic revision-aware replacement",
		Args:  args("note.append", cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, values []string) error {
			return execute(cmd, "note.append", func() (any, error) {
				content, err := readNoteInput(cmd, appendInput, "content-file")
				if err != nil {
					return nil, err
				}
				service, vault, err := resolveService()
				if err != nil {
					return nil, err
				}
				var mutation noteops.Mutation
				if appendFlags.DryRun {
					mutation, err = service.PlanAppend(values[0], content, appendSection, appendFlags.IfMatch)
				} else {
					mutation, err = service.Append(values[0], content, appendSection, appendFlags.IfMatch)
				}
				return mutationResponse(vault, mutation, appendFlags.DryRun, nil), err
			})
		},
	}
	appendCommand.Flags().StringVar(&appendInput, "content-file", "", "Markdown input file, or - for stdin")
	appendCommand.Flags().StringVar(&appendSection, "section", "", "exact ATX heading title to append within")
	bindCommonFlags(appendCommand, &appendFlags, commonFlagSet{DryRun: true, IfMatch: true}, false)
	command.AddCommand(appendCommand)

	var patchMatchInput string
	var patchReplacementInput string
	var patchFlags commonFlags
	patch := &cobra.Command{
		Use:   "patch <path>",
		Short: "Replace one uniquely matching context block",
		Args:  args("note.patch", cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, values []string) error {
			return execute(cmd, "note.patch", func() (any, error) {
				if patchMatchInput == "-" && patchReplacementInput == "-" {
					return nil, protocol.NewError(
						protocol.InvalidArgument,
						"only one patch input may consume stdin",
						map[string]any{"fields": []string{"match-file", "content-file"}},
					)
				}
				match, err := readNoteInput(cmd, patchMatchInput, "match-file")
				if err != nil {
					return nil, err
				}
				replacement, err := readNoteInput(cmd, patchReplacementInput, "content-file")
				if err != nil {
					return nil, err
				}
				service, vault, err := resolveService()
				if err != nil {
					return nil, err
				}
				var mutation noteops.Mutation
				if patchFlags.DryRun {
					mutation, err = service.PlanPatch(values[0], match, replacement, patchFlags.IfMatch)
				} else {
					mutation, err = service.Patch(values[0], match, replacement, patchFlags.IfMatch)
				}
				return mutationResponse(vault, mutation, patchFlags.DryRun, nil), err
			})
		},
	}
	patch.Flags().StringVar(&patchMatchInput, "match-file", "", "unique context input file, or - for stdin")
	patch.Flags().StringVar(&patchReplacementInput, "content-file", "", "replacement input file, or - for stdin")
	bindCommonFlags(patch, &patchFlags, commonFlagSet{DryRun: true, IfMatch: true}, false)
	command.AddCommand(patch)

	var replaceInput string
	var replaceUnsafe bool
	var replaceFlags commonFlags
	replace := &cobra.Command{
		Use:   "replace <path>",
		Short: "Replace an entire note with an explicit revision precondition",
		Args:  args("note.replace", cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, values []string) error {
			return execute(cmd, "note.replace", func() (any, error) {
				content, err := readNoteInput(cmd, replaceInput, "content-file")
				if err != nil {
					return nil, err
				}
				service, vault, err := resolveService()
				if err != nil {
					return nil, err
				}
				revision, risks, err := effectiveRevision(service, values[0], replaceFlags.IfMatch, replaceUnsafe)
				if err != nil {
					return nil, err
				}
				var mutation noteops.Mutation
				if replaceFlags.DryRun {
					mutation, err = service.PlanReplace(values[0], content, revision)
				} else {
					mutation, err = service.Replace(values[0], content, revision)
				}
				return mutationResponse(vault, mutation, replaceFlags.DryRun, risks), err
			})
		},
	}
	replace.Flags().StringVar(&replaceInput, "content-file", "", "Markdown input file, or - for stdin")
	replace.Flags().BoolVar(&replaceUnsafe, "unsafe-no-if-match", false, "replace using the revision read immediately before this operation")
	bindCommonFlags(replace, &replaceFlags, commonFlagSet{DryRun: true, IfMatch: true}, false)
	command.AddCommand(replace)

	var deleteUnsafe bool
	var deleteFlags commonFlags
	deleteCommand := &cobra.Command{
		Use:   "delete <path>",
		Short: "Delete a note while retaining a private recovery copy",
		Args:  args("note.delete", cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, values []string) error {
			return execute(cmd, "note.delete", func() (any, error) {
				service, vault, err := resolveService()
				if err != nil {
					return nil, err
				}
				revision, risks, err := effectiveRevision(service, values[0], deleteFlags.IfMatch, deleteUnsafe)
				if err != nil {
					return nil, err
				}
				var mutation noteops.Mutation
				if deleteFlags.DryRun {
					mutation, err = service.PlanDelete(values[0], revision)
				} else {
					mutation, err = service.Delete(values[0], revision)
				}
				return mutationResponse(vault, mutation, deleteFlags.DryRun, risks), err
			})
		},
	}
	deleteCommand.Flags().BoolVar(&deleteUnsafe, "unsafe-no-if-match", false, "delete using the revision read immediately before this operation")
	bindCommonFlags(deleteCommand, &deleteFlags, commonFlagSet{DryRun: true, IfMatch: true}, false)
	command.AddCommand(deleteCommand)

	var moveFlags commonFlags
	var movePlanHash string
	move := &cobra.Command{
		Use:   "move <source> <target>",
		Short: "Move a note and transactionally rewrite affected links",
		Args:  args("note.move", cobra.ExactArgs(2)),
		RunE: func(cmd *cobra.Command, values []string) error {
			return execute(cmd, "note.move", func() (any, error) {
				if movePlanHash != "" && !storage.IsRevision(movePlanHash) {
					return nil, protocol.NewError(
						protocol.InvalidArgument,
						"--if-plan-hash must use sha256:<64-lowercase-hex>",
						map[string]any{"field": "if-plan-hash"},
					)
				}
				service, vault, err := resolveService()
				if err != nil {
					return nil, err
				}
				plan, err := service.PlanMove(values[0], values[1], moveFlags.IfMatch)
				if err != nil {
					return nil, err
				}
				if moveFlags.DryRun {
					changes := make([]protocol.PlanChange, 0, len(plan.Changes))
					for _, change := range plan.Changes {
						target := change.Path
						if change.Target != "" {
							target += " -> " + change.Target
						}
						changes = append(changes, protocol.PlanChange{
							Action:   change.Action,
							Resource: "note",
							Target:   target,
							Details: map[string]any{
								"expected_revision": change.ExpectedRevision,
								"revision_after":    change.RevisionAfter,
								"link_edits":        change.LinkEdits,
							},
						})
					}
					dry := protocol.NewDryRunData(
						changes,
						plan.Risks,
						[]string{"source and rewritten note revisions remain unchanged", "target remains absent"},
					)
					return map[string]any{
						"vault":     vault,
						"dry_run":   dry.DryRun,
						"applied":   dry.Applied,
						"changed":   dry.Changed,
						"plan_hash": plan.PlanHash,
						"plan":      dry.Plan,
					}, nil
				}
				if movePlanHash != "" && movePlanHash != plan.PlanHash {
					return nil, protocol.NewError(
						protocol.RevisionConflict,
						"the move plan changed after authorization",
						map[string]any{
							"expected_plan_hash": movePlanHash,
							"actual_plan_hash":   plan.PlanHash,
						},
					)
				}
				result, err := service.ApplyMovePlan(plan)
				if err != nil {
					return nil, err
				}
				sourceRevision := ""
				if len(plan.Changes) != 0 {
					sourceRevision = plan.Changes[0].ExpectedRevision
				}
				receipt := map[string]any{
					"operation":            "note.move",
					"request_id":           common.RequestID,
					"transaction_id":       result.TransactionID,
					"plan_hash":            result.PlanHash,
					"vault_id":             vault.ID,
					"source":               result.Source,
					"source_revision":      sourceRevision,
					"source_digest":        sourceRevision,
					"target":               result.Target,
					"target_revision":      result.RevisionAfter,
					"target_body_revision": result.TargetBodyRevision,
				}
				return map[string]any{"vault": vault, "move": result, "receipt": receipt}, nil
			})
		},
	}
	move.Flags().StringVar(&movePlanHash, "if-plan-hash", "", "apply only if the current move plan matches this dry-run hash")
	bindCommonFlags(move, &moveFlags, commonFlagSet{DryRun: true, IfMatch: true}, false)
	command.AddCommand(move)

	return command
}

func readNoteInput(cmd *cobra.Command, input, field string) ([]byte, error) {
	if input == "" {
		return nil, protocol.NewError(
			protocol.InvalidArgument,
			fmt.Sprintf("--%s is required", field),
			map[string]any{"field": field},
		)
	}
	if input == "-" {
		data, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return nil, fmt.Errorf("read %s from stdin: %w", field, err)
		}
		return data, nil
	}
	data, err := os.ReadFile(input)
	if err != nil {
		return nil, protocol.Wrap(
			protocol.InvalidArgument,
			fmt.Sprintf("cannot read --%s input", field),
			err,
			map[string]any{"field": field},
		)
	}
	return data, nil
}

func effectiveRevision(service noteService, path, requested string, unsafe bool) (string, []string, error) {
	if requested != "" {
		if unsafe {
			return "", nil, protocol.NewError(
				protocol.InvalidArgument,
				"--if-match and --unsafe-no-if-match cannot be used together",
				map[string]any{"fields": []string{"if-match", "unsafe-no-if-match"}},
			)
		}
		return requested, []string{}, nil
	}
	if !unsafe {
		return "", nil, protocol.NewError(
			protocol.InvalidArgument,
			"--if-match is required unless --unsafe-no-if-match is explicit",
			map[string]any{"field": "if-match"},
		)
	}
	note, err := service.Get(path)
	if err != nil {
		return "", nil, err
	}
	return note.Revision, []string{"caller did not provide a revision; concurrent intent may be overwritten"}, nil
}

func mutationResponse(
	vault config.VaultRecord,
	mutation noteops.Mutation,
	dryRun bool,
	risks []string,
) any {
	if !dryRun {
		return map[string]any{"vault": vault, "note": mutation}
	}
	dry := protocol.NewDryRunData(
		[]protocol.PlanChange{{
			Action:   mutation.Action,
			Resource: "note",
			Target:   mutation.Path,
			Details: map[string]any{
				"revision_before": mutation.RevisionBefore,
				"revision_after":  mutation.RevisionAfter,
				"changed":         mutation.Changed,
			},
		}},
		risks,
		[]string{"Vault path policy remains satisfied", "revision precondition remains satisfied at apply time"},
	)
	dry.Changed = mutation.Changed
	return map[string]any{
		"vault":   vault,
		"dry_run": dry.DryRun,
		"applied": dry.Applied,
		"changed": dry.Changed,
		"plan":    dry.Plan,
	}
}
