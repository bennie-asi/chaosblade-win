package spec

import (
	"runtime"
	"time"
)

// Registry holds built-in target/action specifications used by the CLI.
var Registry = map[string]TargetSpec{
	"cpu": {
		Name:  "cpu",
		Short: "CPU experiments",
		Actions: map[string]ActionSpec{
			"load": {
				Target: "cpu",
				Name:   "load",
				Short:  "Run a CPU load with optional percent/duration",
				Long:   "Stress CPU cores with a target utilization percent and optional duration.",
				Flags: []FlagSpec{
					{Name: "cores", Shorthand: "c", Type: "int", Default: runtime.NumCPU(), Usage: "Number of CPU cores to stress"},
					{Name: "percent", Type: "int", Default: 100, Usage: "Approximate CPU utilization percent per core (1-100)"},
					{Name: "duration", Type: "duration", Default: time.Duration(0), Usage: "Optional duration before auto-stop (e.g. 30s, 5m)"},
				},
			},
		},
	},
	"mem": {
		Name:  "mem",
		Short: "Memory experiments",
		Actions: map[string]ActionSpec{
			"load": {
				Target: "mem",
				Name:   "load",
				Short:  "Allocate and hold memory",
				Long:   "Allocates memory by size or percent of total and holds it until stopped.",
				Flags: []FlagSpec{
					{Name: "size", Type: "int64", Default: int64(256), Usage: "Memory to allocate in MB"},
					{Name: "percent", Type: "float", Default: float64(0), Usage: "Memory to allocate as percent of total (overrides size if >0)"},
				},
			},
		},
	},
	"disk": {
		Name:  "disk",
		Short: "Disk experiments",
		Actions: map[string]ActionSpec{
			"fill": {
				Target: "disk",
				Name:   "fill",
				Short:  "Fill disk with temporary data",
				Long:   "Writes data to a file until the requested size/percent is reached, then holds it.",
				Flags: []FlagSpec{
					{Name: "size", Type: "int64", Default: int64(512), Usage: "Data size to write in MB"},
					{Name: "percent", Type: "float", Default: float64(0), Usage: "Data to write as percent of disk total (overrides size if >0)"},
					{Name: "path", Type: "string", Default: "", Usage: "Target file path (defaults to temp file)"},
				},
			},
		},
	},
	"net": {
		Name:  "net",
		Short: "Network experiments",
		Actions: map[string]ActionSpec{
			"delay": {
				Target: "net",
				Name:   "delay",
				Short:  "Inject network delay/loss/bandwidth (WinDivert)",
				Long:   "Shapes traffic with delay, jitter, packet loss, and bandwidth caps using WinDivert.",
				Flags: []FlagSpec{
					{Name: "delay", Type: "int", Default: 100, Usage: "Base one-way delay in ms"},
					{Name: "jitter", Type: "int", Default: 0, Usage: "Jitter in ms"},
					{Name: "loss", Type: "float", Default: 0, Usage: "Packet loss percent (0-100)"},
					{Name: "bandwidth", Type: "int", Default: 0, Usage: "Bandwidth cap in kbps (0 means unlimited)"},
					{Name: "filter", Type: "string", Default: "outbound and tcp", Usage: "WinDivert filter expression (e.g., 'outbound and tcp')"},
				},
			},
		},
	},
}

// ActionSpecFor retrieves an action specification.
func ActionSpecFor(target, action string) (ActionSpec, bool) {
	t, ok := Registry[target]
	if !ok {
		return ActionSpec{}, false
	}
	a, ok := t.Actions[action]
	return a, ok
}

// MustActionSpec panics if the target/action is missing; intended for static wiring.
func MustActionSpec(target, action string) ActionSpec {
	a, ok := ActionSpecFor(target, action)
	if !ok {
		panic("spec: missing action spec for " + target + ":" + action)
	}
	return a
}

// TargetSpecFor retrieves a target specification.
func TargetSpecFor(target string) (TargetSpec, bool) {
	t, ok := Registry[target]
	return t, ok
}
