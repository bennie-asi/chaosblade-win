package cmd

import "github.com/spf13/cobra"

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create chaos experiments",
}

func init() {
	rootCmd.AddCommand(createCmd)
}
