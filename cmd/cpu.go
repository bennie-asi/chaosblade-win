package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"chaosblade-win/exec"
	"chaosblade-win/spec"

	"github.com/spf13/cobra"
)

var cpuCores int
var cpuPercent int
var cpuDuration time.Duration
var cpuDetach bool
var cpuDetachedChild bool

var cpuTargetSpec = spec.Registry["cpu"]

var cpuCmd = &cobra.Command{
	Use:   "cpu",
	Short: cpuTargetSpec.Short,
}

var cpuLoadCmd = &cobra.Command{
	Use:   "load",
	Short: spec.MustActionSpec("cpu", "load").Short,
	Long:  spec.MustActionSpec("cpu", "load").Long,
	RunE: func(cmd *cobra.Command, args []string) error {
		maxCores := runtime.NumCPU()
		cores := cpuCores
		if cores <= 0 || cores > maxCores {
			cores = maxCores
		}

		percent := cpuPercent
		if percent < 1 || percent > 100 {
			return fmt.Errorf("percent must be between 1 and 100")
		}

		if cpuDuration < 0 {
			return fmt.Errorf("duration must be zero or positive")
		}

		// If detach requested and this is the parent (not the child), spawn a child process
		if cpuDetach && !cpuDetachedChild {
			args := []string{"create", "cpu", "load", "--cores", strconv.Itoa(cores), "--percent", strconv.Itoa(percent), "--duration", cpuDuration.String(), "--detached-child"}
			pid, err := exec.StartDetachedExperiment(args)
			if err != nil {
				return err
			}
			// child will write its own state; parent returns immediately
			fmt.Printf("Started detached experiment pid=%d\n", pid)
			return nil
		}

		id, cleanup, err := exec.TrackExperiment("cpu", "load", map[string]string{
			"cores":    strconv.Itoa(cores),
			"percent":  strconv.Itoa(percent),
			"duration": cpuDuration.String(),
		})
		if err != nil {
			return err
		}
		defer cleanup()
		fmt.Printf("Started experiment id=%s\n", id)

		runner := exec.NewCPURunner(cores, percent, cpuDuration)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		fmt.Printf("Starting CPU load on %d core(s) at %d%%. Press Ctrl+C to stop.\n", cores, percent)

		if err := runner.Run(ctx); err != nil && err != context.Canceled {
			return err
		}

		fmt.Println("CPU load stopped.")
		return nil
	},
}

func init() {
	createCmd.AddCommand(cpuCmd)
	cpuCmd.AddCommand(cpuLoadCmd)
	mustBindFlags(cpuLoadCmd, spec.MustActionSpec("cpu", "load"), map[string]any{
		"cores":    &cpuCores,
		"percent":  &cpuPercent,
		"duration": &cpuDuration,
	})
	cpuLoadCmd.Flags().BoolVar(&cpuDetach, "detach", false, "run experiment detached (returns immediately)")
	cpuLoadCmd.Flags().BoolVar(&cpuDetachedChild, "detached-child", false, "(internal) run as detached child and write state")
	_ = cpuLoadCmd.Flags().MarkHidden("detached-child")
}
