package cmd

import (
	"fmt"

	"github.com/andy-neoaira/obs-cli/pkg/protocol"
	"github.com/spf13/cobra"
)

type renderedCommandError struct {
	cause error
	exit  int
}

func (e *renderedCommandError) Error() string { return e.cause.Error() }
func (e *renderedCommandError) Unwrap() error { return e.cause }

func renderV2(
	cmd *cobra.Command,
	operation string,
	requestID string,
	run func() (any, error),
) error {
	data, err := run()
	if err == nil {
		return protocol.Render(cmd.OutOrStdout(), protocol.Success(operation, requestID, data, nil))
	}
	domain := protocol.MapError(err)
	if renderErr := protocol.Render(
		cmd.OutOrStdout(),
		protocol.Failure(operation, requestID, domain, nil),
	); renderErr != nil {
		return renderErr
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "[%s] %s: %s\n", requestID, domain.Code, domain.Message)
	return &renderedCommandError{cause: domain, exit: protocol.ExitCodeFor(domain)}
}
