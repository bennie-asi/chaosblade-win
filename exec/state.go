package exec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// ExperimentState captures ownership information for a running experiment.
type ExperimentState struct {
	Target    string            `json:"target"`
	Action    string            `json:"action"`
	PID       int               `json:"pid"`
	StartedAt time.Time         `json:"startedAt"`
	Params    map[string]string `json:"params,omitempty"`
}

// ErrExperimentRunning indicates an experiment of the same target is already tracked.
var ErrExperimentRunning = errors.New("experiment already running; destroy it first")

// TrackExperiment records the caller PID as the owner for a target/action combination.
// It fails if an alive process already owns the same target. The returned cleanup removes
// the state when invoked and only when owned by the same PID.
func TrackExperiment(target, action string, params map[string]string) (func(), error) {
	pid := os.Getpid()

	existing, alive, err := loadState(target)
	if err != nil {
		return nil, err
	}
	if alive && existing.PID != pid {
		return nil, fmt.Errorf("%w (pid=%d)", ErrExperimentRunning, existing.PID)
	}

	state := ExperimentState{
		Target:    target,
		Action:    action,
		PID:       pid,
		StartedAt: time.Now().UTC(),
		Params:    params,
	}

	if err := writeState(target, state); err != nil {
		return nil, err
	}

	cleanup := func() {
		_ = clearState(target, pid)
	}
	return cleanup, nil
}

// KillTrackedExperiment terminates the process recorded for a target and clears state.
func KillTrackedExperiment(target string) (*ExperimentState, error) {
	state, alive, err := loadState(target)
	if err != nil {
		return nil, err
	}
	if state.PID == 0 {
		return nil, fmt.Errorf("no tracked %s experiment", target)
	}

	if !alive {
		// stale record; clean and return
		_ = clearState(target, 0)
		return nil, fmt.Errorf("no active %s experiment (stale record removed)", target)
	}

	proc, err := os.FindProcess(state.PID)
	if err != nil {
		return nil, fmt.Errorf("find process %d: %w", state.PID, err)
	}
	if err := proc.Kill(); err != nil {
		return nil, fmt.Errorf("terminate process %d: %w", state.PID, err)
	}

	_ = clearState(target, state.PID)
	return &state, nil
}

// loadState reads the tracked state and indicates whether the referenced process is alive.
func loadState(target string) (ExperimentState, bool, error) {
	var state ExperimentState

	data, err := os.ReadFile(stateFile(target))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state, false, nil
		}
		return state, false, err
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return state, false, err
	}

	if state.PID == 0 {
		return state, false, nil
	}

	return state, isProcessAlive(state.PID), nil
}

func writeState(target string, state ExperimentState) error {
	dir := filepath.Join(os.TempDir(), "chaosblade-win")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile(target), data, 0o644)
}

func clearState(target string, ownerPID int) error {
	path := stateFile(target)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	// Respect ownership if present.
	if ownerPID != 0 {
		var state ExperimentState
		if err := json.Unmarshal(data, &state); err == nil && state.PID != 0 && state.PID != ownerPID {
			// Not owned by caller; keep record.
			return nil
		}
	}

	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func stateFile(target string) string {
	return filepath.Join(os.TempDir(), "chaosblade-win", fmt.Sprintf("%s.json", target))
}

func isProcessAlive(pid int) bool {
	const (
		processQueryLimitedInformation = 0x1000 // matches Windows PROCESS_QUERY_LIMITED_INFORMATION
		stillActive                    = 259
	)

	if pid <= 0 {
		return false
	}

	handle, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	var code uint32
	if err := syscall.GetExitCodeProcess(handle, &code); err != nil {
		return false
	}

	return code == stillActive
}
