package cmd

import (
	"fmt"
	"runtime/debug"

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

// rootCmd 是 obs-cli 的根命令定义。
// Cobra 框架会根据这里的定义自动生成帮助信息、版本号和子命令树。
var rootCmd = &cobra.Command{
	Use:           "obs-cli",
	Short:         "Interact with Obsidian vaults from the terminal",
	Version:       resolveVersion(),
	Long:          "Interact with Obsidian vaults from the terminal",
	SilenceErrors: true,
	SilenceUsage:  true,
}

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
