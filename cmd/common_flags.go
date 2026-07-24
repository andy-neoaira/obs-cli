package cmd

import "github.com/spf13/cobra"

type commonFlags struct {
	Output    string
	RequestID string
	DryRun    bool
	IfMatch   string
	Vault     string
}

type commonFlagSet struct {
	Output    bool
	RequestID bool
	DryRun    bool
	IfMatch   bool
	Vault     bool
}

func bindCommonFlags(command *cobra.Command, values *commonFlags, enabled commonFlagSet, persistent bool) {
	flags := command.Flags()
	if persistent {
		flags = command.PersistentFlags()
	}
	if enabled.Output {
		flags.StringVar(&values.Output, "output", "json", "output format (json)")
	}
	if enabled.RequestID {
		flags.StringVar(&values.RequestID, "request-id", "", "caller-provided request identifier")
	}
	if enabled.DryRun {
		flags.BoolVar(&values.DryRun, "dry-run", false, "return a change plan without writing")
	}
	if enabled.IfMatch {
		flags.StringVar(&values.IfMatch, "if-match", "", "required current revision")
	}
	if enabled.Vault {
		flags.StringVar(&values.Vault, "vault", "", "vault id or name")
	}
}
