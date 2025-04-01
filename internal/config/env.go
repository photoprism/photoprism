package config

import (
	"os"
	"strings"

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

// Env checks whether the specified boolean command-line or environment flag is set and can be used independently,
// i.e. before the options are initialized with the values found in config files, the environment or CLI flags.
func Env(vars ...string) bool {
	for _, s := range vars {
		if (txt.Bool(os.Getenv(EnvVar(s))) || list.Contains(os.Args, "--"+s)) &&
			!list.Contains(os.Args, "--"+s+"=false") {
			return true
		}
	}

	return false
}

// FlagFileVar returns the name of the environment variable that can contain a filename to load a config value from.
func FlagFileVar(flag string) string {
	return EnvVar(flag) + "_FILE"
}

// FlagFilePath returns the name of the that contains the value of the specified config flag, if any.
func FlagFilePath(flag string) string {
	if envVar := os.Getenv(FlagFileVar(flag)); envVar == "" {
		return ""
	} else if absName := fs.Abs(envVar); fs.FileExistsNotEmpty(absName) {
		return absName
	}

	return ""
}
