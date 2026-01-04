package cmd

import (
	"fmt"
	"strconv"
	"time"

	"chaosblade-win/spec"

	"github.com/spf13/cobra"
)

// mustBindFlags connects Cobra flags to action specs and panics on misconfiguration.
func mustBindFlags(cmd *cobra.Command, action spec.ActionSpec, binds map[string]any) {
	if err := bindFlags(cmd, action, binds); err != nil {
		panic(err)
	}
}

// bindFlags declares Cobra flags according to the ActionSpec and binds them to provided pointers.
func bindFlags(cmd *cobra.Command, action spec.ActionSpec, binds map[string]any) error {
	for _, f := range action.Flags {
		target, ok := binds[f.Name]
		if !ok {
			return fmt.Errorf("missing flag binding for %s:%s", action.Target, f.Name)
		}

		switch f.Type {
		case "string":
			ptr, ok := target.(*string)
			if !ok {
				return fmt.Errorf("flag %s binding must be *string", f.Name)
			}
			def, err := toString(f.Default)
			if err != nil {
				return fmt.Errorf("flag %s default: %w", f.Name, err)
			}
			if f.Shorthand != "" {
				cmd.Flags().StringVarP(ptr, f.Name, f.Shorthand, def, f.Usage)
			} else {
				cmd.Flags().StringVar(ptr, f.Name, def, f.Usage)
			}

		case "int":
			ptr, ok := target.(*int)
			if !ok {
				return fmt.Errorf("flag %s binding must be *int", f.Name)
			}
			def, err := toInt(f.Default)
			if err != nil {
				return fmt.Errorf("flag %s default: %w", f.Name, err)
			}
			if f.Shorthand != "" {
				cmd.Flags().IntVarP(ptr, f.Name, f.Shorthand, def, f.Usage)
			} else {
				cmd.Flags().IntVar(ptr, f.Name, def, f.Usage)
			}

		case "int64":
			ptr, ok := target.(*int64)
			if !ok {
				return fmt.Errorf("flag %s binding must be *int64", f.Name)
			}
			def, err := toInt64(f.Default)
			if err != nil {
				return fmt.Errorf("flag %s default: %w", f.Name, err)
			}
			if f.Shorthand != "" {
				cmd.Flags().Int64VarP(ptr, f.Name, f.Shorthand, def, f.Usage)
			} else {
				cmd.Flags().Int64Var(ptr, f.Name, def, f.Usage)
			}

		case "float":
			ptr, ok := target.(*float64)
			if !ok {
				return fmt.Errorf("flag %s binding must be *float64", f.Name)
			}
			def, err := toFloat64(f.Default)
			if err != nil {
				return fmt.Errorf("flag %s default: %w", f.Name, err)
			}
			if f.Shorthand != "" {
				cmd.Flags().Float64VarP(ptr, f.Name, f.Shorthand, def, f.Usage)
			} else {
				cmd.Flags().Float64Var(ptr, f.Name, def, f.Usage)
			}

		case "duration":
			ptr, ok := target.(*time.Duration)
			if !ok {
				return fmt.Errorf("flag %s binding must be *time.Duration", f.Name)
			}
			def, err := toDuration(f.Default)
			if err != nil {
				return fmt.Errorf("flag %s default: %w", f.Name, err)
			}
			if f.Shorthand != "" {
				cmd.Flags().DurationVarP(ptr, f.Name, f.Shorthand, def, f.Usage)
			} else {
				cmd.Flags().DurationVar(ptr, f.Name, def, f.Usage)
			}

		case "bool":
			ptr, ok := target.(*bool)
			if !ok {
				return fmt.Errorf("flag %s binding must be *bool", f.Name)
			}
			def, err := toBool(f.Default)
			if err != nil {
				return fmt.Errorf("flag %s default: %w", f.Name, err)
			}
			if f.Shorthand != "" {
				cmd.Flags().BoolVarP(ptr, f.Name, f.Shorthand, def, f.Usage)
			} else {
				cmd.Flags().BoolVar(ptr, f.Name, def, f.Usage)
			}

		default:
			return fmt.Errorf("unsupported flag type %q for %s", f.Type, f.Name)
		}
	}
	return nil
}

func toString(v any) (string, error) {
	switch t := v.(type) {
	case nil:
		return "", nil
	case string:
		return t, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func toInt(v any) (int, error) {
	switch t := v.(type) {
	case nil:
		return 0, nil
	case int:
		return t, nil
	case int8, int16, int32:
		return int(reflectValueInt64(t)), nil
	case int64:
		return int(t), nil
	case float32:
		return int(t), nil
	case float64:
		return int(t), nil
	case string:
		if t == "" {
			return 0, nil
		}
		i, err := strconv.Atoi(t)
		return i, err
	default:
		return 0, fmt.Errorf("unsupported default type %T", v)
	}
}

func toInt64(v any) (int64, error) {
	switch t := v.(type) {
	case nil:
		return 0, nil
	case int:
		return int64(t), nil
	case int8, int16, int32:
		return reflectValueInt64(t), nil
	case int64:
		return t, nil
	case float32:
		return int64(t), nil
	case float64:
		return int64(t), nil
	case string:
		if t == "" {
			return 0, nil
		}
		i, err := strconv.ParseInt(t, 10, 64)
		return i, err
	default:
		return 0, fmt.Errorf("unsupported default type %T", v)
	}
}

func toFloat64(v any) (float64, error) {
	switch t := v.(type) {
	case nil:
		return 0, nil
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int8, int16, int32, int64:
		return float64(reflectValueInt64(t)), nil
	case string:
		if t == "" {
			return 0, nil
		}
		f, err := strconv.ParseFloat(t, 64)
		return f, err
	default:
		return 0, fmt.Errorf("unsupported default type %T", v)
	}
}

func toDuration(v any) (time.Duration, error) {
	switch t := v.(type) {
	case nil:
		return 0, nil
	case time.Duration:
		return t, nil
	case int:
		return time.Duration(t), nil
	case int64:
		return time.Duration(t), nil
	case string:
		if t == "" {
			return 0, nil
		}
		return time.ParseDuration(t)
	default:
		return 0, fmt.Errorf("unsupported default type %T", v)
	}
}

func toBool(v any) (bool, error) {
	switch t := v.(type) {
	case nil:
		return false, nil
	case bool:
		return t, nil
	case string:
		if t == "" {
			return false, nil
		}
		b, err := strconv.ParseBool(t)
		return b, err
	default:
		return false, fmt.Errorf("unsupported default type %T", v)
	}
}

func reflectValueInt64(v any) int64 {
	switch t := v.(type) {
	case int8:
		return int64(t)
	case int16:
		return int64(t)
	case int32:
		return int64(t)
	default:
		return 0
	}
}
