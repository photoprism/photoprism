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
		// Values this wide cannot be configured, but a stored radius written before a
		// recalibration can still be read back, and it must not accept every pair.
		setThresholds(t, AcceptDistMax, AcceptDistMax)
		assert.InDelta(t, AcceptDistMax, AcceptDist(AcceptDistMax), 1e-9)
	})
	t.Run("CeilingBelowRandomPairDistance", func(t *testing.T) {
		// Independent unit vectors average sqrt(2) apart, so the ceiling stays under it.
		assert.Less(t, float64(AcceptDistMax), math.Sqrt2)
	})
	t.Run("ConfigurableRangeBelowCeiling", func(t *testing.T) {
		// What an operator may set has to stay under the runtime backstop, or a configured
		// value would be accepted, reported, and then clipped where it is read.
		assert.Less(t, float64(ConfigDistMax), float64(AcceptDistMax))
	})
}

// TestAmbiguityDist pins the cutoff to Epsilon rather than to a literal. The two came apart once
// already: the cutoff stayed at 0.02 while Epsilon was scaled per model.
func TestAmbiguityDist(t *testing.T) {
	restore := Epsilon
	t.Cleanup(func() { Epsilon = restore })

	t.Run("Default", func(t *testing.T) {
		Epsilon = EpsilonDefault
		assert.InDelta(t, 0.02, AmbiguityDist(), 0.0001)
	})
	t.Run("FollowsEpsilon", func(t *testing.T) {
		// FACE_EPSILON_DIST reaches this through Config.Propagate, so an operator narrowing the
		// gap narrows the cutoff with it rather than leaving a band that resolves neither way.
		Epsilon = 0.004
		assert.InDelta(t, 0.008, AmbiguityDist(), 0.0001)
		assert.Greater(t, AmbiguityDist(), Epsilon, "a cutoff at or below Epsilon would record a non-positive radius")
	})
}

// TestDetectorScore checks that a report can always state a cutoff some detector enforces. It
// returned zero for an unknown name before, which reads as "nothing is filtered".
func TestDetectorScore(t *testing.T) {
	t.Run("Registered", func(t *testing.T) {
		assert.InDelta(t, float64(FindDetector(DetectorYuNet).MinScore), DetectorScore(DetectorYuNet), 0.5)
		assert.InDelta(t, float64(FindDetector(DetectorSCRFD).MinScore), DetectorScore(DetectorSCRFD), 0.5)
	})
	t.Run("Unregistered", func(t *testing.T) {
		assert.Equal(t, DetectorScore(DefaultDetector().Name), DetectorScore("nonexistent"))
	})
	t.Run("DetectionDisabled", func(t *testing.T) {
		assert.Positive(t, DetectorScore(DetectorNone))
	})
}

// TestNoScoreThreshold pins that "no cutoff" is expressible. Zero cannot say it, because zero is
// taken by "let the detector decide" and every detector registers one.
func TestNoScoreThreshold(t *testing.T) {
	assert.Negative(t, NoScoreThreshold)
	assert.NotEqual(t, ScoreThresholdDefault, NoScoreThreshold)
}
