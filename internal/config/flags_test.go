package config

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
)

// TestFaceDocDefaults pins that the face options publish the number that actually applies on a
// default install. The generated end-user reference and "--help" both print a numeric option
// with no flag default as 0, and a word like "detector" tells a reader deciding what to set
// nothing at all.
func TestFaceDocDefaults(t *testing.T) {
	defaults := make(map[string]string, len(Flags))

	for _, flag := range Flags {
		defaults[flag.Name()] = flag.Default()
	}

	t.Run("ResolvedModels", func(t *testing.T) {
		assert.Equal(t, face.DefaultDetectorName(), defaults["face-detector"])
		assert.Equal(t, face.DefaultModelName(), defaults["face-model"])
	})
	t.Run("Scores", func(t *testing.T) {
		assert.Equal(t, "9", defaults["face-score"])
		assert.Equal(t, "20", defaults["face-cluster-score"])
	})
	t.Run("CalibratedDistances", func(t *testing.T) {
		m := face.DefaultModel()
		require.NotNil(t, m)

		assert.Equal(t, strconv.FormatFloat(m.ClusterDist, 'g', -1, 64), defaults["face-cluster-dist"])
		assert.Equal(t, strconv.FormatFloat(m.ClusterRadius, 'g', -1, 64), defaults["face-cluster-radius"])
		assert.Equal(t, strconv.FormatFloat(m.CollisionDist, 'g', -1, 64), defaults["face-collision-dist"])
		assert.Equal(t, strconv.FormatFloat(m.Epsilon, 'g', -1, 64), defaults["face-epsilon-dist"])
		assert.Equal(t, strconv.FormatFloat(m.MatchDist, 'g', -1, 64), defaults["face-match-dist"])
	})
	t.Run("NoneIsZero", func(t *testing.T) {
		// The thread counts are derived from the CPU, so there is no number to publish and
		// "auto" is the honest answer.
		assert.Equal(t, "auto", defaults["face-detector-threads"])
		assert.Equal(t, "auto", defaults["face-model-threads"])
	})
}

func TestFaceDocDefault(t *testing.T) {
	assert.Equal(t, "0.85", faceDocDefault(0.85))
	assert.Equal(t, "9", faceDocDefault(9))
	assert.Empty(t, faceDocDefault(0), "an absent value must not be published as a setting")
	assert.Empty(t, faceDocDefault(-1))
}

func TestFaceModelDocDefault(t *testing.T) {
	assert.Equal(t, "0.85", faceModelDocDefault(func(m *face.EmbeddingModel) float64 { return m.ClusterDist }))
	assert.Empty(t, faceModelDocDefault(func(m *face.EmbeddingModel) float64 { return 0 }))
}
