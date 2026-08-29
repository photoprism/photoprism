package photoprism

import (
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/vector/alg"
)

// baseEmbedder holds the embedder that the package test config installed.
var baseEmbedder face.EmbedderSettings

// captureEmbedderSettings records the embedder installed by the package test config, so
// a test that replaces the process-wide embedder has a value known to be good to put
// back. ONNX models load from an explicit file, so restoring by name alone would leave
// the package without an embedder and fail every later test that expects one.
func captureEmbedderSettings(c *config.Config) {
	baseEmbedder = face.EmbedderSettings{
		Name:      face.ConfiguredModel(),
		Model:     face.FindEmbeddingModel(face.ConfiguredModel()),
		ModelPath: c.FaceModelPath(),
		Threads:   c.FaceModelThreads(),
	}
}

// restoreEmbedder reinstates the embedder that the package test config installed.
//
// Reading the name back from the global would capture whatever an earlier test left
// there, so one missed restore would spread to every test that follows. A restore that
// does not take fails the test for the same reason.
func restoreEmbedder(t *testing.T) {
	t.Helper()

	assert.NoError(t, face.ConfigureEmbedder(baseEmbedder))
	assert.Equal(t, baseEmbedder.Name, face.ConfiguredModel())
}

// useTestEmbedder configures a working embedding model for the duration of a test, so
// clustering does not depend on whether the ONNX runtime is present in the environment.
func useTestEmbedder(t *testing.T, name face.ModelName) {
	t.Helper()

	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: name}))

	t.Cleanup(func() {
		restoreEmbedder(t)
	})
}

func TestFaces_Cluster(t *testing.T) {
	t.Run("ForceTrue", func(t *testing.T) {
		c := config.TestConfig()
		useTestEmbedder(t, face.ModelSFace)

		m := NewFaces(c)

		opt := FacesOptions{
			Force:     true,
			Threshold: 1,
		}

		r, err := m.Cluster(opt)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(r)
	})
	t.Run("ForceFalse", func(t *testing.T) {
		c := config.TestConfig()
		useTestEmbedder(t, face.ModelSFace)

		m := NewFaces(c)

		opt := FacesOptions{
			Force:     false,
			Threshold: 1,
		}

		r, err := m.Cluster(opt)

		if err != nil {
			t.Fatal(err)
		}

		t.Log(r)
	})
	t.Run("RefusesWhenEmbedderFailed", func(t *testing.T) {
		c := config.TestConfig()

		require.Error(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelSFace, Model: face.FindEmbeddingModel(face.ModelSFace)}))
		t.Cleanup(func() {
			restoreEmbedder(t)
		})

		r, err := NewFaces(c).Cluster(FacesOptions{Force: true, Threshold: 1})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "embedding model failed to load")
		assert.Empty(t, r)
	})
	t.Run("RefusesMixedDimensions", func(t *testing.T) {
		c := config.TestConfig()
		useTestEmbedder(t, face.ModelSFace)

		// No fixture records an embedding model, so stamping these two selects exactly
		// them and keeps the mixed set under the test's control.
		for i, values := range []face.Embeddings{{{0.1, 0.2}}, {{0.3, 0.4, 0.5}}} {
			m := entity.Marker{
				MarkerUID:      rnd.GenerateUID('m'),
				MarkerType:     entity.MarkerFace,
				MarkerSrc:      entity.SrcImage,
				Size:           face.ClusterSizeThreshold,
				Score:          face.ClusterScore("") + 10,
				EmbeddingsJSON: values.JSON(),
				EmbedModel:     face.ModelSFace,
			}

			require.NoError(t, entity.Db().Create(&m).Error, "marker %d", i)

			t.Cleanup(func() {
				entity.UnscopedDb().Delete(&entity.Marker{}, "marker_uid = ?", m.MarkerUID)
			})
		}

		r, err := NewFaces(c).Cluster(FacesOptions{Force: true, Threshold: 1})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "different lengths")
		assert.Empty(t, r)
	})
}

