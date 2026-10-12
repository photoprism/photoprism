package onnx

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// telemetryChildEnv marks the subprocess TestTelemetryDisabled runs to observe a fresh process.
const telemetryChildEnv = "PHOTOPRISM_TEST_ONNX_TELEMETRY_CHILD"

func TestDisableTelemetryEnv(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		t.Setenv(TelemetryEnv, "")
		require.NoError(t, os.Unsetenv(TelemetryEnv))
		require.NoError(t, disableTelemetryEnv())
		assert.Equal(t, "1", os.Getenv(TelemetryEnv))
	})
	t.Run("Disabled", func(t *testing.T) {
		t.Setenv(TelemetryEnv, "0")
		require.NoError(t, disableTelemetryEnv())
		assert.Equal(t, "1", os.Getenv(TelemetryEnv))
	})
	t.Run("Other", func(t *testing.T) {
		t.Setenv(TelemetryEnv, "true")
		require.NoError(t, disableTelemetryEnv())
		assert.Equal(t, "1", os.Getenv(TelemetryEnv))
	})
	t.Run("AlreadySet", func(t *testing.T) {
		t.Setenv(TelemetryEnv, "1")
		require.NoError(t, disableTelemetryEnv())
		assert.Equal(t, "1", os.Getenv(TelemetryEnv))
	})
}

// TestTelemetryDisabled runs the test binary again with telemetry enabled in its environment,
// since only a new process shows what package initialization and the native runtime do.
func TestTelemetryDisabled(t *testing.T) {
	if os.Getenv(telemetryChildEnv) == "1" {
		// Package initialization must have replaced the value the parent passed.
		require.Equal(t, "1", os.Getenv(TelemetryEnv))
		requireRuntime(t, detectorModelPath)
		_, err := Inspect(detectorModelPath, nil)
		require.NoError(t, err)
		return
	}

	// The runtime keeps its device identifier and event cache below this directory.
	cacheDir := t.TempDir()

	cmd := exec.Command(os.Args[0], "-test.run=^TestTelemetryDisabled$", "-test.v", "-test.count=1") //nolint:gosec // current test binary with fixed arguments
	cmd.Env = append(os.Environ(),
		telemetryChildEnv+"=1",
		TelemetryEnv+"=0",
		"XDG_CACHE_HOME="+cacheDir,
	)

	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	if strings.Contains(string(out), "--- SKIP") {
		t.Skipf("onnx: skipping, runtime or model unavailable:\n%s", out)
	}

	entries, err := os.ReadDir(cacheDir)
	require.NoError(t, err)
	assert.Empty(t, entries, "runtime wrote telemetry state")
}
