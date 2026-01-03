package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"chaosblade-win/exec"
	"chaosblade-win/spec"

	"github.com/spf13/cobra"
)

var netTargetSpec = spec.Registry["net"]

var netCmd = &cobra.Command{
	Use:   "net",
	Short: netTargetSpec.Short,
}

var netDelayCmd = &cobra.Command{
	Use:   "delay",
	Short: spec.MustActionSpec("net", "delay").Short,
	Long:  spec.MustActionSpec("net", "delay").Long,
	RunE: func(cmd *cobra.Command, args []string) error {
		runner := exec.NewNetworkDelayRunner(netDelayMs, netJitterMs, netLossPercent, netFilter, netBandwidthKbps)

		cleanup, err := exec.TrackExperiment("net", "delay", map[string]string{
			"delay":         strconv.Itoa(netDelayMs),
			"jitter":        strconv.Itoa(netJitterMs),
			"loss":          fmt.Sprintf("%.2f", netLossPercent),
			"filter":        netFilter,
			"bandwidthKbps": strconv.Itoa(netBandwidthKbps),
		})
		if err != nil {
			return err
		}
		defer cleanup()

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		fmt.Printf("Requested net delay=%dms jitter=%dms loss=%.2f%% bandwidth=%dkbps filter=%q. WinDivert must be installed. Press Ctrl+C to stop.\n", netDelayMs, netJitterMs, netLossPercent, netBandwidthKbps, netFilter)
		if err := runner.Run(ctx); err != nil && err != context.Canceled {
			return err
		}
		return nil
	},
}

var (
	netDelayMs       int
	netJitterMs      int
	netLossPercent   float64
	netFilter        string
	netBandwidthKbps int
)

func init() {
	createCmd.AddCommand(netCmd)
	netCmd.AddCommand(netDelayCmd)

	netDelayCmd.Flags().IntVar(&netDelayMs, "delay", 100, "Base one-way delay in ms")
	netDelayCmd.Flags().IntVar(&netJitterMs, "jitter", 0, "Jitter in ms")
	netDelayCmd.Flags().Float64Var(&netLossPercent, "loss", 0, "Packet loss percent (0-100)")
	netDelayCmd.Flags().StringVar(&netFilter, "filter", "true", "WinDivert filter expression (e.g., 'outbound and tcp')")
	netDelayCmd.Flags().IntVar(&netBandwidthKbps, "bandwidth", 0, "Bandwidth cap in kbps (0 means unlimited)")
}
