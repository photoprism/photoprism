package config

import (
	"math"
	"os"
	"path/filepath"
	"runtime"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
)

// FaceEngine returns the configured face detection engine. When the config is
// nil or the vision subsystem is not initialised it reports `face.EngineNone`
// so callers can short-circuit gracefully.
func (c *Config) FaceEngine() string {
	if c == nil {
		return face.EngineNone
	} else if c.options.FaceEngine == face.EnginePigo || c.options.FaceEngine == face.EngineONNX {
		return c.options.FaceEngine
	}

	if vision.Config == nil {
		return face.EngineNone
	}

	desired := face.ParseEngine(c.options.FaceEngine)
	modelPath := c.FaceEngineModelPath()

	if desired == face.EngineAuto {
		if modelPath != "" {
			if _, err := os.Stat(modelPath); err == nil {
				desired = face.EngineONNX
			} else {
				desired = face.EnginePigo
			}
		} else {
			desired = face.EnginePigo
		}

		c.options.FaceEngine = desired
	}

	return desired
}

// FaceEngineRunType returns the effective run type for the face detection engine.
// Detection and embedding always run together, so we defer to the face model
// configuration in the vision subsystem. If no detection model is configured,
// or faces are disabled entirely, the run type falls back to RunNever.
func (c *Config) FaceEngineRunType() vision.RunType {
	if c == nil {
		return vision.RunNever
	}

	if vision.Config == nil {
		return vision.RunNever
	}

	if c.DisableFaces() || c.FaceEngine() == face.EngineNone {
		return vision.RunNever
	}

	return vision.Config.RunType(vision.ModelTypeFace)
}

// FaceEngineShouldRun reports whether the face detection engine should execute in the
// specified scheduling context. The decision mirrors the face model run schedule in
// the vision subsystem, so detection stays aligned with embedding generation.
func (c *Config) FaceEngineShouldRun(when vision.RunType) bool {
	if c == nil {
		return false
	}

	if c.DisableFaces() || c.FaceEngine() == face.EngineNone {
		return false
	}

	run := c.FaceEngineRunType()
	when = vision.ParseRunType(when)

	switch run {
	case vision.RunNever:
		return false
	case vision.RunManual:
		return when == vision.RunManual
	case vision.RunAlways:
		return when != vision.RunNever
	case vision.RunNewlyIndexed:
		return when == vision.RunManual || when == vision.RunNewlyIndexed || when == vision.RunOnDemand
	case vision.RunOnDemand:
		return when == vision.RunAuto || when == vision.RunManual || when == vision.RunNewlyIndexed || when == vision.RunOnDemand
	case vision.RunOnSchedule:
		return when == vision.RunAuto || when == vision.RunManual || when == vision.RunOnSchedule || when == vision.RunOnDemand
	case vision.RunOnIndex:
		return when == vision.RunManual || when == vision.RunOnIndex
	case vision.RunAuto:
		fallthrough
	default:
		switch when {
		case vision.RunAuto, vision.RunAlways, vision.RunManual, vision.RunOnDemand:
			return true
		case vision.RunOnIndex:
			return c.FaceEngineThreads() > 2
		case vision.RunNewlyIndexed:
			return c.FaceEngineThreads() <= 2
		case vision.RunOnSchedule, vision.RunNever:
			return false
		}
	}

	return false
}

// FaceEngineRetry controls whether detection retries at a higher resolution should be performed.
func (c *Config) FaceEngineRetry() bool {
	if c == nil {
		return false
	}

	return c.FaceEngine() == face.EnginePigo && c.IndexWorkers() > 2
}

// FaceEngineThreads returns the configured thread count for ONNX inference.
func (c *Config) FaceEngineThreads() int {
	if c == nil {
		return 1
	} else if c.options.FaceEngineThreads <= 0 {
		threads := runtime.NumCPU() / 2
		if threads < 1 {
			threads = 1
		}

		c.options.FaceEngineThreads = threads

		return threads
	}

	return c.options.FaceEngineThreads
}

// FaceEngineModelPath returns the absolute path to the bundled SCRFD ONNX detector.
func (c *Config) FaceEngineModelPath() string {
	if c == nil {
		return ""
	}

	dir := filepath.Join(c.ModelsPath(), "scrfd")
	primary := filepath.Join(dir, face.DefaultONNXModelFilename)

	if _, err := os.Stat(primary); err == nil {
		return primary
	}

	alt := filepath.Join(dir, "scrfd_500m_bnkps_shape640x640.onnx")

	if _, err := os.Stat(alt); err == nil {
		return alt
	}

	return primary
}

