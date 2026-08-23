package face

import (
	"errors"
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

// stubEngine returns a fixed result under a chosen name so detection can be exercised
// without a model file.
type stubEngine struct {
	name  EngineName
	faces Faces
	err   error
}

// Name returns the engine name this stub reports.
func (e *stubEngine) Name() EngineName { return e.name }

// Detect returns the canned result, leaving provenance to the caller.
func (e *stubEngine) Detect(string, int) (Faces, error) { return e.faces, e.err }

// Close releases nothing.
func (e *stubEngine) Close() error { return nil }

func TestDetect(t *testing.T) {
	t.Run("RecordsTheDetector", func(t *testing.T) {
		restoreEngine(t)

		// Every face carries the detector that found it, because its landmarks decide the
		// aligned crop: a vector whose detector is unknown cannot be compared with confidence.
		stub := &stubEngine{name: "stub", faces: Faces{{Score: 42}, {Score: 21}}}
		if prev := UseEngine(stub); prev != nil {
			_ = prev.Close()
		}

		faces, err := Detect("testdata/face.jpg", 20)

		require.NoError(t, err)
		require.Len(t, faces, 2)
		assert.Equal(t, EngineName("stub"), faces[0].DetectModel)
		assert.Equal(t, EngineName("stub"), faces[1].DetectModel)
	})
	t.Run("PartialResultWithError", func(t *testing.T) {
		restoreEngine(t)

		// An engine that fails partway still hands back what it found, and those faces are
		// stamped: the caller decides what to do with them, and an unattributed vector is
		// what this column exists to prevent.
		stub := &stubEngine{name: "stub", faces: Faces{{Score: 7}}, err: errors.New("partial")}
		if prev := UseEngine(stub); prev != nil {
			_ = prev.Close()
		}

		faces, err := Detect("testdata/face.jpg", 20)

		require.Error(t, err)
		require.Len(t, faces, 1)
		assert.Equal(t, EngineName("stub"), faces[0].DetectModel)
	})
	t.Run("NoEngine", func(t *testing.T) {
		restoreEngine(t)

		if prev := UseEngine(nil); prev != nil {
			_ = prev.Close()
		}

		_, err := Detect("testdata/face.jpg", 20)

		require.Error(t, err)
	})
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
