package face

import (
	"fmt"
	"strings"
	"sync"
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

// ActiveEngineName returns the name of the active engine.
// If there is no active engine, it returns "none."
func ActiveEngineName() EngineName {
	if engine := ActiveEngine(); engine != nil {
		return engine.Name()
	}

	return EngineNone
}

// Detect runs the active engine on the provided file and returns the detected faces.
//
// Each face records the detector that found it, because this is the last frame where the
// producer of the landmarks is known: the crop they align is what makes an embedding
// comparable, so everything downstream would have to ask global configuration instead.
func Detect(fileName string, minSize int) (Faces, error) {
	engine := ActiveEngine()
	if engine == nil {
		return Faces{}, fmt.Errorf("faces: detection engine not configured")
	}

	faces, err := engine.Detect(fileName, minSize)

	for i := range faces {
		faces[i].DetectModel = engine.Name()
	}

	return faces, err
}
