package exec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// StartDetachedExperiment launches a new process of the current binary with the
// provided args and returns the child PID and any error. The child process is
// expected to write its own state (TrackExperiment) when it starts.
func StartDetachedExperiment(args []string) (int, error) {
	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("resolve executable: %w", err)
	}

	// Make sure the path is absolute for CreateProcess.
	exe, err = filepath.Abs(exe)
	if err != nil {
		return 0, fmt.Errorf("abs executable: %w", err)
	}

	cmd := exec.Command(exe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Detach: on Windows, Start will create a new process; we avoid inheriting
	// console handles explicitly by clearing SysProcAttr here if needed later.

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start detached process: %w", err)
	}

	return cmd.Process.Pid, nil
}
