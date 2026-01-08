package cmd

import (
	"fmt"

	"chaosblade-win/exec"

	"github.com/spf13/cobra"
)

var destroyCpuCmd = &cobra.Command{
	Use:   "cpu [id]",
	Short: "Stop a running CPU experiment (optionally by id)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := ""
		if len(args) == 1 {
			id = args[0]
		}
		state, err := exec.KillTrackedExperiment("cpu", id)
		if err != nil {
			return err
		}
		if state != nil {
			fmt.Printf("Stopped CPU experiment id=%s pid=%d.\n", state.ID, state.PID)
		}
		return nil
	},
}

var destroyMemCmd = &cobra.Command{
	Use:   "mem [id]",
	Short: "Stop a running memory experiment (optionally by id)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := ""
		if len(args) == 1 {
			id = args[0]
		}
		state, err := exec.KillTrackedExperiment("mem", id)
		if err != nil {
			return err
		}
		if state != nil {
			fmt.Printf("Stopped memory experiment id=%s pid=%d.\n", state.ID, state.PID)
		}
		return nil
	},
}

var destroyDiskCmd = &cobra.Command{
	Use:   "disk [id]",
	Short: "Stop a running disk experiment (optionally by id)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := ""
		if len(args) == 1 {
			id = args[0]
		}
		state, err := exec.KillTrackedExperiment("disk", id)
		if err != nil {
			return err
		}
		if state != nil {
			fmt.Printf("Stopped disk experiment id=%s pid=%d.\n", state.ID, state.PID)
		}
		return nil
	},
}

var destroyNetCmd = &cobra.Command{
	Use:   "net [id]",
	Short: "Stop a running network experiment (optionally by id)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := ""
		if len(args) == 1 {
			id = args[0]
		}
		state, err := exec.KillTrackedExperiment("net", id)
		if err != nil {
			return err
		}
		if state != nil {
			fmt.Printf("Stopped network experiment id=%s pid=%d.\n", state.ID, state.PID)
		}
		return nil
	},
}

func init() {
	destroyCmd.AddCommand(destroyCpuCmd)
	destroyCmd.AddCommand(destroyMemCmd)
	destroyCmd.AddCommand(destroyDiskCmd)
	destroyCmd.AddCommand(destroyNetCmd)
}
