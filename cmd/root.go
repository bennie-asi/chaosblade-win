package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "chaosblade-win",
	Short: "Chaos experiment CLI for Windows",
	Long:  "A lightweight skeleton for chaos experiments on Windows using Cobra.",
}

// Execute runs the root Cobra command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
