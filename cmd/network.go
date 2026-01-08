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
var netDelayAction = spec.MustActionSpec("net", "delay")
var netDefaultFilter = stringDefault(netDelayAction.Flags, "filter", "true")
var netDetach bool
var netDetachedChild bool

var netCmd = &cobra.Command{
	Use:   "net",
	Short: netTargetSpec.Short,
}

var netDelayCmd = &cobra.Command{
	Use:     "delay [delay_ms]",
	Short:   netDelayAction.Short,
	Long:    netDelayAction.Long,
	Example: "chaosblade-win create net delay 120 --jitter 20 --loss 1.5 --filter \"outbound and tcp\"",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			v, err := strconv.Atoi(args[0])
			if err != nil || v < 0 {
				return fmt.Errorf("delay must be a non-negative integer milliseconds value")
			}
			netDelayMs = v
		}

		if netDelayMs < 0 || netJitterMs < 0 || netBandwidthKbps < 0 {
			return fmt.Errorf("delay, jitter, and bandwidth must be non-negative")
		}
		if netLossPercent < 0 || netLossPercent > 100 {
			return fmt.Errorf("loss must be between 0 and 100")
		}
		if netFilter == "" {
			netFilter = netDefaultFilter
		}

		runner := exec.NewNetworkDelayRunner(netDelayMs, netJitterMs, netLossPercent, netFilter, netBandwidthKbps)

		if netDetach && !netDetachedChild {
			args := []string{"create", "net", "delay", strconv.Itoa(netDelayMs), "--jitter", strconv.Itoa(netJitterMs), "--loss", fmt.Sprintf("%.2f", netLossPercent), "--filter", netFilter, "--detached-child"}
			pid, err := exec.StartDetachedExperiment(args)
			if err != nil {
				return err
			}
			fmt.Printf("Started detached experiment pid=%d\n", pid)
			return nil
		}

		id, cleanup, err := exec.TrackExperiment("net", "delay", map[string]string{
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
		fmt.Printf("Started experiment id=%s\n", id)

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

	mustBindFlags(netDelayCmd, netDelayAction, map[string]any{
		"delay":     &netDelayMs,
		"jitter":    &netJitterMs,
		"loss":      &netLossPercent,
		"filter":    &netFilter,
		"bandwidth": &netBandwidthKbps,
	})
	netDelayCmd.Flags().BoolVar(&netDetach, "detach", false, "run experiment detached (returns immediately)")
	netDelayCmd.Flags().BoolVar(&netDetachedChild, "detached-child", false, "(internal) run as detached child and write state")
	_ = netDelayCmd.Flags().MarkHidden("detached-child")
}

func stringDefault(flags []spec.FlagSpec, name, fallback string) string {
	for _, f := range flags {
		if f.Name == name {
			if v, ok := f.Default.(string); ok {
				return v
			}
		}
	}
	return fallback
}
