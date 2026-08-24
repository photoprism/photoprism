package onnx

import (
	"os"
	"path/filepath"
	"testing"

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
		require.Equal(t, EnsureRuntime(""), EnsureRuntime(""))
	})
}
