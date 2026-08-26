package photoprism

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
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
				Size:           100,
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

// TestFaces_reportClusteringSkipped pins that a library where nothing ever clusters says so. It
// was a debug line, which made it indistinguishable from one where clustering ran and found
// nothing - and the thresholds that exclude a marker are not the ones a person judges a face by.
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
func setFaceThresholds(t *testing.T, clusterDist, clusterRadius, matchDist float64) {
	dist, radius, match := face.ClusterDist, face.ClusterRadius, face.MatchDist

	t.Cleanup(func() {
		face.ClusterDist, face.ClusterRadius, face.MatchDist = dist, radius, match
	})

	face.ClusterDist, face.ClusterRadius, face.MatchDist = clusterDist, clusterRadius, matchDist
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
	setFaceThresholds(t, 0.85, 0.60, 0.35)

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
}

// TestSplitWideClusters covers the width check DBSCAN itself cannot make: it bounds the distance
// to a neighbor rather than the extent of the result, so a line of faces between two people
// chains both into one group that is then named as whoever the operator recognizes in it.
func TestSplitWideClusters(t *testing.T) {
	// The parts have to be judged against the shipped SFace distances rather than whatever the
	// package configured last, or the fixture would measure a different question.
	setFaceThresholds(t, 0.85, 0.60, 0.35)

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

		parts := splitWideClusters(groupA, face.ClusterDist, 1)

		require.Len(t, parts, 1, "a group that already fits must not be re-clustered")
		assert.Len(t, parts[0], len(groupA))
	})
	t.Run("SplitsAChainedGroup", func(t *testing.T) {
		embeddings, group := chainedFixture()

		parts := splitWideClusters(embeddings, face.ClusterDist, 1)

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

		parts := splitWideClusters(embeddings, face.ClusterDist, 1)

		for _, part := range parts {
			assert.True(t, face.ClusterFits(part.Radius()))
		}

		var warned bool

		for _, entry := range hook.AllEntries() {
			if entry.Level == logrus.WarnLevel && strings.Contains(entry.Message, "stays wider than") {
				warned = true
			}
		}

		assert.True(t, warned, "giving up on a group must be reported rather than silent")
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, splitWideClusters(face.Embeddings{}, face.ClusterDist, 1))
	})
}

// TestSplitCluster covers the one round the split takes, which shortens the link distance in
// proportion to how far past its accept distance the group reaches.
func TestSplitCluster(t *testing.T) {
	setFaceThresholds(t, 0.85, 0.60, 0.35)

	embeddings, _ := chainedFixture()
	radius := embeddings.Radius()

	t.Run("ShortensTheLinkDistance", func(t *testing.T) {
		parts, err := splitCluster(faceClusterPart{embeddings: embeddings, dist: face.ClusterDist}, radius, 1)

		require.NoError(t, err)
		require.NotEmpty(t, parts)

		for _, part := range parts {
			assert.Less(t, part.dist, face.ClusterDist)
			assert.Equal(t, 1, part.round)
		}
	})
	t.Run("MakesProgressWhenBarelyTooWide", func(t *testing.T) {
		// A group just past its accept distance would otherwise be re-clustered at the distance
		// that produced it, and every round would return it unchanged.
		barely := face.AcceptDist(radius) + 0.001

		parts, err := splitCluster(faceClusterPart{embeddings: embeddings, dist: face.ClusterDist}, barely, 1)

		require.NoError(t, err)
		require.NotEmpty(t, parts)
		assert.InDelta(t, face.ClusterDist*faceClusterSplitShrink, parts[0].dist, 1e-9)
	})
	t.Run("InvalidDistance", func(t *testing.T) {
		_, err := splitCluster(faceClusterPart{embeddings: embeddings, dist: 0}, radius, 1)
		assert.Error(t, err)
	})
}
