package face

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// restoreEngine resets the active detection engine and what it was configured from, so a test
// that reconfigures it does not decide what its neighbors see.
func restoreEngine(t *testing.T) {
	engineMu.RLock()
	prev := activeEngine
	prevSettings := engineSettings
	prevLoaded := engineLoaded
	engineMu.RUnlock()

	t.Cleanup(func() {
		engineMu.Lock()
		current := activeEngine
		activeEngine = prev
		engineSettings = prevSettings
		engineLoaded = prevLoaded
		engineMu.Unlock()

		if current != nil && current != prev {
			_ = current.Close()
		}
	})
}

func TestParseEngine(t *testing.T) {
	cases := map[string]EngineName{
		"":         EngineAuto,
		"AUTO":     EngineAuto,
		"pigo":     EngineONNX,
		"  PIGO  ": EngineONNX,
		"onnx":     EngineONNX,
		"OnNx":     EngineONNX,
		"unknown":  EngineAuto,
		"none":     EngineNone,
	}

	for input, expected := range cases {
		if got := ParseEngine(input); got != expected {
			t.Fatalf("ParseEngine(%q) = %q, expected %q", input, got, expected)
		}
	}
}

func TestActiveEngineName(t *testing.T) {
	assert.Equal(t, EngineNone, ActiveEngineName())
}

func TestConfigureEngineReuse(t *testing.T) {
	restoreEngine(t)

	if _, err := os.Stat(detectorModelPath); err != nil {
		t.Skipf("faces: detector model is not installed (%s)", err)
	}

	settings := EngineSettings{Name: EngineONNX, ONNX: ONNXOptions{ModelPath: detectorModelPath, Threads: 1}}

	t.Run("UnchangedSettings", func(t *testing.T) {
		require.NoError(t, ConfigureEngine(settings))
		first := ActiveEngine()
		require.NotNil(t, first)

		require.NoError(t, ConfigureEngine(settings))

		assert.Same(t, first, ActiveEngine())
	})
	t.Run("ChangedSettings", func(t *testing.T) {
		require.NoError(t, ConfigureEngine(settings))
		first := ActiveEngine()
		require.NotNil(t, first)

		changed := settings
		changed.ONNX.Threads = settings.ONNX.Threads + 1
		require.NoError(t, ConfigureEngine(changed))

		assert.NotSame(t, first, ActiveEngine())
	})
	t.Run("AfterUseEngine", func(t *testing.T) {
		require.NoError(t, ConfigureEngine(settings))

		if prev := UseEngine(nil); prev != nil {
			_ = prev.Close()
		}

		require.NoError(t, ConfigureEngine(settings))

		assert.NotNil(t, ActiveEngine())
	})
	t.Run("DisabledThenConfigured", func(t *testing.T) {
		require.NoError(t, ConfigureEngine(EngineSettings{Name: EngineNone}))
		require.Nil(t, ActiveEngine())

		require.NoError(t, ConfigureEngine(settings))

		assert.NotNil(t, ActiveEngine())
	})
}
