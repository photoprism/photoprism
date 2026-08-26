package face

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/pkg/clean"
)

// EngineName identifies a face detection engine implementation.
type EngineName = string

const (
	// EngineAuto selects the default engine based on availability.
	EngineAuto EngineName = "auto"
	// EngineONNX enables the ONNX runtime-powered SCRFD detector.
	EngineONNX EngineName = "onnx"
	// EngineNone disables face detection.
	EngineNone EngineName = "none"
)

// ParseEngine normalizes user input and returns a supported engine name or EngineAuto when unknown.
// Legacy "pigo" values map to EngineONNX so older configs continue to work after detector removal.
func ParseEngine(s string) EngineName {
	s = strings.ToLower(strings.TrimSpace(s))

	switch s {
	case "pigo":
		return EngineONNX
	case EngineONNX, EngineNone:
		return s
	default:
		return EngineAuto
	}
}

// DetectionEngine represents a strategy for locating faces in an image.
type DetectionEngine interface {
	Name() EngineName
	// Detector names the detection model, which is what provenance records: the engine name
	// says which runtime produced a face, not which detector, and two detectors under one
	// runtime place different landmarks.
	Detector() DetectorName
	Detect(fileName string, minSize int) (Faces, error)
	Close() error
}

// EngineSettings capture configuration required to initialize a detection engine.
type EngineSettings struct {
	Name EngineName
	ONNX ONNXOptions
}

var (
	engineMu       sync.RWMutex
	activeEngine   DetectionEngine
	engineSettings EngineSettings
	engineLoaded   bool
)

// UseEngine replaces the active detection engine and returns the previous instance.
func UseEngine(engine DetectionEngine) (previous DetectionEngine) {
	return setEngine(engine, EngineSettings{}, false)
}

// setEngine replaces the active detection engine and records the settings it was built from,
// so a later call with the same settings can keep it. An engine installed through UseEngine
// records none, because the settings would not describe it.
func setEngine(engine DetectionEngine, settings EngineSettings, loaded bool) (previous DetectionEngine) {
	engineMu.Lock()
	previous = activeEngine
	activeEngine = engine
	engineSettings = settings
	engineLoaded = loaded
	engineMu.Unlock()

	return previous
}

// reuseEngine reports whether the active detection engine was loaded from these settings and
// can serve them again. An engine that failed to load is never reusable.
func reuseEngine(settings EngineSettings) bool {
	engineMu.RLock()
	defer engineMu.RUnlock()

	return engineLoaded && activeEngine != nil && engineSettings == settings
}

// ConfigureEngine selects and initializes the face detection engine based on the provided settings.
func ConfigureEngine(settings EngineSettings) error {
	// The detector holds an inference session like the embedding model does, so configuring
	// the same one again keeps it rather than reading and verifying the weights a second time.
	if reuseEngine(settings) {
		return nil
	}

	desired := ParseEngine(settings.Name)

	if desired == EngineAuto {
		desired = EngineONNX
	}

	var (
		newEngine DetectionEngine
		initErr   error
	)

	switch desired {
	case EngineNone:
		newEngine = nil
	case EngineONNX:
		if settings.ONNX.ModelPath == "" {
			initErr = fmt.Errorf("faces: ONNX model path is empty")
			break
		}

		newEngine, initErr = NewONNXEngine(settings.ONNX)
	default:
		initErr = fmt.Errorf("faces: unsupported detection engine %q", desired)
	}

	prev := setEngine(newEngine, settings, newEngine != nil && initErr == nil)

	if prev != nil {
		_ = prev.Close()
	}

	return initErr
}

// ActiveEngine returns the currently configured detection engine.
func ActiveEngine() DetectionEngine {
	engineMu.RLock()
	engine := activeEngine
	engineMu.RUnlock()
	return engine
}

// ActiveEngineSettings returns the settings the active engine was configured with, so a caller
// that changed one - a migration lowering the detection cutoff - can verify it took effect.
// The cutoff lives inside the inference session, where nothing else can observe it.
func ActiveEngineSettings() EngineSettings {
	engineMu.RLock()
	settings := engineSettings
	engineMu.RUnlock()

	return settings
}

// ActiveEngineName returns the name of the active engine.
// If there is no active engine, it returns "none."
func ActiveEngineName() EngineName {
	if engine := ActiveEngine(); engine != nil {
		return engine.Name()
	}

	return EngineNone
}

// ActiveDetector returns the name of the detector the active engine runs, or DetectorNone when
// there is none.
//
// An engine that cannot name its detector falls back to the engine name, which is the rule Detect
// records provenance under, so what this returns is comparable with a stored value.
func ActiveDetector() DetectorName {
	engine := ActiveEngine()

	if engine == nil {
		return DetectorNone
	}

	if detector := engine.Detector(); detector != "" {
		return detector
	}

	return engine.Name()
}

// DetectWithRetry runs the detector and, when it finds nothing, tries once more at a smaller
// minimum size. A retrySize of zero, or one not smaller than minSize, disables that.
//
// A crowd reduces every face below the ordinary minimum, so the frame is indexed as holding
// nobody. Retrying only on an empty result keeps bystanders unmarked everywhere else.
func DetectWithRetry(fileName string, minSize, retrySize int) (Faces, error) {
	faces, err := Detect(fileName, minSize)

	if err != nil || len(faces) > 0 || retrySize < 1 || retrySize >= minSize {
		return faces, err
	}

	retried, retryErr := Detect(fileName, retrySize)

	if retryErr != nil || len(retried) == 0 {
		return faces, err
	}

	log.Debugf("faces: found %d face(s) in %s below the %d px minimum", len(retried), clean.Log(filepath.Base(fileName)), minSize)

	return retried, nil
}

// Detect runs the active engine on the provided file and returns the detected faces.
//
// Each face records the detector that found it, this being the last frame where the producer of
// the landmarks is known. Provenance has three levels - none, the engine, the detector - and
// internal/ai/face/README.md records what each rules out.
func Detect(fileName string, minSize int) (Faces, error) {
	engine := ActiveEngine()
	if engine == nil {
		return Faces{}, fmt.Errorf("faces: detection engine not configured")
	}

	faces, err := engine.Detect(fileName, minSize)

	detector := engine.Detector()

	if detector == "" {
		detector = engine.Name()
	}

	// A detection below the configured quality is discarded here rather than filtered downstream,
	// so it never becomes a marker: what a detector reports as a face and is not costs an operator
	// a thumbnail to reject, whatever happens to it afterwards.
	kept := make(Faces, 0, len(faces))

	for i := range faces {
		if float64(faces[i].Score) < ScoreThreshold {
			continue
		}

		faces[i].DetectModel = detector
		kept = append(kept, faces[i])
	}

	return kept, err
}
