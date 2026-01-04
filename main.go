package main

import (
	"chaosblade-win/cmd"
	"fmt"
	"os"
)

func main() {
	// Ensure process is elevated; if a relaunch was attempted, exit current process.
	if cmd.EnsureElevated() {
		return
	}
	// Optional debug: print args and cwd when CHAOS_DEBUG=1
	if os.Getenv("CHAOS_DEBUG") == "1" {
		cwd, _ := os.Getwd()
		fmt.Fprintf(os.Stderr, "DEBUG: args=%v cwd=%s\n", os.Args, cwd)
	}
	cmd.Execute()
}
