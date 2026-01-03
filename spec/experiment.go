package spec

// Experiment describes a chaos experiment model.
type Experiment struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Scope       string   `json:"scope"`
	Actions     []Action `json:"actions"`
}

// Action represents an action within an experiment.
type Action struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Target      string `json:"target"`
}

// FlagSpec documents a CLI flag for an action.
type FlagSpec struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"` // string, int, float, duration, bool
	Default   any    `json:"default,omitempty"`
	Usage     string `json:"usage"`
}

// ActionSpec captures metadata for one action.
type ActionSpec struct {
	Target string     `json:"target"`
	Name   string     `json:"name"`
	Short  string     `json:"short"`
	Long   string     `json:"long"`
	Flags  []FlagSpec `json:"flags,omitempty"`
}

// TargetSpec groups actions for a target.
type TargetSpec struct {
	Name    string                `json:"name"`
	Short   string                `json:"short"`
	Actions map[string]ActionSpec `json:"actions"`
}
