package main

import (
	"os"
	"testing"
)

func TestMainDelegatesExitCode(t *testing.T) {
	originalExit, originalArgs := exitProcess, os.Args
	t.Cleanup(func() {
		exitProcess = originalExit
		os.Args = originalArgs
	})
	exitCode := -1
	exitProcess = func(code int) { exitCode = code }
	os.Args = []string{"obs-cli", "--version"}
	main()
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
}
