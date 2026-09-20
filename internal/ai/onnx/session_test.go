package onnx

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// embeddingGraph returns the input and output names of the bundled embedding model, which has
// a fully static 112x112 input geometry and is therefore usable for a warm-up.
func embeddingGraph(t *testing.T) (inputName, outputName string) {
	t.Helper()

	info, err := Inspect(embeddingModelPath, nil)
	require.NoError(t, err)

	return info.Input.Name, info.Output.Name
}

// forceProvider makes the session config report the given provider as applied without needing
// a GPU, so the warm-up and its fall back can be exercised on any machine.
func forceProvider(t *testing.T) {
	t.Helper()

	orig := appendProviderVar
	t.Cleanup(func() { appendProviderVar = orig })
	appendProviderVar = func(*onnxruntime.SessionOptions) error { return nil }
}

func TestNewSessionConfig(t *testing.T) {
	t.Run("CPU", func(t *testing.T) {
		requireSessionRuntime(t)

		cfg, err := NewSessionConfig(SessionSettings{Provider: ProviderCPU, IntraOpThreads: 2, InterOpThreads: 1})
		require.NoError(t, err)
		require.NotNil(t, cfg)
		t.Cleanup(cfg.Destroy)
		assert.Equal(t, ProviderCPU, cfg.Provider)
		assert.NotNil(t, cfg.Options)
	})
	t.Run("CUDAFallsBack", func(t *testing.T) {
		requireSessionRuntime(t)
		captureProviderLog(t)

		orig := appendProviderVar
		t.Cleanup(func() { appendProviderVar = orig })
		appendProviderVar = func(*onnxruntime.SessionOptions) error { return errors.New("not enabled in this build") }

		cfg, err := NewSessionConfig(SessionSettings{Provider: ProviderCUDA, IntraOpThreads: 2})
		require.NoError(t, err)
		require.NotNil(t, cfg)
		t.Cleanup(cfg.Destroy)
		assert.Equal(t, ProviderCPU, cfg.Provider)
	})
}

func TestSessionConfig_Destroy(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var cfg *SessionConfig
		assert.NotPanics(t, cfg.Destroy)
	})
	t.Run("Twice", func(t *testing.T) {
		// Destroy clears the options, so a second call must not free them again.
		requireSessionRuntime(t)

		cfg, err := NewSessionConfig(SessionSettings{IntraOpThreads: 1})
		require.NoError(t, err)

		cfg.Destroy()
		assert.Nil(t, cfg.Options)
		assert.NotPanics(t, cfg.Destroy)
	})
}

