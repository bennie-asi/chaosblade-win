package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"chaosblade-win/exec"
	"chaosblade-win/spec"

	"github.com/shirou/gopsutil/v4/mem"
	"github.com/spf13/cobra"
)

var memSizeMB int64
var memPercent float64
var memDetach bool
var memDetachedChild bool

var memTargetSpec = spec.Registry["mem"]

var memCmd = &cobra.Command{
	Use:   "mem",
	Short: memTargetSpec.Short,
}

var memLoadCmd = &cobra.Command{
	Use:   "load",
	Short: spec.MustActionSpec("mem", "load").Short,
	Long:  spec.MustActionSpec("mem", "load").Long,
	RunE: func(cmd *cobra.Command, args []string) error {
		sizeBytes := memSizeMB * 1024 * 1024
		if memPercent > 0 {
			stats, err := mem.VirtualMemory()
			if err != nil {
				return fmt.Errorf("query memory: %w", err)
			}
			sizeBytes = int64(float64(stats.Total) * memPercent / 100)
		}

		if memDetach && !memDetachedChild {
			args := []string{"create", "mem", "load", "--size", fmt.Sprintf("%d", memSizeMB), "--detached-child"}
			pid, err := exec.StartDetachedExperiment(args)
			if err != nil {
				return err
			}
			fmt.Printf("Started detached experiment pid=%d\n", pid)
			return nil
		}

		id, cleanup, err := exec.TrackExperiment("mem", "load", map[string]string{
			"bytes":   fmt.Sprintf("%d", sizeBytes),
			"percent": fmt.Sprintf("%.2f", memPercent),
		})
		if err != nil {
			return err
		}
		defer cleanup()
		fmt.Printf("Started experiment id=%s\n", id)

		runner := exec.NewMemoryRunner(sizeBytes)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		fmt.Printf("Allocating ~%.1f MB. Press Ctrl+C to stop.\n", float64(sizeBytes)/1024.0/1024.0)
		if err := runner.Run(ctx); err != nil && err != context.Canceled {
			return err
		}
		fmt.Println("Memory load stopped.")
		return nil
	},
}

func init() {
	createCmd.AddCommand(memCmd)
	memCmd.AddCommand(memLoadCmd)
	mustBindFlags(memLoadCmd, spec.MustActionSpec("mem", "load"), map[string]any{
		"size":    &memSizeMB,
		"percent": &memPercent,
	})
	memLoadCmd.Flags().BoolVar(&memDetach, "detach", false, "run experiment detached (returns immediately)")
	memLoadCmd.Flags().BoolVar(&memDetachedChild, "detached-child", false, "(internal) run as detached child and write state")
	_ = memLoadCmd.Flags().MarkHidden("detached-child")
}
