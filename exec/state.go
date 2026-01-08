package exec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// ExperimentState captures ownership information for a running experiment.
type ExperimentState struct {
	ID        string            `json:"id"`
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
// TrackExperiment records the caller PID as the owner for a target/action combination.
// It returns the new experiment id, a cleanup function, and an error.
func TrackExperiment(target, action string, params map[string]string) (string, func(), error) {
	pid := os.Getpid()

	// create per-target dir
	dir := filepath.Join(os.TempDir(), "chaosblade-win", target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", nil, err
	}

	id := uuid.New().String()
	state := ExperimentState{
		ID:        id,
		Target:    target,
		Action:    action,
		PID:       pid,
		StartedAt: time.Now().UTC(),
		Params:    params,
	}

	if err := writeStateFileForID(target, id, state); err != nil {
		return "", nil, err
	}

	cleanup := func() {
		_ = clearStateByID(target, id, pid)
	}
	return id, cleanup, nil
}

// KillTrackedExperiment terminates the process recorded for a target and clears state.
// KillTrackedExperiment terminates the process recorded for a target and clears state.
// If id is empty, attempts to stop all experiments for the target.
func KillTrackedExperiment(target, id string) (*ExperimentState, error) {
	if id != "" {
		state, alive, err := loadStateByID(target, id)
		if err != nil {
			return nil, err
		}
		if state.PID == 0 {
			return nil, fmt.Errorf("no tracked %s experiment with id %s", target, id)
		}
		if !alive {
			_ = clearStateByID(target, id, 0)
			return nil, fmt.Errorf("no active %s experiment (stale record removed)", target)
		}
		proc, err := os.FindProcess(state.PID)
		if err != nil {
			return nil, fmt.Errorf("find process %d: %w", state.PID, err)
		}
		if err := proc.Kill(); err != nil {
			return nil, fmt.Errorf("terminate process %d: %w", state.PID, err)
		}
		_ = clearStateByID(target, id, state.PID)
		return &state, nil
	}

	// kill all
	states, err := listStatesForTarget(target)
	if err != nil {
		return nil, err
	}
	var last *ExperimentState
	for _, s := range states {
		if s.PID == 0 {
			_ = clearStateByID(target, s.ID, 0)
			continue
		}
		alive := isProcessAlive(s.PID)
		if !alive {
			_ = clearStateByID(target, s.ID, 0)
			continue
		}
		proc, err := os.FindProcess(s.PID)
		if err == nil {
			_ = proc.Kill()
		}
		_ = clearStateByID(target, s.ID, s.PID)
		last = &s
	}
	if last == nil {
		return nil, fmt.Errorf("no active %s experiment(s)", target)
	}
	return last, nil
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

func stateDirForTarget(target string) string {
	return filepath.Join(os.TempDir(), "chaosblade-win", target)
}

func stateFileForID(target, id string) string {
	return filepath.Join(stateDirForTarget(target), fmt.Sprintf("%s.json", id))
}

func writeStateFileForID(target, id string, state ExperimentState) error {
	dir := stateDirForTarget(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(stateFileForID(target, id), data, 0o644)
}

func loadStateByID(target, id string) (ExperimentState, bool, error) {
	var state ExperimentState
	data, err := os.ReadFile(stateFileForID(target, id))
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

func clearStateByID(target, id string, ownerPID int) error {
	path := stateFileForID(target, id)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if ownerPID != 0 {
		var state ExperimentState
		if err := json.Unmarshal(data, &state); err == nil && state.PID != 0 && state.PID != ownerPID {
			return nil
		}
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func listStatesForTarget(target string) ([]ExperimentState, error) {
	dir := stateDirForTarget(target)
	files, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []ExperimentState
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		path := filepath.Join(dir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var s ExperimentState
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

// ListStates returns all tracked ExperimentState entries for a target.
func ListStates(target string) ([]ExperimentState, error) {
	return listStatesForTarget(target)
}

// IsProcessAlive exposes process liveness check for external callers.
func IsProcessAlive(pid int) bool {
	return isProcessAlive(pid)
}