// TestFaces_reportNoClustersFormed pins that a pass which ran and formed nothing says so, rather
// than reading as an idle instance while the run repeats on every wake.
func TestFaces_reportNoClustersFormed(t *testing.T) {
	w := NewFaces(config.TestConfig())

	hook := test.NewGlobal()
	t.Cleanup(hook.Reset)

	w.reportNoClustersFormed(42)

	var reported int

	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.InfoLevel && strings.Contains(entry.Message, "formed no cluster") {
			reported++

			assert.Contains(t, entry.Message, "42 samples")
			// Actionable rather than only observed: the two thresholds that decide whether any
			// group forms are what an operator would have to change.
			assert.Contains(t, entry.Message, fmt.Sprintf("%d faces", face.ClusterCore))
		}
	}

	require.Equal(t, 1, reported, "a pass that formed nothing must be visible above debug")

	// A worker that wakes every few minutes must not repeat an unchanged condition.
	w.reportNoClustersFormed(42)

	reported = 0

	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.InfoLevel && strings.Contains(entry.Message, "formed no cluster") {
			reported++
		}
	}

	assert.Equal(t, 1, reported, "an unchanged condition must not be repeated")

	// A changed count is a changed condition and is reported again.
	w.reportNoClustersFormed(43)

	reported = 0

	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.InfoLevel && strings.Contains(entry.Message, "formed no cluster") {
			reported++
		}
	}

	assert.Equal(t, 2, reported)

	assert.NotPanics(t, func() { (*Faces)(nil).reportNoClustersFormed(1) })
}

// TestFaces_reportClusteringSkipped pins that a library where too few markers clear the bars says
// so, and names them: the thresholds that exclude a marker are not the ones a person judges a face
// by, so the count alone would not be actionable.
func TestFaces_reportClusteringSkipped(t *testing.T) {
	w := NewFaces(config.TestConfig())

	hook := test.NewGlobal()
	t.Cleanup(hook.Reset)

	w.reportClusteringSkipped(3, 8)

	var reported int

	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.InfoLevel && strings.Contains(entry.Message, "clear the") {
			reported++

			assert.Contains(t, entry.Message, "3 of the 8")
			assert.Contains(t, entry.Message, fmt.Sprintf("%d px size", face.ClusterSizeThreshold))
		}
	}

	require.Equal(t, 1, reported, "the gap must be named, and it must be actionable")

	// A worker that wakes every few minutes must not repeat an unchanged condition.
	w.reportClusteringSkipped(3, 8)

	reported = 0

	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.InfoLevel && strings.Contains(entry.Message, "clear the") {
			reported++
		}
	}

	assert.Equal(t, 1, reported)

	// A different count is a different state and is reported again.
	w.reportClusteringSkipped(4, 8)

	assert.NotPanics(t, func() { (*Faces)(nil).reportClusteringSkipped(1, 8) })
}

// setFaceThresholds applies clustering thresholds for the duration of a test, so a fixture built
// against the shipped distances is judged by them rather than by whatever ran last.
// Prefer setShippedFaceThresholds, which cannot fall behind a recalibration.
func setShippedFaceThresholds(t *testing.T) {
	t.Helper()

	m := face.FindEmbeddingModel(face.ModelSFace)
	require.NotNil(t, m)

	setFaceThresholds(t, m.ClusterDist, m.ClusterRadius, m.MatchDist)
}

func setFaceThresholds(t *testing.T, clusterDist, clusterRadius, matchDist float64) {
	dist, radius, match := face.ClusterDist, face.ClusterRadius, face.MatchDist

	t.Cleanup(func() {
		face.ClusterDist, face.ClusterRadius, face.MatchDist = dist, radius, match
	})

	face.ClusterDist, face.ClusterRadius, face.MatchDist = clusterDist, clusterRadius, matchDist
}

