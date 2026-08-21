package face

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setThresholds applies clustering thresholds for the duration of a test.
func setThresholds(t *testing.T, clusterRadius, matchDist float64) {
	radius, dist := ClusterRadius, MatchDist

	t.Cleanup(func() {
		ClusterRadius, MatchDist = radius, dist
	})

	ClusterRadius, MatchDist = clusterRadius, matchDist
}

func TestClampSampleRadius(t *testing.T) {
	t.Run("WithinRange", func(t *testing.T) {
		setThresholds(t, ClusterRadiusDefault, MatchDistDefault)
		assert.InDelta(t, 0.2, ClampSampleRadius(0.2), 1e-9)
	})
	t.Run("AboveClusterRadius", func(t *testing.T) {
		setThresholds(t, ClusterRadiusDefault, MatchDistDefault)
		assert.InDelta(t, ClusterRadiusDefault, ClampSampleRadius(2), 1e-9)
	})
	t.Run("Negative", func(t *testing.T) {
		setThresholds(t, ClusterRadiusDefault, MatchDistDefault)
		assert.Zero(t, ClampSampleRadius(-1))
	})
	t.Run("FollowsConfiguredRadius", func(t *testing.T) {
		setThresholds(t, 0.9, MatchDistDefault)
		assert.InDelta(t, 0.9, ClampSampleRadius(2), 1e-9)
	})
}

func TestAcceptDist(t *testing.T) {
	t.Run("BelowCeiling", func(t *testing.T) {
		setThresholds(t, ClusterRadiusDefault, MatchDistDefault)
		assert.InDelta(t, 0.6, AcceptDist(0.2), 1e-9)
	})
	t.Run("ClampsStoredRadius", func(t *testing.T) {
		// A row written under a wider calibration must not widen the current gate.
		setThresholds(t, ClusterRadiusDefault, MatchDistDefault)
		assert.InDelta(t, ClusterRadiusDefault+MatchDistDefault, AcceptDist(2), 1e-9)
	})
	t.Run("NegativeRadius", func(t *testing.T) {
		setThresholds(t, ClusterRadiusDefault, MatchDistDefault)
		assert.InDelta(t, MatchDistDefault, AcceptDist(-1), 1e-9)
	})
	t.Run("CapsAtCeiling", func(t *testing.T) {
		// Both thresholds at their configurable maximum would otherwise accept every pair.
		setThresholds(t, AcceptDistMax, AcceptDistMax)
		assert.InDelta(t, AcceptDistMax, AcceptDist(AcceptDistMax), 1e-9)
	})
	t.Run("CeilingBelowRandomPairDistance", func(t *testing.T) {
		// Independent unit vectors average sqrt(2) apart, so the ceiling stays under it.
		assert.Less(t, float64(AcceptDistMax), math.Sqrt2)
	})
}
