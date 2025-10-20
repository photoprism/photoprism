package clean

import "strings"

// EnvVar returns the environment variable name for the specified flag.
func EnvVar(flag string) string {
	return "PHOTOPRISM_" + strings.ToUpper(strings.ReplaceAll(flag, "-", "_"))
}

// EnvVars returns the environment variable names for the specified flags.
func EnvVars(flags ...string) []string {
	vars := make([]string, len(flags))

	for i, flag := range flags {
		vars[i] = EnvVar(flag)
	}

	return vars
}