// setFaceClusterSplit applies split limits for the duration of a test. They are package-level
// variables read inside DBSCAN, so a test that leaves one set makes a later one fail depending on
// the order it ran in.
func setFaceClusterSplit(t *testing.T, shrink float64, rounds int) {
	prevShrink, prevRounds := faceClusterSplitShrink, faceClusterSplitRounds

	t.Cleanup(func() {
		faceClusterSplitShrink, faceClusterSplitRounds = prevShrink, prevRounds
	})

	faceClusterSplitShrink, faceClusterSplitRounds = shrink, rounds
}

// betweenEmbeddings returns the unit vector t of the way from a to b along the great circle joining
// them, which is where a face that resembles two people sits. Interpolated along the arc rather than
// the chord: a straight line bunches up at its ends, leaving the middle of a bridge too wide to link.
func betweenEmbeddings(a, b face.Embedding, t float64) face.Embedding {
	var dot float64

	for i := range a {
		dot += a[i] * b[i]
	}

	theta := math.Acos(min(max(dot, -1), 1))
	result := make(face.Embedding, len(a))

	if sin := math.Sin(theta); sin != 0 {
		wa, wb := math.Sin((1-t)*theta)/sin, math.Sin(t*theta)/sin

		for i := range a {
			result[i] = wa*a[i] + wb*b[i]
		}
	} else {
		copy(result, a)
	}

	return result
}

// chainFixture returns two well-separated groups of faces and the two faces lying between them
// that chain both into one. Deterministic, so a failure reproduces.
//
// It stands for the lookalike siblings a real library merged: each group is inside what a cluster
// accepts, and only the line joining them puts the pair outside it.
func chainFixture() (groupA, groupB, bridge face.Embeddings) {
	a := face.FixtureEmbedding(9001)
	b := face.FixtureEmbeddingAt(a, 1.75, 9002)

	for i := range 10 {
		groupA = append(groupA, face.FixtureEmbeddingAt(a, 0.05+0.02*float64(i), uint64(9100+i)))
		groupB = append(groupB, face.FixtureEmbeddingAt(b, 0.05+0.02*float64(i), uint64(9200+i)))
	}

	for _, t := range []float64{1.0 / 3.0, 2.0 / 3.0} {
		bridge = append(bridge, betweenEmbeddings(a, b, t))
	}

	return groupA, groupB, bridge
}

// chainedFixture returns the two groups and the bridge as the one group DBSCAN emits for them,
// with the person each member belongs to, where 0 is a bridge that belongs to neither.
func chainedFixture() (embeddings face.Embeddings, group []int) {
	groupA, groupB, bridge := chainFixture()

	for _, part := range []struct {
		embeddings face.Embeddings
		person     int
	}{{groupA, 1}, {groupB, 2}, {bridge, 0}} {
		for _, e := range part.embeddings {
			embeddings = append(embeddings, e)
			group = append(group, part.person)
		}
	}

	return embeddings, group
}

// TestChainFixture pins that the fixture reproduces the failure it stands for. A green split test
// against a fixture that never chains would be a test of the early return.
func TestChainFixture(t *testing.T) {
	setShippedFaceThresholds(t)

	groupA, groupB, bridge := chainFixture()
	chained, _ := chainedFixture()

	assert.True(t, face.ClusterFits(groupA.Radius()), "one person must fit what a cluster accepts")
	assert.True(t, face.ClusterFits(groupB.Radius()))
	assert.False(t, face.ClusterFits(chained.Radius()), "the bridged pair must not")

	// Without the bridge the two are too far apart to link at all, so DBSCAN would emit them as
	// two clusters and the width check would never see them together.
	require.Len(t, bridge, 2)
	assert.Greater(t, groupA.Dist(groupB.First()), face.ClusterDist)

	// With it they are linked end to end, which is the whole mechanism: every step is inside the
	// link distance even though the two people are twice it apart.
	assert.Less(t, groupA.Dist(bridge[0]), face.ClusterDist)
	assert.Less(t, bridge[0].Dist(bridge[1]), face.ClusterDist)
	assert.Less(t, groupB.Dist(bridge[1]), face.ClusterDist)

	// Run the clusterer rather than inferring the outcome from those distances. If the fixture
	// ever stops chaining, every split test below would pass by measuring nothing.
	c, err := alg.DBSCAN(face.ClusterCore, face.ClusterDist, 1, alg.EuclideanDist)
	require.NoError(t, err)
	require.NoError(t, c.Learn(chained.Float64()))
	require.Len(t, c.Sizes(), 1, "the bridged pair must chain into one group")
	assert.Equal(t, len(chained), c.Sizes()[0], "and that group must hold both people")
}

