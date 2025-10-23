package cluster

import (
	"os"
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
)

func TestApplyPolicyEnv(t *testing.T) {
	originalAutoJoin := BootstrapAutoJoinEnabled
	originalAutoTheme := BootstrapAutoThemeEnabled
	originalAttempts := BootstrapRegisterMaxAttempts
	originalDelay := BootstrapRegisterRetryDelay
	originalTimeout := BootstrapRegisterTimeout

	t.Cleanup(func() {
		BootstrapAutoJoinEnabled = originalAutoJoin
		BootstrapAutoThemeEnabled = originalAutoTheme
		BootstrapRegisterMaxAttempts = originalAttempts
		BootstrapRegisterRetryDelay = originalDelay
		BootstrapRegisterTimeout = originalTimeout
	})

	t.Setenv(clean.EnvVar("cluster-bootstrap-auto-join-enabled"), "false")
	t.Setenv(clean.EnvVar("cluster-bootstrap-auto-theme-enabled"), "0")
	t.Setenv(clean.EnvVar("cluster-bootstrap-max-attempts"), "3")
	t.Setenv(clean.EnvVar("cluster-bootstrap-retry-delay"), "45s")
	t.Setenv(clean.EnvVar("cluster-bootstrap-timeout"), "120")

	applyPolicyEnv()

	if BootstrapAutoJoinEnabled {
		t.Fatalf("expected BootstrapAutoJoinEnabled=false")
	}
	if BootstrapAutoThemeEnabled {
		t.Fatalf("expected BootstrapAutoThemeEnabled=false")
	}
	if BootstrapRegisterMaxAttempts != 3 {
		t.Fatalf("expected BootstrapRegisterMaxAttempts=3, got %d", BootstrapRegisterMaxAttempts)
	}

	expectedDelay := 45 * time.Second
	if BootstrapRegisterRetryDelay != expectedDelay {
		t.Fatalf("expected BootstrapRegisterRetryDelay=%v, got %v", expectedDelay, BootstrapRegisterRetryDelay)
	}

	expectedTimeout := 120 * time.Second
	if BootstrapRegisterTimeout != expectedTimeout {
		t.Fatalf("expected BootstrapRegisterTimeout=%v, got %v", expectedTimeout, BootstrapRegisterTimeout)
	}

	// invalid inputs should leave values unchanged
	t.Setenv(clean.EnvVar("cluster-bootstrap-max-attempts"), "bad")
	t.Setenv(clean.EnvVar("cluster-bootstrap-retry-delay"), "bad")
	t.Setenv(clean.EnvVar("cluster-bootstrap-timeout"), "bad")

	applyPolicyEnv()

	if BootstrapRegisterMaxAttempts != 3 {
		t.Fatalf("expected BootstrapRegisterMaxAttempts to remain 3 after invalid override, got %d", BootstrapRegisterMaxAttempts)
	}
	if BootstrapRegisterRetryDelay != expectedDelay {
		t.Fatalf("expected BootstrapRegisterRetryDelay to remain %v after invalid override, got %v", expectedDelay, BootstrapRegisterRetryDelay)
	}
	if BootstrapRegisterTimeout != expectedTimeout {
		t.Fatalf("expected BootstrapRegisterTimeout to remain %v after invalid override, got %v", expectedTimeout, BootstrapRegisterTimeout)
	}
}

func TestParseDurationEnvFallbackSeconds(t *testing.T) {
	d, ok := parseDurationEnv("30")
	if !ok {
		t.Fatalf("expected ok for numeric seconds input")
	}
	if d != 30*time.Second {
		t.Fatalf("expected 30s, got %v", d)
	}

	if _, ok := parseDurationEnv("not-a-duration"); ok {
		t.Fatalf("expected parseDurationEnv to fail for invalid input")
	}

	// ensure non-negative check allows zero
	d, ok = parseDurationEnv("0")
	if !ok || d != 0 {
		t.Fatalf("expected zero duration allowed, got %v, ok=%v", d, ok)
	}

	// ensure valid duration string works
	d, ok = parseDurationEnv("1m30s")
	if !ok || d != time.Minute+30*time.Second {
		t.Fatalf("expected 1m30s, got %v, ok=%v", d, ok)
	}
}

func TestApplyPolicyEnvNoEnvSet(t *testing.T) {
	// Clear the relevant environment variables to ensure defaults remain unchanged.
	vars := []string{
		clean.EnvVar("cluster-bootstrap-auto-join-enabled"),
		clean.EnvVar("cluster-bootstrap-auto-theme-enabled"),
		clean.EnvVar("cluster-bootstrap-max-attempts"),
		clean.EnvVar("cluster-bootstrap-retry-delay"),
		clean.EnvVar("cluster-bootstrap-timeout"),
	}

	for _, v := range vars {
		os.Unsetenv(v)
	}

	originalAutoJoin := BootstrapAutoJoinEnabled
	originalAutoTheme := BootstrapAutoThemeEnabled
	originalAttempts := BootstrapRegisterMaxAttempts
	originalDelay := BootstrapRegisterRetryDelay
	originalTimeout := BootstrapRegisterTimeout

	applyPolicyEnv()

	if BootstrapAutoJoinEnabled != originalAutoJoin {
		t.Fatalf("expected BootstrapAutoJoinEnabled to remain %v", originalAutoJoin)
	}
	if BootstrapAutoThemeEnabled != originalAutoTheme {
		t.Fatalf("expected BootstrapAutoThemeEnabled to remain %v", originalAutoTheme)
	}
	if BootstrapRegisterMaxAttempts != originalAttempts {
		t.Fatalf("expected BootstrapRegisterMaxAttempts to remain %d", originalAttempts)
	}
	if BootstrapRegisterRetryDelay != originalDelay {
		t.Fatalf("expected BootstrapRegisterRetryDelay to remain %v", originalDelay)
	}
	if BootstrapRegisterTimeout != originalTimeout {
		t.Fatalf("expected BootstrapRegisterTimeout to remain %v", originalTimeout)
	}
}
