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
	name     EngineName
	detector DetectorName
	faces    Faces
	err      error
}

// Name returns the engine name this stub reports.
func (e *stubEngine) Name() EngineName { return e.name }

// Detector returns the stub's detector name.
func (e *stubEngine) Detector() DetectorName { return e.detector }

// Detect returns the canned result, leaving provenance to the caller.
func (e *stubEngine) Detect(string, int) (Faces, error) { return e.faces, e.err }

// Close releases nothing.
func (e *stubEngine) Close() error { return nil }

func TestDetect(t *testing.T) {
	t.Run("RecordsTheDetector", func(t *testing.T) {
		restoreEngine(t)

		// Every face carries the detector that found it, because its landmarks decide the
		// aligned crop: a vector whose detector is unknown cannot be compared with confidence.
		stub := &stubEngine{name: "stub", detector: "stub-detector", faces: Faces{{Score: 90}, {Score: 80}}}
		if prev := UseEngine(stub); prev != nil {
			_ = prev.Close()
		}

		faces, err := Detect("testdata/face.jpg", 20)

		require.NoError(t, err)
		require.Len(t, faces, 2)
		assert.Equal(t, DetectorName("stub-detector"), faces[0].DetectModel)
		assert.Equal(t, DetectorName("stub-detector"), faces[1].DetectModel)
	})
	t.Run("PartialResultWithError", func(t *testing.T) {
		restoreEngine(t)

		// An engine that fails partway still hands back what it found, and those faces are
		// stamped: the caller decides what to do with them, and an unattributed vector is
		// what this column exists to prevent.
		stub := &stubEngine{name: "stub", detector: "stub-detector", faces: Faces{{Score: 70}}, err: errors.New("partial")}
		if prev := UseEngine(stub); prev != nil {
			_ = prev.Close()
		}

		faces, err := Detect("testdata/face.jpg", 20)

		require.Error(t, err)
		require.Len(t, faces, 1)
		assert.Equal(t, DetectorName("stub-detector"), faces[0].DetectModel)
	})
	t.Run("EngineWithoutADetectorName", func(t *testing.T) {
		restoreEngine(t)

		// The engine name is coarser provenance but not none: it rules out the legacy Go
		// detector, so the five canonical landmarks are known to be there.
		stub := &stubEngine{name: "onnx", faces: Faces{{Score: 90}}}
		if prev := UseEngine(stub); prev != nil {
			_ = prev.Close()
		}

		faces, err := Detect("testdata/1.jpg", 20)

		require.NoError(t, err)
		require.Len(t, faces, 1)
		assert.Equal(t, DetectorName("onnx"), faces[0].DetectModel)
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

func TestDetectWithRetry(t *testing.T) {
	t.Run("RetriesOnlyWhenNothingWasFound", func(t *testing.T) {
		restoreEngine(t)

		var calls []detectCall

		stub := &recordingEngine{
			calls: &calls,
			// A face is returned for the smaller minimum only, which is the crowd case.
			result: func(minSize int) Faces {
				if minSize <= 10 {
					return Faces{{Score: 90}}
				}
				return nil
			},
		}

		if prev := UseEngine(stub); prev != nil {
			_ = prev.Close()
		}

		faces, err := DetectWithRetry("testdata/face.jpg", 25, 10)

		require.NoError(t, err)
		require.Len(t, faces, 1, "the second pass result is returned")
		require.Len(t, calls, 2, "the first pass runs before the second")
		assert.Equal(t, 25, calls[0].minSize)
		assert.Equal(t, 10, calls[1].minSize)
		assert.Equal(t, DetectorName("stub-detector"), faces[0].DetectModel)
	})
	t.Run("NoRetryWhenTheFirstPassFoundSomething", func(t *testing.T) {
		restoreEngine(t)

		var calls []detectCall

		stub := &recordingEngine{calls: &calls, result: func(int) Faces { return Faces{{Score: 90}} }}

		if prev := UseEngine(stub); prev != nil {
			_ = prev.Close()
		}

		faces, err := DetectWithRetry("testdata/face.jpg", 25, 10)

		require.NoError(t, err)
		require.Len(t, faces, 1)
		assert.Len(t, calls, 1, "a picture whose subject was found must not be searched again")
	})
	t.Run("Disabled", func(t *testing.T) {
		restoreEngine(t)

		var calls []detectCall

		stub := &recordingEngine{calls: &calls, result: func(int) Faces { return nil }}

		if prev := UseEngine(stub); prev != nil {
			_ = prev.Close()
		}

		for _, retrySize := range []int{0, -1, 25, 30} {
			calls = nil
			_, err := DetectWithRetry("testdata/face.jpg", 25, retrySize)
			require.NoError(t, err)
			assert.Len(t, calls, 1, "retry size %d must not trigger a second pass", retrySize)
		}
	})
	t.Run("FirstPassErrorIsNotRetried", func(t *testing.T) {
		restoreEngine(t)

		var calls []detectCall

		stub := &recordingEngine{calls: &calls, err: errors.New("broken"), result: func(int) Faces { return nil }}

		if prev := UseEngine(stub); prev != nil {
			_ = prev.Close()
		}

		_, err := DetectWithRetry("testdata/face.jpg", 25, 10)

		require.Error(t, err)
		assert.Len(t, calls, 1, "a detector that failed is not asked again")
	})
}

// detectCall records what one detection pass was asked for, so the two passes can be told
// apart by their minimum size.
type detectCall struct{ minSize int }

// recordingEngine records the minimum size each pass asked for.
type recordingEngine struct {
	calls  *[]detectCall
	result func(minSize int) Faces
	err    error
}

// Name returns the stub engine name.
func (e *recordingEngine) Name() EngineName { return "stub" }

// Detector returns the stub detector name.
func (e *recordingEngine) Detector() DetectorName { return "stub-detector" }

// Detect records the call and returns whatever the result function yields.
func (e *recordingEngine) Detect(_ string, minSize int) (Faces, error) {
	*e.calls = append(*e.calls, detectCall{minSize})
	return e.result(minSize), e.err
}

// Close releases nothing.
func (e *recordingEngine) Close() error { return nil }

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

func TestActiveDetector(t *testing.T) {
	t.Run("NoEngine", func(t *testing.T) {
		prev := UseEngine(nil)
		t.Cleanup(func() {
			if current := UseEngine(prev); current != nil {
				_ = current.Close()
			}
		})

		assert.Equal(t, DetectorNone, ActiveDetector())
	})
	t.Run("NamedDetector", func(t *testing.T) {
		prev := UseEngine(&stubEngine{name: EngineONNX, detector: DetectorYuNet})
		t.Cleanup(func() { UseEngine(prev) })

		assert.Equal(t, DetectorYuNet, ActiveDetector())
	})
	t.Run("FallsBackToTheEngineName", func(t *testing.T) {
		// An engine that cannot name its detector still rules out the legacy landmark
		// vocabulary, which is the coarser level provenance records.
		prev := UseEngine(&stubEngine{name: EngineONNX})
		t.Cleanup(func() { UseEngine(prev) })

		assert.Equal(t, EngineONNX, ActiveDetector())
	})
}

func TestDetectScoreThreshold(t *testing.T) {
	t.Run("DiscardsLowQualityDetections", func(t *testing.T) {
		// Filtered before a marker exists rather than after: what a detector reports as a
		// face and is not costs an operator a thumbnail to reject once it is in the index.
		restoreEngine(t)

		restore := ScoreThreshold
		t.Cleanup(func() { ScoreThreshold = restore })
		ScoreThreshold = 50

		stub := &stubEngine{name: "stub", detector: "stub-detector", faces: Faces{{Score: 49}, {Score: 50}, {Score: 90}}}
		if prev := UseEngine(stub); prev != nil {
			_ = prev.Close()
		}

		faces, err := Detect("testdata/face.jpg", 20)

		require.NoError(t, err)
		require.Len(t, faces, 2, "only the detections at or above the threshold are kept")
		assert.Equal(t, 50, faces[0].Score)
		assert.Equal(t, 90, faces[1].Score)
		assert.Equal(t, DetectorName("stub-detector"), faces[0].DetectModel)
	})
	t.Run("AllDiscarded", func(t *testing.T) {
		restoreEngine(t)

		restore := ScoreThreshold
		t.Cleanup(func() { ScoreThreshold = restore })
		ScoreThreshold = 99

		stub := &stubEngine{name: "stub", detector: "stub-detector", faces: Faces{{Score: 90}}}
		if prev := UseEngine(stub); prev != nil {
			_ = prev.Close()
		}

		faces, err := Detect("testdata/face.jpg", 20)

		require.NoError(t, err)
		assert.Empty(t, faces)
	})
}

func TestActiveEngineSettings(t *testing.T) {
	prev := UseEngine(nil)
	t.Cleanup(func() {
		if current := UseEngine(prev); current != nil {
			_ = current.Close()
		}
	})

	t.Run("NoEngine", func(t *testing.T) {
		assert.NotPanics(t, func() { _ = ActiveEngineSettings() })
	})
}