func TestSessionConfig_NewSession(t *testing.T) {
	t.Run("CPU", func(t *testing.T) {
		requireRuntime(t, embeddingModelPath)
		hook := captureProviderLog(t)
		inputName, outputName := embeddingGraph(t)

		cfg, err := NewSessionConfig(SessionSettings{Provider: ProviderCPU, IntraOpThreads: 2, InterOpThreads: 1})
		require.NoError(t, err)
		t.Cleanup(cfg.Destroy)

		session, err := cfg.NewSession(embeddingModelPath, []string{inputName}, []string{outputName}, []int64{1, 3, 112, 112})
		require.NoError(t, err)
		require.NotNil(t, session)
		t.Cleanup(func() { DestroySession(session) })
		assert.Equal(t, ProviderCPU, cfg.Provider)
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("WarmUpVerifies", func(t *testing.T) {
		// The provider is reported as applied, so the warm-up runs; it succeeds here and the
		// session must be returned as is.
		requireRuntime(t, embeddingModelPath)
		hook := captureProviderLog(t)
		forceProvider(t)
		inputName, outputName := embeddingGraph(t)

		cfg, err := NewSessionConfig(SessionSettings{Provider: ProviderCUDA, IntraOpThreads: 2})
		require.NoError(t, err)
		t.Cleanup(cfg.Destroy)
		require.Equal(t, ProviderCUDA, cfg.Provider)

		session, err := cfg.NewSession(embeddingModelPath, []string{inputName}, []string{outputName}, []int64{1, 3, 112, 112})
		require.NoError(t, err)
		require.NotNil(t, session)
		t.Cleanup(func() { DestroySession(session) })
		assert.Equal(t, ProviderCUDA, cfg.Provider)
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("WarmUpFailureReloadsOnCPU", func(t *testing.T) {
		// A geometry the graph rejects stands in for a provider that loads but cannot run,
		// which is the failure an incomplete CUDA installation produces at the first node.
		requireRuntime(t, embeddingModelPath)
		hook := captureProviderLog(t)
		forceProvider(t)
		inputName, outputName := embeddingGraph(t)

		cfg, err := NewSessionConfig(SessionSettings{Provider: ProviderCUDA, IntraOpThreads: 2})
		require.NoError(t, err)
		t.Cleanup(cfg.Destroy)

		session, err := cfg.NewSession(embeddingModelPath, []string{inputName}, []string{outputName}, []int64{1, 3, 64, 64})
		require.NoError(t, err)
		require.NotNil(t, session)
		t.Cleanup(func() { DestroySession(session) })

		// The model still loads, and the config stops asking for the provider.
		assert.Equal(t, ProviderCPU, cfg.Provider)
		assert.NotNil(t, cfg.Options)

		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Equal(t, logrus.WarnLevel, entry.Level)
		assert.Contains(t, entry.Message, "cuda")
		assert.Contains(t, entry.Message, "loading it on the cpu")
	})
	t.Run("MissingModel", func(t *testing.T) {
		requireRuntime(t, embeddingModelPath)
		inputName, outputName := embeddingGraph(t)

		cfg, err := NewSessionConfig(SessionSettings{IntraOpThreads: 1})
		require.NoError(t, err)
		t.Cleanup(cfg.Destroy)

		session, err := cfg.NewSession(filepath.Join(t.TempDir(), "absent.onnx"), []string{inputName}, []string{outputName}, []int64{1, 3, 112, 112})
		require.Error(t, err)
		assert.Nil(t, session)
	})
}

func TestWarmUpSession(t *testing.T) {
	newSession := func(t *testing.T) (*onnxruntime.DynamicAdvancedSession, func()) {
		t.Helper()
		inputName, outputName := embeddingGraph(t)

		cfg, err := NewSessionConfig(SessionSettings{IntraOpThreads: 1})
		require.NoError(t, err)

		session, err := cfg.NewSession(embeddingModelPath, []string{inputName}, []string{outputName}, nil)
		require.NoError(t, err)

		return session, func() {
			DestroySession(session)
			cfg.Destroy()
		}
	}

	t.Run("Success", func(t *testing.T) {
		requireRuntime(t, embeddingModelPath)
		session, cleanup := newSession(t)
		t.Cleanup(cleanup)

		assert.NoError(t, warmUpSession(session, []int64{1, 3, 112, 112}, 1))
	})
	t.Run("WrongGeometry", func(t *testing.T) {
		requireRuntime(t, embeddingModelPath)
		session, cleanup := newSession(t)
		t.Cleanup(cleanup)

		assert.Error(t, warmUpSession(session, []int64{1, 3, 64, 64}, 1))
	})
	t.Run("NoSession", func(t *testing.T) {
		err := warmUpSession(nil, []int64{1, 3, 112, 112}, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no session")
	})
	t.Run("NoGeometry", func(t *testing.T) {
		err := warmUpSession(nil, nil, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no session")
	})
	t.Run("DynamicGeometry", func(t *testing.T) {
		// A graph that leaves an axis dynamic cannot be verified until the caller resolves it,
		// so an unset dimension is reported rather than guessed.
		requireRuntime(t, embeddingModelPath)
		session, cleanup := newSession(t)
		t.Cleanup(cleanup)

		err := warmUpSession(session, []int64{1, 3, 0, 112}, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not fixed")
	})
	t.Run("EmptyGeometry", func(t *testing.T) {
		requireRuntime(t, embeddingModelPath)
		session, cleanup := newSession(t)
		t.Cleanup(cleanup)

		err := warmUpSession(session, nil, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no input geometry")
	})
}

func TestDestroySession(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.NotPanics(t, func() { DestroySession(nil) })
	})
	t.Run("Success", func(t *testing.T) {
		requireRuntime(t, embeddingModelPath)
		hook := captureProviderLog(t)
		inputName, outputName := embeddingGraph(t)

		cfg, err := NewSessionConfig(SessionSettings{IntraOpThreads: 1})
		require.NoError(t, err)
		t.Cleanup(cfg.Destroy)

		session, err := cfg.NewSession(embeddingModelPath, []string{inputName}, []string{outputName}, nil)
		require.NoError(t, err)

		DestroySession(session)
		assert.Empty(t, hook.AllEntries())
	})
}

func TestDestroyValue(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.NotPanics(t, func() { DestroyValue(nil) })
	})
	t.Run("Success", func(t *testing.T) {
		requireSessionRuntime(t)
		hook := captureProviderLog(t)

		tensor, err := onnxruntime.NewEmptyTensor[float32](onnxruntime.NewShape(1, 3, 8, 8))
		require.NoError(t, err)

		DestroyValue(tensor)
		assert.Empty(t, hook.AllEntries())
	})
}