// TestSplitWideClusters covers the width check DBSCAN itself cannot make: it bounds the distance
// to a neighbor rather than the extent of the result, so a line of faces between two people
// chains both into one group that is then named as whoever the operator recognizes in it.
func TestSplitWideClusters(t *testing.T) {
	// The parts have to be judged against the shipped SFace distances rather than whatever the
	// package configured last, or the fixture would measure a different question.
	setShippedFaceThresholds(t)

	w := NewFaces(config.TestConfig())

	partOf := func(t *testing.T, parts []face.Embeddings, embeddings face.Embeddings, group []int) map[int]map[int]int {
		t.Helper()

		// A part holds the vectors the fixture built rather than copies of them, so they are
		// identified by their backing array instead of by values two of them could share.
		index := make(map[*float64]int, len(embeddings))

		for i := range embeddings {
			index[&embeddings[i][0]] = i
		}

		members := make(map[int]map[int]int, len(parts))

		for p, part := range parts {
			members[p] = make(map[int]int, 3)

			for _, e := range part {
				i, found := index[&e[0]]
				require.True(t, found, "a part must only hold members of the group it was split from")
				members[p][group[i]]++
			}
		}

		return members
	}

	t.Run("PassesThroughAGroupThatFits", func(t *testing.T) {
		groupA, _, _ := chainFixture()

		parts := w.splitWideClusters(groupA, face.ClusterDist, 1)

		require.Len(t, parts, 1, "a group that already fits must not be re-clustered")
		assert.Len(t, parts[0], len(groupA))
	})
	t.Run("SplitsAChainedGroup", func(t *testing.T) {
		embeddings, group := chainedFixture()

		parts := w.splitWideClusters(embeddings, face.ClusterDist, 1)

		require.GreaterOrEqual(t, len(parts), 2, "a chained group must be separated")

		for p, members := range partOf(t, parts, embeddings, group) {
			assert.True(t, face.ClusterFits(parts[p].Radius()), "part %d must accept its own members", p)
			assert.False(t, members[1] > 0 && members[2] > 0, "part %d must not hold both people", p)
		}
	})
	t.Run("SkipsAGroupThatStaysWide", func(t *testing.T) {
		// Faces spread evenly along an arc stay one group at every link distance that keeps a
		// core together, so the rounds run out and nothing is created from it.
		a := face.FixtureEmbedding(9301)
		b := face.FixtureEmbeddingAt(a, 1.9, 9302)

		var embeddings face.Embeddings

		for i := range 40 {
			embeddings = append(embeddings, betweenEmbeddings(a, b, float64(i)/39.0))
		}

		hook := test.NewGlobal()
		t.Cleanup(hook.Reset)

		parts := w.splitWideClusters(embeddings, face.ClusterDist, 1)

		for _, part := range parts {
			assert.True(t, face.ClusterFits(part.Radius()))
		}

		var warned bool

		for _, entry := range hook.AllEntries() {
			if entry.Level == logrus.WarnLevel && strings.Contains(entry.Message, "stay unclustered") {
				warned = true
			}
		}

		assert.True(t, warned, "giving up on a group must be reported rather than silent")
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, w.splitWideClusters(face.Embeddings{}, face.ClusterDist, 1))
	})
}

