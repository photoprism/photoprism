package config

import (
	"math"

	"github.com/photoprism/photoprism/internal/ai/face"
)

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
	if c.options.FaceClusterDist < 0.1 || c.options.FaceClusterDist > 1.5 {
		return face.ClusterDist
	}

	return c.options.FaceClusterDist
}

// FaceMatchDist returns the offset distance when matching faces with clusters.
func (c *Config) FaceMatchDist() float64 {
	if c.options.FaceMatchDist < 0.1 || c.options.FaceMatchDist > 1.5 {
		return face.MatchDist
	}

	return c.options.FaceMatchDist
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
