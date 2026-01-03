package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"chaosblade-win/exec"
	"chaosblade-win/spec"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/spf13/cobra"
)

var diskSizeMB int64
var diskPath string
var diskPercent float64

var diskTargetSpec = spec.Registry["disk"]

var diskCmd = &cobra.Command{
	Use:   "disk",
	Short: diskTargetSpec.Short,
}

var diskFillCmd = &cobra.Command{
	Use:   "fill",
	Short: spec.MustActionSpec("disk", "fill").Short,
	Long:  spec.MustActionSpec("disk", "fill").Long,
	RunE: func(cmd *cobra.Command, args []string) error {
		sizeBytes := diskSizeMB * 1024 * 1024

		usagePath := diskPath
		if usagePath == "" {
			usagePath = os.TempDir()
		} else {
			usagePath = filepath.Dir(usagePath)
		}

		if diskPercent > 0 {
			stats, err := disk.Usage(usagePath)
			if err != nil {
				return fmt.Errorf("query disk usage: %w", err)
			}
			sizeBytes = int64(float64(stats.Total) * diskPercent / 100)

			const safetyMargin = 64 << 20 // 64 MB
			maxBytes := int64(0)
			if stats.Free > safetyMargin {
				maxBytes = int64(stats.Free - safetyMargin)
			}
			if maxBytes > 0 && sizeBytes > maxBytes {
				sizeBytes = maxBytes
			}
		}

		cleanup, err := exec.TrackExperiment("disk", "fill", map[string]string{
			"bytes":   fmt.Sprintf("%d", sizeBytes),
			"path":    diskPath,
			"percent": fmt.Sprintf("%.2f", diskPercent),
		})
		if err != nil {
			return err
		}
		defer cleanup()

		runner := exec.NewDiskFillRunner(diskPath, sizeBytes)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		target := diskPath
		if target == "" {
			target = "temporary file"
		}

		fmt.Printf("Writing ~%.1f MB to %s. Press Ctrl+C to stop.\n", float64(sizeBytes)/1024.0/1024.0, target)
		if err := runner.Run(ctx); err != nil && err != context.Canceled {
			return err
		}
		fmt.Println("Disk fill stopped and cleaned up.")
		return nil
	},
}

func init() {
	createCmd.AddCommand(diskCmd)
	diskCmd.AddCommand(diskFillCmd)
	diskFillCmd.Flags().Int64Var(&diskSizeMB, "size", 512, "Data size to write in MB")
	diskFillCmd.Flags().StringVar(&diskPath, "path", "", "Target file path (defaults to temp file)")
	diskFillCmd.Flags().Float64Var(&diskPercent, "percent", 0, "Data to write as percent of disk total (overrides size if >0)")
}
