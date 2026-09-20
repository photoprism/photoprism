package onnx

import (
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureProviderLog redirects the package logger for the duration of the test and returns its
// entries.
func captureProviderLog(t *testing.T) *test.Hook {
	t.Helper()

	orig := log
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	log = logger

	t.Cleanup(func() { log = orig })

	return hook
}

// requireSessionRuntime skips a test when the ONNX Runtime cannot be loaded, which is the case
// in build environments that did not run "make dep".
func requireSessionRuntime(t *testing.T) {
	t.Helper()

	if err := EnsureRuntime(""); err != nil {
		t.Skipf("onnx: skipping, %s", err)
	}
}

func TestProvider_String(t *testing.T) {
	t.Run("CPU", func(t *testing.T) {
		assert.Equal(t, "cpu", ProviderCPU.String())
	})
	t.Run("CUDA", func(t *testing.T) {
		assert.Equal(t, "cuda", ProviderCUDA.String())
	})
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, ProviderCPU, DefaultProvider)
	})
}

func TestParseProvider(t *testing.T) {
	t.Run("CPU", func(t *testing.T) {
		provider, ok := ParseProvider("cpu")
		assert.True(t, ok)
		assert.Equal(t, ProviderCPU, provider)
	})
	t.Run("CUDA", func(t *testing.T) {
		provider, ok := ParseProvider("cuda")
		assert.True(t, ok)
		assert.Equal(t, ProviderCUDA, provider)
	})
	t.Run("Empty", func(t *testing.T) {
		// An unset option is the default, not a mistake, so it must not be reported as one.
		provider, ok := ParseProvider("")
		assert.True(t, ok)
		assert.Equal(t, ProviderCPU, provider)
	})
	t.Run("Whitespace", func(t *testing.T) {
		provider, ok := ParseProvider("  CUDA  ")
		assert.True(t, ok)
		assert.Equal(t, ProviderCUDA, provider)
	})
	t.Run("Unknown", func(t *testing.T) {
		provider, ok := ParseProvider("rocm")
		assert.False(t, ok)
		assert.Equal(t, ProviderCPU, provider)
	})
}

func TestNewSessionOptions(t *testing.T) {
	t.Run("CPU", func(t *testing.T) {
		requireSessionRuntime(t)
		hook := captureProviderLog(t)

		opts, applied, err := NewSessionOptions(SessionSettings{Provider: ProviderCPU, IntraOpThreads: 2, InterOpThreads: 1})
		require.NoError(t, err)
		require.NotNil(t, opts)
		t.Cleanup(func() { DestroySessionOptions(opts) })
		assert.Equal(t, ProviderCPU, applied)
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("DefaultThreads", func(t *testing.T) {
		// Zero leaves the runtime's own defaults, so an unconfigured caller still gets options.
		requireSessionRuntime(t)

		opts, applied, err := NewSessionOptions(SessionSettings{})
		require.NoError(t, err)
		require.NotNil(t, opts)
		t.Cleanup(func() { DestroySessionOptions(opts) })
		assert.Equal(t, ProviderCPU, applied)
	})
	t.Run("CUDAFallsBack", func(t *testing.T) {
		// Proves the fall back on every machine, including one where the provider works.
		requireSessionRuntime(t)
		hook := captureProviderLog(t)

		orig := appendProviderVar
		t.Cleanup(func() { appendProviderVar = orig })
		appendProviderVar = func(*onnxruntime.SessionOptions) error {
			return errors.New("no CUDA-capable device is detected")
		}

		opts, applied, err := NewSessionOptions(SessionSettings{Provider: ProviderCUDA, IntraOpThreads: 2, InterOpThreads: 1})
		require.NoError(t, err)
		require.NotNil(t, opts)
		t.Cleanup(func() { DestroySessionOptions(opts) })
		assert.Equal(t, ProviderCPU, applied)

		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Equal(t, logrus.WarnLevel, entry.Level)
		assert.Contains(t, entry.Message, "cuda")
		assert.Contains(t, entry.Message, "no CUDA-capable device is detected")
		// The runtime prints its own red error line first, so ours must say it is not fatal.
		assert.Contains(t, entry.Message, "running inference on the cpu")
		assert.Len(t, hook.AllEntries(), 1)
	})
	t.Run("CUDAApplied", func(t *testing.T) {
		requireSessionRuntime(t)

		orig := appendProviderVar
		t.Cleanup(func() { appendProviderVar = orig })
		appendProviderVar = func(*onnxruntime.SessionOptions) error { return nil }

		opts, applied, err := NewSessionOptions(SessionSettings{Provider: ProviderCUDA, IntraOpThreads: 2})
		require.NoError(t, err)
		require.NotNil(t, opts)
		t.Cleanup(func() { DestroySessionOptions(opts) })
		assert.Equal(t, ProviderCUDA, applied)
	})
}

func TestNewBaseSessionOptions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		requireSessionRuntime(t)

		opts, err := newBaseSessionOptions(SessionSettings{IntraOpThreads: 2, InterOpThreads: 1})
		require.NoError(t, err)
		require.NotNil(t, opts)
		DestroySessionOptions(opts)
	})
	t.Run("InvalidThreads", func(t *testing.T) {
		// A negative count is left to the runtime rather than passed on, so no error is
		// expected here; the options must still be usable.
		requireSessionRuntime(t)

		opts, err := newBaseSessionOptions(SessionSettings{IntraOpThreads: -4, InterOpThreads: -1})
		require.NoError(t, err)
		require.NotNil(t, opts)
		DestroySessionOptions(opts)
	})
}

func TestAppendCUDAProvider(t *testing.T) {
	t.Run("ReportsAvailability", func(t *testing.T) {
		// Both outcomes are valid: a CPU-only build reports the provider as unavailable, and a
		// GPU build with a visible device appends it. Neither may panic or leak the options.
		requireSessionRuntime(t)

		opts, err := newBaseSessionOptions(SessionSettings{IntraOpThreads: 1})
		require.NoError(t, err)
		t.Cleanup(func() { DestroySessionOptions(opts) })

		if err = appendCUDAProvider(opts); err != nil {
			t.Logf("cuda provider unavailable: %s", err)
			assert.NotEmpty(t, err.Error())
		}
	})
}

func TestDestroySessionOptions(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.NotPanics(t, func() { DestroySessionOptions(nil) })
	})
	t.Run("Success", func(t *testing.T) {
		requireSessionRuntime(t)
		hook := captureProviderLog(t)

		opts, err := newBaseSessionOptions(SessionSettings{IntraOpThreads: 1})
		require.NoError(t, err)

		DestroySessionOptions(opts)
		assert.Empty(t, hook.AllEntries())
	})
}
