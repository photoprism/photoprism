package cluster

import (
	"os"
	"strconv"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
)

// BootstrapAutoJoinEnabled indicates whether cluster bootstrap logic is enabled
// for nodes by default. Portal nodes ignore this value; gating is decided by
// runtime checks (e.g., conf.Portal() and conf.NodeRole()).
var BootstrapAutoJoinEnabled = true

// BootstrapAutoThemeEnabled indicates whether bootstrap should attempt to
// download and install a Portal-provided theme when appropriate.
var BootstrapAutoThemeEnabled = true

// BootstrapRegisterMaxAttempts defines how many attempts the bootstrap logic
// makes when contacting the Portal for registration before giving up.
var BootstrapRegisterMaxAttempts = 6

// BootstrapRegisterRetryDelay defines the delay between registration attempts
// when the Portal is temporarily unavailable.
var BootstrapRegisterRetryDelay = 15 * time.Second

// BootstrapRegisterTimeout defines the HTTP client timeout per registration
// request to the Portal.
var BootstrapRegisterTimeout = 15 * time.Second

func init() {
	applyPolicyEnv()
}

// applyPolicyEnv allows advanced users to fine-tune bootstrap behaviour via environment
// variables without exposing additional user-facing configuration options.
func applyPolicyEnv() {
	if v := os.Getenv(clean.EnvVar("cluster-bootstrap-auto-join-enabled")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			BootstrapAutoJoinEnabled = b
		}
	}

	if v := os.Getenv(clean.EnvVar("cluster-bootstrap-auto-theme-enabled")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			BootstrapAutoThemeEnabled = b
		}
	}

	if v := os.Getenv(clean.EnvVar("cluster-bootstrap-max-attempts")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			BootstrapRegisterMaxAttempts = n
		}
	}

	if v := os.Getenv(clean.EnvVar("cluster-bootstrap-retry-delay")); v != "" {
		if d, ok := parseDurationEnv(v); ok {
			BootstrapRegisterRetryDelay = d
		}
	}

	if v := os.Getenv(clean.EnvVar("cluster-bootstrap-timeout")); v != "" {
		if d, ok := parseDurationEnv(v); ok {
			BootstrapRegisterTimeout = d
		}
	}
}

func parseDurationEnv(value string) (time.Duration, bool) {
	if d, err := time.ParseDuration(value); err == nil {
		return d, true
	}

	if n, err := strconv.Atoi(value); err == nil && n >= 0 {
		return time.Duration(n) * time.Second, true
	}

	return 0, false
}
