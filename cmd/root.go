package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/andy-neoaira/obs-cli/pkg/obsidian"
	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

// ldflagsVersion 由 Makefile / GoReleaser 通过 -ldflags 注入。
// 当使用 go install 安装时，此值为空，版本从 BuildInfo 读取。
var ldflagsVersion string

// resolveVersion 返回 CLI 的版本号。
// 优先级：1) ldflags 注入值 2) runtime/debug BuildInfo 3) "dev"
func resolveVersion() string {
	if ldflagsVersion != "" {
		return ldflagsVersion
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return "dev"
}

func newRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:           "obs-cli",
		Short:         "Agent-first Obsidian Vault operations",
		Version:       resolveVersion(),
		Long:          "A non-interactive, machine-readable execution layer for Obsidian Vaults.",
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	command.AddCommand(
		newCapabilitiesCommand(),
		newVaultCommand(defaultVaultRegistryFactory, obsidian.DiscoverObsidianVaults),
		newNoteCommand(defaultVaultRegistryFactory, defaultNoteServiceFactory),
		newDailyCommand(defaultVaultRegistryFactory, defaultNoteServiceFactory, nil),
		newMetadataCommand(defaultVaultRegistryFactory, defaultNoteServiceFactory),
		newSearchCommand(defaultVaultRegistryFactory, defaultNoteServiceFactory),
		newLinkCommand(defaultVaultRegistryFactory, defaultNoteServiceFactory),
	)
	for _, namespace := range []string{"template", "batch", "doctor"} {
		command.AddCommand(newReservedNamespaceCommand(namespace))
	}
	return command
}

func newReservedNamespaceCommand(name string) *cobra.Command {
	var common commonFlags
	command := &cobra.Command{
		Use:   name,
		Short: "Reserved namespace; inspect capabilities before use",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			requestID, err := protocol.ResolveRequestID(common.RequestID)
			if err != nil {
				requestID, _ = protocol.ResolveRequestID("")
				return renderEnvelope(cmd, name+".status", requestID, func() (any, error) { return nil, err })
			}
			if common.Output != "json" {
				err = protocol.NewError(
					protocol.InvalidArgument,
					"reserved namespaces only support JSON output",
					map[string]any{"field": "output"},
				)
			} else {
				err = protocol.NewError(
					protocol.CapabilityUnsupported,
					"this namespace has no implemented operations",
					map[string]any{"namespace": name, "required": []string{name + ".*"}},
				)
			}
			return renderEnvelope(cmd, name+".status", requestID, func() (any, error) { return nil, err })
		},
	}
	bindCommonFlags(command, &common, commonFlagSet{Output: true, RequestID: true, Vault: true}, false)
	return command
}

var rootCmd = newRootCommand()

// Execute 是 CLI 的入口函数，由 main.go 调用。
// 它会解析命令行参数并执行对应的子命令；如果出错则打印错误信息并退出程序。
func Execute() int {
	return executeRoot(rootCmd)
}

func executeRoot(command *cobra.Command) int {
	if err := command.Execute(); err != nil {
		if rendered, ok := err.(*renderedCommandError); ok {
			return rendered.exit
		}
		fmt.Fprintf(command.ErrOrStderr(), "obs-cli: %s\n", err)
		return 10
	}
	return 0
}