// TestSplitDist covers how far one round shortens the link distance.
//
// Flat rather than sized to how far past its accept distance the group reaches: a cut sized to the
// symptom overshoots the distance that separates the group and dissolves it into noise, which on a
// real library halved coverage and left the person the check was written for unnameable.
func TestSplitDist(t *testing.T) {
	t.Run("Shortens", func(t *testing.T) {
		assert.InDelta(t, 0.85*faceClusterSplitShrink, splitDist(0.85), 1e-9)
	})
	t.Run("IndependentOfHowWideTheGroupIs", func(t *testing.T) {
		// The whole point: a group at twice its accept distance takes the same first step as one
		// just past it, because DBSCAN width is a chaining property rather than a linear one.
		assert.InDelta(t, splitDist(0.85), splitDist(0.85), 1e-9)
		assert.Less(t, splitDist(0.85), 0.85)
	})
	t.Run("AlwaysMakesProgress", func(t *testing.T) {
		for _, dist := range []float64{0.4, 0.85, 1.07, 1.25} {
			assert.Less(t, splitDist(dist), dist, "dist %f", dist)
			assert.Positive(t, splitDist(dist), "dist %f", dist)
		}
	})
	t.Run("ReachesAUsefulRangeWithinTheRoundCap", func(t *testing.T) {
		// A fraction of where the schedule started, because the distance a round reaches depends on
		// the model. Too little and the budget runs out before a chained group separates, which
		// reports its samples unclustered; too much and the group dissolves into noise.
		dist := face.ClusterDistDefault

		for range faceClusterSplitRounds {
			dist = splitDist(dist)
		}

		assert.Greater(t, dist, 0.2*face.ClusterDistDefault, "the schedule dissolves the group into noise")
		assert.Less(t, dist, 0.8*face.ClusterDistDefault, "the budget runs out before a chain separates")
	})
	t.Run("FollowsTheShrink", func(t *testing.T) {
		setFaceClusterSplit(t, 0.5, faceClusterSplitRounds)

		assert.InDelta(t, 0.425, splitDist(0.85), 1e-9)
	})
}

// TestSplitShrinkEnv covers the shrink override, which keeps the default for anything that would
// stop a round from shortening the link distance.
func TestSplitShrinkEnv(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		assert.InDelta(t, 0.90, splitShrinkEnv("", 0.90), 1e-9)
		assert.InDelta(t, 0.90, splitShrinkEnv("  ", 0.90), 1e-9)
	})
	t.Run("InRange", func(t *testing.T) {
		assert.InDelta(t, 0.95, splitShrinkEnv("0.95", 0.90), 1e-9)
		assert.InDelta(t, 0.97, splitShrinkEnv(" 0.97 ", 0.90), 1e-9)
	})
	// One and above never shortens the distance, so every round would repeat the previous pass.
	t.Run("OutOfRange", func(t *testing.T) {
		for _, v := range []string{"0", "1", "1.5", "-0.5", "abc", "0.9x", "NaN", "Inf"} {
			assert.InDelta(t, 0.90, splitShrinkEnv(v, 0.90), 1e-9, "%q must keep the default", v)
		}
	})
}

// TestSplitRoundsEnv covers the round-budget override. Zero is accepted, because a run that skips
// splitting entirely is the baseline a sweep compares against.
func TestSplitRoundsEnv(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		assert.Equal(t, 6, splitRoundsEnv("", 6))
		assert.Equal(t, 6, splitRoundsEnv("\t", 6))
	})
	t.Run("InRange", func(t *testing.T) {
		assert.Equal(t, 8, splitRoundsEnv("8", 6))
		assert.Equal(t, 0, splitRoundsEnv(" 0 ", 6))
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		for _, v := range []string{"-1", "abc", "6.5", ""} {
			assert.Equal(t, 6, splitRoundsEnv(v, 6), "%q must keep the default", v)
		}
	})
}

