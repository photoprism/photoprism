package face

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClusterSplitDisabled covers the sentinel that switches the width guard off, which is the one
// negative round budget with a meaning: every other value below it resolves to the default.
func TestClusterSplitDisabled(t *testing.T) {
	restore := ClusterSplitRounds
	t.Cleanup(func() { ClusterSplitRounds = restore })

	t.Run("Off", func(t *testing.T) {
		ClusterSplitRounds = ClusterSplitOff
		assert.True(t, ClusterSplitDisabled())
	})
	t.Run("DiscardIsNotOff", func(t *testing.T) {
		// The distinction the sentinel exists for: zero is the strictest setting, not the loosest.
		ClusterSplitRounds = 0
		assert.False(t, ClusterSplitDisabled())
	})
	t.Run("Default", func(t *testing.T) {
		ClusterSplitRounds = ClusterSplitRoundsDefault
		assert.False(t, ClusterSplitDisabled())
	})
}

// TestClusterSplitDefaults pins the shipped limits, which decide how far a chained group is cut and
// therefore what a library forms without any override.
func TestClusterSplitDefaults(t *testing.T) {
	assert.Equal(t, ClusterSplitRoundsDefault, ClusterSplitRounds)
	assert.InDelta(t, ClusterSplitShrinkDefault, ClusterSplitShrink, 1e-9)
	assert.Positive(t, ClusterSplitRoundsDefault, "the default must split rather than discard")
	assert.Greater(t, ClusterSplitShrinkDefault, 0.0)
	assert.Less(t, ClusterSplitShrinkDefault, 1.0, "a factor of one would never shorten the distance")
}
