package config

import "github.com/photoprism/photoprism/internal/face"

// FaceSize returns the face size threshold in pixels.
func (c *Config) FaceSize() int {
	if c.options.FaceSize < 20 || c.options.FaceSize > 10000 {
		return 50
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