// FaceSize returns the face size threshold in pixels.
func (c *Config) FaceSize() int {
	if c.options.FaceSize < 20 || c.options.FaceSize > 10000 {
		return face.SizeThreshold
	}

	return c.options.FaceSize
}

// FaceScore returns the face quality score threshold.
func (c *Config) FaceScore() float64 {
	if c.options.FaceScore < 1 || c.options.FaceScore > 100 {
		return face.ScoreThreshold
	}

	return c.options.FaceScore
}

// FaceOverlap returns the face area overlap threshold in percent.
func (c *Config) FaceOverlap() int {
	if c.options.FaceOverlap < 1 || c.options.FaceOverlap > 100 {
		return face.OverlapThreshold
	}

	return c.options.FaceOverlap
}

// FaceClusterSize returns the size threshold for faces forming a cluster in pixels.
func (c *Config) FaceClusterSize() int {
	if c.options.FaceClusterSize < 20 || c.options.FaceClusterSize > 10000 {
		return face.ClusterSizeThreshold
	}

	return c.options.FaceClusterSize
}

// FaceClusterScore returns the quality threshold for faces forming a cluster.
func (c *Config) FaceClusterScore() int {
	if c.options.FaceClusterScore < 1 || c.options.FaceClusterScore > 100 {
		return face.ClusterScoreThreshold
	}

	return c.options.FaceClusterScore
}

// FaceClusterCore returns the number of faces forming a cluster core.
func (c *Config) FaceClusterCore() int {
	if c.options.FaceClusterCore < 1 || c.options.FaceClusterCore > 100 {
		return face.ClusterCore
	}

	return c.options.FaceClusterCore
}

// FaceClusterDist returns the radius of faces forming a cluster core.
func (c *Config) FaceClusterDist() float64 {
	if c.options.FaceClusterDist < c.FaceCollisionDist() || c.options.FaceClusterDist > 1.5 {
		return face.ClusterDist
	}

	return c.options.FaceClusterDist
}

// FaceClusterRadius returns the maximum radius used when matching face clusters.
func (c *Config) FaceClusterRadius() float64 {
	if c.options.FaceClusterRadius < c.FaceCollisionDist() || c.options.FaceClusterRadius > 1.5 {
		return face.ClusterRadius
	}

	return c.options.FaceClusterRadius
}

// FaceCollisionDist returns the minimum distance used to differentiate embeddings.
func (c *Config) FaceCollisionDist() float64 {
	if c.options.FaceCollisionDist <= 0 || c.options.FaceCollisionDist > 1 {
		return face.CollisionDist
	}

	return c.options.FaceCollisionDist
}

// FaceEpsilonDist returns the distance slack applied to collision checks.
func (c *Config) FaceEpsilonDist() float64 {
	if c.options.FaceEpsilonDist <= 0 || c.options.FaceEpsilonDist > 0.1 {
		return face.Epsilon
	}

	return c.options.FaceEpsilonDist
}

// FaceMatchDist returns the offset distance when matching faces with clusters.
func (c *Config) FaceMatchDist() float64 {
	if c.options.FaceMatchDist < c.FaceCollisionDist() || c.options.FaceMatchDist > 1.5 {
		return face.MatchDist
	}

	return c.options.FaceMatchDist
}

// FaceSkipChildren reports whether child embeddings should be skipped when matching.
func (c *Config) FaceSkipChildren() bool {
	if c == nil {
		return face.SkipChildren
	}

	return c.options.FaceSkipChildren
}

// FaceAllowBackground reports whether background embeddings should not be ignored.
func (c *Config) FaceAllowBackground() bool {
	if c == nil {
		return !face.IgnoreBackground
	}

	return c.options.FaceAllowBackground
}

// FaceAngles returns the set of detection angles in radians.
func (c *Config) FaceAngles() []float64 {
	if len(c.options.FaceAngles) == 0 {
		return append([]float64(nil), face.DefaultAngles...)
	}

	angles := make([]float64, 0, len(c.options.FaceAngles))
	seen := make(map[float64]struct{}, len(c.options.FaceAngles))

	for _, angle := range c.options.FaceAngles {
		if math.IsNaN(angle) || math.IsInf(angle, 0) {
			continue
		}

		if angle < -math.Pi || angle > math.Pi {
			continue
		}

		if _, ok := seen[angle]; ok {
			continue
		}

		seen[angle] = struct{}{}
		angles = append(angles, angle)
	}

	if len(angles) == 0 {
		return append([]float64(nil), face.DefaultAngles...)
	}

	return angles
}