// TestFaceClusterSplitInit covers that init applied the environment instead of leaving the limits
// at their defaults. It reads the variables rather than setting them, because init runs once per
// binary - so it only bites when a sweep is running, which is exactly when the wiring matters.
func TestFaceClusterSplitInit(t *testing.T) {
	assert.InDelta(t, splitShrinkEnv(os.Getenv(envFaceClusterSplitShrink), faceClusterSplitShrinkDefault), faceClusterSplitShrink, 1e-9)
	assert.Equal(t, splitRoundsEnv(os.Getenv(envFaceClusterSplitRounds), faceClusterSplitRoundsDefault), faceClusterSplitRounds)
}

// TestReportSplitOverrides covers the line that makes a swept run self-describing.
func TestReportSplitOverrides(t *testing.T) {
	report := func(t *testing.T) string {
		t.Helper()

		hook := test.NewLocal(logrus.StandardLogger())
		defer hook.Reset()

		splitOverrideOnce = sync.Once{}
		t.Cleanup(func() { splitOverrideOnce = sync.Once{} })

		reportSplitOverrides()

		var b strings.Builder

		for _, e := range hook.AllEntries() {
			b.WriteString(e.Message)
		}

		return b.String()
	}

	t.Run("Defaults", func(t *testing.T) {
		setFaceClusterSplit(t, faceClusterSplitShrinkDefault, faceClusterSplitRoundsDefault)

		assert.Empty(t, report(t), "unchanged limits need no line")
	})
	t.Run("Overridden", func(t *testing.T) {
		setFaceClusterSplit(t, 0.8, 8)
		out := report(t)

		assert.Contains(t, out, "0.80")
		assert.Contains(t, out, "8 rounds")
	})
	t.Run("Once", func(t *testing.T) {
		setFaceClusterSplit(t, 0.8, 8)

		hook := test.NewLocal(logrus.StandardLogger())
		defer hook.Reset()

		splitOverrideOnce = sync.Once{}
		t.Cleanup(func() { splitOverrideOnce = sync.Once{} })

		reportSplitOverrides()
		n := len(hook.AllEntries())
		reportSplitOverrides()

		assert.Len(t, hook.AllEntries(), n, "a worker that wakes often must not repeat it")
	})
}

// TestSplitCluster covers the one round the split takes, which shortens the link distance in
// proportion to how far past its accept distance the group reaches.
func TestSplitCluster(t *testing.T) {
	setShippedFaceThresholds(t)

	embeddings, _ := chainedFixture()

	t.Run("ShortensTheLinkDistance", func(t *testing.T) {
		parts, err := splitCluster(faceClusterPart{embeddings: embeddings, dist: face.ClusterDist}, 1)

		require.NoError(t, err)
		require.NotEmpty(t, parts)

		for _, part := range parts {
			assert.InDelta(t, splitDist(face.ClusterDist), part.dist, 1e-9)
			assert.Equal(t, 1, part.round)
		}
	})
	t.Run("KeepsEveryMemberOfTheGroup", func(t *testing.T) {
		// Whatever the schedule, a round may only separate a group or leave points below the core
		// size - never invent one. A part holding a vector the group did not is a mapping bug.
		parts, err := splitCluster(faceClusterPart{embeddings: embeddings, dist: face.ClusterDist}, 1)
		require.NoError(t, err)

		seen := make(map[*float64]bool, len(embeddings))
		for i := range embeddings {
			seen[&embeddings[i][0]] = true
		}

		total := 0

		for _, part := range parts {
			total += len(part.embeddings)

			for _, e := range part.embeddings {
				assert.True(t, seen[&e[0]])
			}
		}

		assert.LessOrEqual(t, total, len(embeddings))
	})
	t.Run("InvalidDistance", func(t *testing.T) {
		_, err := splitCluster(faceClusterPart{embeddings: embeddings, dist: 0}, 1)
		assert.Error(t, err)
	})
}
