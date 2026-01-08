package main

import (
	"chaosblade-win/cmd"
	"fmt"
	"os"
)

func main() {
	// Ensure process is elevated; if a relaunch was attempted, exit current process.
	// if cmd.EnsureElevated() {
	// 	return
	// }
	// Optional debug: print args and cwd when CHAOS_DEBUG=1
	if os.Getenv("CHAOS_DEBUG") == "1" {
		cwd, _ := os.Getwd()
		fmt.Fprintf(os.Stderr, "DEBUG: args=%v cwd=%s\n", os.Args, cwd)
	}

	// Helpful warning: if the user accidentally places `--` after `go run .`,
	// some shells or the go tool may not forward arguments as expected. Print
	// a short note to help diagnose this common issue.
	for _, a := range os.Args {
		if a == "--" {
			fmt.Fprintln(os.Stderr, "Warning: detected '--' in args. Prefer 'go run . create ...' instead of 'go run . -- create ...' to avoid argument forwarding issues.")
			break
		}
	}

	// If `--` appears as a standalone argument from shells/`go run`, remove it
	// before passing to Cobra so subcommands are parsed consistently.
	cleanArgs := make([]string, 0, len(os.Args))
	for _, a := range os.Args {
		if a == "--" {
			continue
		}
		cleanArgs = append(cleanArgs, a)
	}
	// Replace os.Args so Cobra receives the cleaned argument list.
	os.Args = cleanArgs
	cmd.Execute()
}
