package face

// Defaults for the limits the cluster width guard works within.
const (
	ClusterSplitRoundsDefault = 6
	ClusterSplitShrinkDefault = 0.95
)

// ClusterSplitOff keeps a group wider than its own accept distance whole: no width test, no split,
// and no discard. Distinct from a budget of zero, which discards such a group outright.
const ClusterSplitOff = -1

// ClusterSplitRounds is how often a group wider than its own accept distance may be re-clustered
// before it is given up on. A gentler shrink needs more rounds to reach the same separation.
//
// ClusterSplitOff keeps such a group whole, zero discards it at once, and a positive budget
// re-clusters it and then discards what still does not fit - so zero is the strictest of the three.
var ClusterSplitRounds = ClusterSplitRoundsDefault

// ClusterSplitShrink is how much a round shortens the link distance by.
//
// Flat rather than sized to the overrun, since width is a chaining property and a cut sized to the
// symptom dissolves the group into noise. Gentle because a cut is large in this many dimensions: a
// tenth off the distance takes the median sample from ten neighbors to two.
var ClusterSplitShrink = ClusterSplitShrinkDefault

// ClusterSplitDisabled reports whether the width guard is switched off, which leaves an anonymous
// cluster with no width limit at all and is why the setting is reported rather than assumed.
func ClusterSplitDisabled() bool {
	return ClusterSplitRounds == ClusterSplitOff
}
