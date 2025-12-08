package config

import (
	"os"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/txt"
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

// EnvVar returns the environment variable that backs the given CLI flag name.
func EnvVar(flag string) string {
	return clean.EnvVar(flag)
}

// EnvVars converts a list of flag names to their corresponding environment variables.
func EnvVars(flags ...string) (vars []string) {
	return clean.EnvVars(flags...)
}

// Env reports whether any of the provided boolean flags are set via environment
// variable or CLI switch, before configuration files are processed.
func Env(vars ...string) bool {
	for _, s := range vars {
		if (txt.Bool(os.Getenv(EnvVar(s))) || list.Contains(os.Args, "--"+s)) &&
			!list.Contains(os.Args, "--"+s+"=false") {
			return true
		}
	}

	return false
}

// FlagFileVar returns the environment variable that contains a file path for a flag.
func FlagFileVar(flag string) string {
	return EnvVar(flag) + "_FILE"
}

// FlagFilePath resolves the path provided via the *_FILE environment variable for a flag.
func FlagFilePath(flag string) string {
	if envVar := os.Getenv(FlagFileVar(flag)); envVar == "" {
		return ""
	} else if absName := fs.Abs(envVar); fs.FileExistsNotEmpty(absName) {
		return absName
	}

	return ""
}
