package cmd

import (
	"fmt"

	"chaosblade-win/exec"

	"github.com/spf13/cobra"
)

var destroyCpuCmd = &cobra.Command{
	Use:   "cpu",
	Short: "Stop a running CPU experiment",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := exec.KillTrackedExperiment("cpu")
		if err != nil {
			return err
		}
		if state != nil {
			fmt.Printf("Stopped CPU experiment (pid=%d).\n", state.PID)
		}
		return nil
	},
}

var destroyMemCmd = &cobra.Command{
	Use:   "mem",
	Short: "Stop a running memory experiment",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := exec.KillTrackedExperiment("mem")
		if err != nil {
			return err
		}
		if state != nil {
			fmt.Printf("Stopped memory experiment (pid=%d).\n", state.PID)
		}
		return nil
	},
}

var destroyDiskCmd = &cobra.Command{
	Use:   "disk",
	Short: "Stop a running disk experiment",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := exec.KillTrackedExperiment("disk")
		if err != nil {
			return err
		}
		if state != nil {
			fmt.Printf("Stopped disk experiment (pid=%d).\n", state.PID)
		}
		return nil
	},
}

var destroyNetCmd = &cobra.Command{
	Use:   "net",
	Short: "Stop a running network experiment",
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := exec.KillTrackedExperiment("net")
		if err != nil {
			return err
		}
		if state != nil {
			fmt.Printf("Stopped network experiment (pid=%d).\n", state.PID)
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
