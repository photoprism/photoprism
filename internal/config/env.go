package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/pkg/list"
)

// Environment names.
const (
	EnvProd    = "prod"
	EnvUnsafe  = "unsafe"
	EnvDebug   = "debug"
	EnvTrace   = "trace"
	EnvDemo    = "demo"
	EnvSponsor = "sponsor"
	EnvTest    = "test"
)

// EnvVar returns the name of the environment variable for the specified config flag.
func EnvVar(flag string) string {
	return "PHOTOPRISM_" + strings.ToUpper(strings.ReplaceAll(flag, "-", "_"))
}

// Env checks the presence of environment and command-line flags.
func Env(vars ...string) bool {
	for _, s := range vars {
		if (envIsTrue(EnvVar(s)) || list.Contains(os.Args, "--"+s)) && !list.Contains(os.Args, "--"+s+"=false") {
			return true
		}
	}

	return false
}

func envIsTrue(env string) bool {
	v := os.Getenv(env)
	b, err := strconv.ParseBool(v)
	if err != nil {
		log.Warnf("Invalid boolean value for env %s, assuming false", env)
	}
	return b
}
