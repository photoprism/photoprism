package config

import (
	"os"
	"strings"

	"github.com/photoprism/photoprism/pkg/list"
)

// Environment names.
const (
	EnvUnsafe  = "unsafe"
	EnvDebug   = "debug"
	EnvTrace   = "trace"
	EnvDemo    = "demo"
	EnvSponsor = "sponsor"
	EnvTest    = "test"
)

// Env checks the presence of environment and command-line flags.
func Env(vars ...string) bool {
	for _, s := range vars {
		if os.Getenv("PHOTOPRISM_"+strings.ToUpper(s)) == "true" || list.Contains(os.Args, "--"+s) {
			return true
		}
	}

	return false
}
