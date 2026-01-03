package cmd

import "github.com/spf13/cobra"

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy chaos experiments",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
