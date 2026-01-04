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

		cleanup, err := exec.TrackExperiment("mem", "load", map[string]string{
			"bytes":   fmt.Sprintf("%d", sizeBytes),
			"percent": fmt.Sprintf("%.2f", memPercent),
		})
		if err != nil {
			return err
		}
		defer cleanup()

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
}
