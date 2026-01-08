package cmd

import (
	"fmt"
	"time"

	"chaosblade-win/exec"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [target]",
	Short: "List tracked experiments (optionally for a given target)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := ""
		if len(args) == 1 {
			target = args[0]
		}
		var targets []string
		if target != "" {
			targets = []string{target}
		} else {
			// Show common targets
			targets = []string{"cpu", "mem", "disk", "net"}
		}

		for _, t := range targets {
			states, err := exec.ListStates(t)
			if err != nil {
				return err
			}
			if len(states) == 0 {
				fmt.Printf("%s: none\n", t)
				continue
			}
			fmt.Printf("%s:\n", t)
			for _, s := range states {
				alive := "stale"
				if s.PID != 0 && isPIDAlive(s.PID) {
					alive = "alive"
				}
				fmt.Printf("  id=%s pid=%d started=%s status=%s params=%v\n", s.ID, s.PID, s.StartedAt.Format(time.RFC3339), alive, s.Params)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

// isPIDAlive is a small wrapper using exec.isProcessAlive
func isPIDAlive(pid int) bool {
	return exec.IsProcessAlive(pid)
}
