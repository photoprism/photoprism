package onnx

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedLibraryCandidates(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		t.Cleanup(func() { executableVar = os.Executable })
		executableVar = func() (string, error) {
			return filepath.Join("/opt/photoprism", "bin", "photoprism"), nil
		}

		candidates := SharedLibraryCandidates("")
		require.NotEmpty(t, candidates)
		require.Equal(t, "libonnxruntime.so", candidates[0])
		require.Contains(t, candidates, filepath.Join("/opt/photoprism", "lib", "libonnxruntime.so"))
	})
	t.Run("ExplicitFirst", func(t *testing.T) {
		t.Cleanup(func() { executableVar = os.Executable })
		executableVar = func() (string, error) { return "/tmp/photoprism", nil }

		explicit := "/custom/libonnxruntime.so"
		candidates := SharedLibraryCandidates(explicit)
		require.NotEmpty(t, candidates)
		require.Equal(t, explicit, candidates[0])
	})
	t.Run("NoExecutable", func(t *testing.T) {
		t.Cleanup(func() { executableVar = os.Executable })
		executableVar = func() (string, error) { return "", os.ErrNotExist }

		candidates := SharedLibraryCandidates("")
		require.Equal(t, []string{"libonnxruntime.so", "libonnxruntime.so.1", "onnxruntime.so"}, candidates)
	})
	t.Run("Deduplicates", func(t *testing.T) {
		t.Cleanup(func() { executableVar = os.Executable })
		executableVar = func() (string, error) { return "", os.ErrNotExist }

		candidates := SharedLibraryCandidates("libonnxruntime.so")
		require.Equal(t, []string{"libonnxruntime.so", "libonnxruntime.so.1", "onnxruntime.so"}, candidates)
	})
}

func TestEnsureRuntime(t *testing.T) {
	t.Run("Idempotent", func(t *testing.T) {
		// The result is cached, so a second call must report exactly what the first did
		// rather than retrying the candidate list.
		first := EnsureRuntime("")
		require.Equal(t, first, EnsureRuntime(""))
		if first == nil {
			require.NotEmpty(t, RuntimeLibraryPath())
		}
	})
}

// stubRuntime replaces the native runtime calls loadRuntime makes and records the ones it receives.
func stubRuntime(t *testing.T, loadable string, disableErr error) *[]string {
	t.Helper()

	setPath, initEnv, disable, destroy := setLibraryPathVar, initEnvironmentVar, disableTelemetryVar, destroyEnvironmentVar
	t.Cleanup(func() {
		setLibraryPathVar, initEnvironmentVar, disableTelemetryVar, destroyEnvironmentVar = setPath, initEnv, disable, destroy
	})

	calls := &[]string{}
	current := ""

	setLibraryPathVar = func(path string) { current = path }
	initEnvironmentVar = func(...onnxruntime.EnvironmentOption) error {
		*calls = append(*calls, "init:"+current)
		if current != loadable {
			return errors.New("not found")
		}
		return nil
	}
	disableTelemetryVar = func() error {
		*calls = append(*calls, "disable")
		return disableErr
	}
	destroyEnvironmentVar = func() error {
		*calls = append(*calls, "destroy")
		return nil
	}

	return calls
}

func TestLoadRuntime(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Setenv(TelemetryEnv, "0")
		calls := stubRuntime(t, "b.so", nil)

		path, err := loadRuntime([]string{"a.so", "b.so", "c.so"})
		require.NoError(t, err)
		assert.Equal(t, "b.so", path)
		assert.Equal(t, []string{"init:a.so", "init:b.so", "disable"}, *calls)
		assert.Equal(t, "1", os.Getenv(TelemetryEnv))
	})
	t.Run("DisableFailed", func(t *testing.T) {
		calls := stubRuntime(t, "a.so", errors.New("unsupported"))

		path, err := loadRuntime([]string{"a.so", "b.so"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "telemetry")
		assert.Empty(t, path)
		assert.Equal(t, []string{"init:a.so", "disable", "destroy"}, *calls)
	})
	t.Run("NotFound", func(t *testing.T) {
		calls := stubRuntime(t, "", nil)

		path, err := loadRuntime([]string{"a.so", "b.so"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "a.so (not found); b.so (not found)")
		assert.Empty(t, path)
		assert.Equal(t, []string{"init:a.so", "init:b.so"}, *calls)
	})
	t.Run("NoCandidates", func(t *testing.T) {
		calls := stubRuntime(t, "", nil)

		path, err := loadRuntime(nil)
		require.EqualError(t, err, "no ONNX runtime library candidates")
		assert.Empty(t, path)
		assert.Empty(t, *calls)
	})
}
