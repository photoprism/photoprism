package config

import (
	"os"
	"strings"

	"github.com/photoprism/photoprism/pkg/list"
)

// Develop indicates whether the application is running in development mode.
var Develop = false

// Environment names.
const (
	EnvProd    = "prod"
	EnvUnsafe  = "unsafe"
	EnvDebug   = "debug"
	EnvTrace   = "trace"
	EnvDemo    = "demo"
	EnvSponsor = "sponsor"
	EnvDevelop = "develop"
	EnvTest    = "test"
)

// EnvVar returns the name of the environment variable for the specified config flag.
func EnvVar(flag string) string {
	return "PHOTOPRISM_" + strings.ToUpper(strings.ReplaceAll(flag, "-", "_"))
}

// EnvVars returns the names of the environment variable for the specified config flag.
func EnvVars(flags ...string) (vars []string) {
	vars = make([]string, len(flags))

	for i, flag := range flags {
		vars[i] = EnvVar(flag)
	}

	return vars
}

// Env checks the presence of environment and command-line flags.
func Env(vars ...string) bool {
	for _, s := range vars {
		if (os.Getenv(EnvVar(s)) == "true" || list.Contains(os.Args, "--"+s)) && !list.Contains(os.Args, "--"+s+"=false") {
			return true
		}
	}

	return false
}
