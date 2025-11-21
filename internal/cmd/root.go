package cmd

import (
	"context"
	"io"
	"os/exec"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "volta",
	Short: "Perform operations using HL7v2 messages",
}

func Execute(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	rootCmd.SetArgs(args)
	rootCmd.SetIn(stdin)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)

	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		} else {
			return 1
		}
	}

	return 0
}
