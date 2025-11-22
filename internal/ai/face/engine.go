package face

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

type EngineName = string

const (
	EngineAuto EngineName = "auto"
	EnginePigo EngineName = "pigo"
	EngineONNX EngineName = "onnx"
	EngineNone EngineName = "none"
)

// ParseEngine normalizes user input and returns a supported engine name or EngineAuto when unknown.
func ParseEngine(s string) EngineName {
	s = strings.ToLower(strings.TrimSpace(s))

	switch s {
	case EnginePigo, EngineONNX, EngineNone:
		return s
	default:
		return EngineAuto
	}
}

// DetectionEngine represents a strategy for locating faces in an image.
type DetectionEngine interface {
	Name() EngineName
	Detect(fileName string, findLandmarks bool, minSize int) (Faces, error)
	Close() error
}

// EngineSettings capture configuration required to initialize a detection engine.
type EngineSettings struct {
	Name EngineName
	ONNX ONNXOptions
}

var (
	engineMu     sync.RWMutex
	activeEngine DetectionEngine
)

func init() {
	activeEngine = newPigoEngine()
}

// UseEngine replaces the active detection engine and returns the previous instance.
func UseEngine(engine DetectionEngine) (previous DetectionEngine) {
	engineMu.Lock()
	prev := activeEngine
	if engine == nil {
		activeEngine = newPigoEngine()
	} else {
		activeEngine = engine
	}
	engineMu.Unlock()
	return prev
}

// ConfigureEngine selects and initializes the face detection engine based on the provided settings.
func ConfigureEngine(settings EngineSettings) error {
	desired := ParseEngine(settings.Name)

	if desired == EngineAuto {
		desired = EnginePigo
	}

	var (
		newEngine DetectionEngine
		initErr   error
	)

	switch desired {
	case EngineNone:
		return errors.New("no engine selected")
	case EngineONNX:
		if settings.ONNX.ModelPath == "" {
			initErr = fmt.Errorf("ONNX model path is empty")
			newEngine = newPigoEngine()
			break
		}

		if newEngine, initErr = NewONNXEngine(settings.ONNX); initErr != nil {
			newEngine = newPigoEngine()
		}
	case EnginePigo:
		fallthrough
	default:
		if desired != EnginePigo {
			log.Warnf("faces: unknown detection engine %q, falling back to pigo", desired)
		}
		newEngine = newPigoEngine()
	}

	prev := UseEngine(newEngine)
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
func Detect(fileName string, findLandmarks bool, minSize int) (Faces, error) {
	engine := ActiveEngine()
	if engine == nil {
		return Faces{}, fmt.Errorf("faces: detection engine not configured")
	}
	return engine.Detect(fileName, findLandmarks, minSize)
}
