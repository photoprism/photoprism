package config

import (
	"os"
)

// Vars represents a map of variable names to values.
type Vars = map[string]string

// ExpandVars replaces variables in the format ${NAME} with their corresponding values.
func ExpandVars(s string, vars Vars) string {
	if s == "" {
		return s
	}

	return os.Expand(s, func(key string) string { return vars[key] })
}
