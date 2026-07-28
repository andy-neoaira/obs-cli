package cmd

import (
	"errors"
	"path"
	"strings"
	"time"

	"github.com/andy-neoaira/obs-cli/pkg/noteops"
	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

type dailyTarget struct {
	Path     string                    `json:"path"`
	Date     string                    `json:"date"`
	Config   obsidian.DailyNotesConfig `json:"config"`
	Warnings []string                  `json:"warnings"`
	content  []byte
}

func newDailyCommand(registryFactory vaultRegistryFactory, serviceFactory noteServiceFactory, now func() time.Time) *cobra.Command {
	var common commonFlags
	command := newNamespace("daily", "Resolve and safely mutate Obsidian Daily Notes", &common)
	if now == nil {
		now = time.Now
	}
	resolve := func(service noteService, vaultPath, date string, includeTemplate bool) (dailyTarget, error) {
		selected := now()
		if date != "" {
			parsed, err := time.ParseInLocation("2006-01-02", date, selected.Location())
			if err != nil {
				return dailyTarget{}, protocol.NewError(protocol.InvalidArgument, "date must use YYYY-MM-DD", map[string]any{"field": "date"})
			}
			selected = parsed
		}
		config, _, err := obsidian.LoadDailyNotesConfig(vaultPath)
		if err != nil {
			var configErr *obsidian.ConfigFileError
			details := map[string]any{"field": "daily.config"}
			if errors.As(err, &configErr) {
				details["config_file"] = configErr.File
				details["failure_kind"] = configErr.Kind
			}
			return dailyTarget{}, protocol.NewError(
				protocol.InvalidArgument,
				"Obsidian Daily Notes config is not usable",
				details,
			)
		}
		if config.Format == "" {
			config.Format = "YYYY-MM-DD"
		}
		layout, err := obsidian.ParseMomentToGoFormat(config.Format)
		if err != nil {
			return dailyTarget{}, protocol.NewError(protocol.InvalidArgument, err.Error(), map[string]any{"field": "daily.format"})
		}
		title := selected.Format(layout)
		target := dailyTarget{
			Path: path.Join(strings.ReplaceAll(config.Folder, "\\", "/"), title+".md"),
			Date: selected.Format("2006-01-02"), Config: config, Warnings: []string{},
		}
		if config.Template == "" || !includeTemplate {
			return target, nil
		}
		template, err := service.Get(config.Template)
		if err != nil {
			return dailyTarget{}, err
		}
		rendered, warnings, err := obsidian.RenderDailyTemplate(template.Content, selected, config.Format, title)
		if err != nil {
			return dailyTarget{}, err
		}
		target.content, target.Warnings = []byte(rendered), warnings
		return target, nil
	}

	var getDate string
	get := &cobra.Command{Use: "get", Args: namespaceArgs(&common, "daily.get", cobra.NoArgs)}
	get.RunE = func(cmd *cobra.Command, _ []string) error {
		return renderEnvelope(cmd, "daily.get", common.RequestID, func() (any, error) {
			service, vault, err := resolveNoteContext(common, registryFactory, serviceFactory)
			if err != nil {
				return nil, err
			}
			target, err := resolve(service, vault.Path, getDate, false)
			if err != nil {
				return nil, err
			}
			note, err := service.Get(target.Path)
			if errors.Is(err, noteops.ErrNoteNotFound) {
				return map[string]any{"vault": vault, "target": target, "exists": false}, nil
			}
			return map[string]any{"vault": vault, "target": target, "exists": true, "note": note}, err
		})
	}
	get.Flags().StringVar(&getDate, "date", "", "calendar date in YYYY-MM-DD (defaults to today)")
	command.AddCommand(get)

	var createDate string
	var createFlags commonFlags
	create := &cobra.Command{Use: "create", Args: namespaceArgs(&common, "daily.create", cobra.NoArgs)}
	create.RunE = func(cmd *cobra.Command, _ []string) error {
		return renderEnvelope(cmd, "daily.create", common.RequestID, func() (any, error) {
			service, vault, err := resolveNoteContext(common, registryFactory, serviceFactory)
			if err != nil {
				return nil, err
			}
			target, err := resolve(service, vault.Path, createDate, true)
			if err != nil {
				return nil, err
			}
			var mutation noteops.Mutation
			if createFlags.DryRun {
				mutation, err = service.PlanCreate(target.Path, target.content)
			} else {
				mutation, err = service.Create(target.Path, target.content)
			}
			return map[string]any{"target": target, "result": mutationResponse(vault, mutation, createFlags.DryRun, nil)}, err
		})
	}
	create.Flags().StringVar(&createDate, "date", "", "calendar date in YYYY-MM-DD (defaults to today)")
	bindCommonFlags(create, &createFlags, commonFlagSet{DryRun: true}, false)
	command.AddCommand(create)

	var appendDate, appendInput, appendSection string
	var appendFlags commonFlags
	append := &cobra.Command{Use: "append", Args: namespaceArgs(&common, "daily.append", cobra.NoArgs)}
	append.RunE = func(cmd *cobra.Command, _ []string) error {
		return renderEnvelope(cmd, "daily.append", common.RequestID, func() (any, error) {
			content, err := readNoteInput(cmd, appendInput, "content-file")
			if err != nil {
				return nil, err
			}
			service, vault, err := resolveNoteContext(common, registryFactory, serviceFactory)
			if err != nil {
				return nil, err
			}
			target, err := resolve(service, vault.Path, appendDate, false)
			if err != nil {
				return nil, err
			}
			var mutation noteops.Mutation
			if appendFlags.DryRun {
				mutation, err = service.PlanAppend(target.Path, content, appendSection, appendFlags.IfMatch)
			} else {
				mutation, err = service.Append(target.Path, content, appendSection, appendFlags.IfMatch)
			}
			return map[string]any{"target": target, "result": mutationResponse(vault, mutation, appendFlags.DryRun, nil)}, err
		})
	}
	append.Flags().StringVar(&appendDate, "date", "", "calendar date in YYYY-MM-DD (defaults to today)")
	append.Flags().StringVar(&appendInput, "content-file", "", "Markdown input file, or - for stdin")
	append.Flags().StringVar(&appendSection, "section", "", "exact ATX heading title to append within")
	bindCommonFlags(append, &appendFlags, commonFlagSet{DryRun: true, IfMatch: true}, false)
	command.AddCommand(append)
	return command
}
