package photoprism

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// retryTestMarkers creates clusterable markers spread around a base embedding and returns their
// UIDs. The spread stays well inside the link distance, so how many of them cluster is decided by
// the core size alone - which is the variable these tests move.
func retryTestMarkers(t *testing.T, base face.Embedding, n int, seed uint64) []string {
	t.Helper()

	src := rand.New(rand.NewPCG(seed, seed+1)) //nolint:gosec // deterministic fixtures, not security
	uids := make([]string, 0, n)

	for range n {
		m := &entity.Marker{
			MarkerUID:      rnd.GenerateUID('m'),
			FileUID:        "fs6sg6bw45bnlqdw",
			MarkerType:     entity.MarkerFace,
			MarkerSrc:      entity.SrcImage,
			Size:           face.ClusterSizeThresholdDefault + 40,
			ThumbSize:      face.ClusterSizeThresholdDefault + 40,
			EmbedDetail:    face.EmbedDetailFull,
			Score:          100,
			EmbedModel:     face.EmbeddingModelName(),
			DetectModel:    face.DetectorYuNet,
			EmbeddingsJSON: face.Embeddings{benchmarkEmbeddingAt(base, 0.1, src)}.JSON(),
			W:              0.1,
			H:              0.1,
		}

		require.NoError(t, entity.Db().Create(m).Error)

		uids = append(uids, m.MarkerUID)
	}

	return uids
}

// retryTestClusters returns the cluster ids the given markers were attached to, without the empty
// one, so a test can name what a pass actually formed rather than counting rows.
func retryTestClusters(t *testing.T, uids []string) map[string][]string {
	t.Helper()

	result := make(map[string][]string)

	for _, uid := range uids {
		m := entity.FindMarker(uid)
		require.NotNil(t, m, uid)

		if m.FaceID == "" {
			continue
		}

		result[m.FaceID] = append(result[m.FaceID], uid)
	}

	return result
}

// TestFaces_startClusterRetry covers the second clustering pass, which runs at a lower core over
// the markers matching left unclustered.
//
// The library is the same in every case and only face-cluster-core-retry moves, so what the pass
// contributes is measured rather than inferred from a total.
func TestFaces_startClusterRetry(t *testing.T) {
	w := isolatedTestFaces(t, "facesclusterretry")

	restore := face.ClusterCore
	t.Cleanup(func() { face.ClusterCore = restore })
	face.ClusterCore = face.ClusterCoreDefault

	w.conf.Options().FaceClusterCore = face.ClusterCoreDefault

	// Four faces of one person: one short of the core the first pass needs, which is the whole
	// population this option exists for. The subtests share them on purpose and run in order: the
	// disabled case is the control that shows the first pass cannot cluster them.
	base := face.RandomEmbedding()
	residue := retryTestMarkers(t, base, 4, 11)

	t.Run("DisabledFormsNoCluster", func(t *testing.T) {
		w.conf.Options().FaceClusterCoreRetry = -1
		require.Equal(t, -1, w.conf.FaceClusterCoreRetry())

		// Threshold 1 because face.SampleThreshold is derived from the first core and would skip
		// both passes on a fixture this small - it decides whether a pass is worth running, which
		// is not what these cases are about.
		result, err := w.start(FacesOptions{Force: true, Threshold: 1})
		require.NoError(t, err)

		assert.Zero(t, result.Retried, "no second pass runs")
		assert.Empty(t, retryTestClusters(t, residue), "four faces do not reach a core of five")
	})
	t.Run("RefusesACoreBelowTwo", func(t *testing.T) {
		// A cluster seeded from one embedding has no centroid and is never offered to the matcher,
		// which is the same floor face-cluster-core has.
		retried, err := w.ClusterRetry(1)
		require.NoError(t, err)
		assert.Empty(t, retried)

		retried, err = w.ClusterRetry(0)
		require.NoError(t, err)
		assert.Empty(t, retried)
	})
	t.Run("FormsAClusterFromTheResidue", func(t *testing.T) {
		w.conf.Options().FaceClusterCoreRetry = 4
		require.Equal(t, 4, w.conf.FaceClusterCoreRetry())

		result, err := w.start(FacesOptions{Force: true, Threshold: 1})
		require.NoError(t, err)

		assert.Equal(t, 1, result.Retried, "the second pass forms the cluster the first could not")

		// Asserted on membership rather than on a count: a count pins the fixture, while this
		// pins that the pass clustered exactly what was left over and nothing else.
		formed := retryTestClusters(t, residue)
		require.Len(t, formed, 1)

		for _, members := range formed {
			assert.ElementsMatch(t, residue, members)
		}
	})
}

// TestFaces_startClusterRetryRunsUnforced pins that an ordinary worker run reaches the second pass.
//
// The trigger the first pass evaluates counts markers newer than the newest cluster, so the
// clusters the first pass has just created close it. Re-asking it for the retry would skip the
// second pass in exactly the runs where the first one did work - and every scheduled run is
// unforced, so that is the path this option exists on.
func TestFaces_startClusterRetryRunsUnforced(t *testing.T) {
	w := isolatedTestFaces(t, "facesclusterretryunforced")

	restore := face.ClusterCore
	t.Cleanup(func() { face.ClusterCore = restore })
	face.ClusterCore = face.ClusterCoreDefault

	w.conf.Options().FaceClusterCore = face.ClusterCoreDefault
	w.conf.Options().FaceClusterCoreRetry = 4
	require.Equal(t, 4, w.conf.FaceClusterCoreRetry())

	// Five faces of one person, which the first pass clusters, and four of another, which only the
	// second can. The first pass forming something is what closes the trigger.
	first := retryTestMarkers(t, face.RandomEmbedding(), 5, 31)
	residue := retryTestMarkers(t, face.RandomEmbedding(), 4, 32)

	result, err := w.start(FacesOptions{Threshold: 1})
	require.NoError(t, err)

	assert.Equal(t, 1, result.Added, "the first pass clusters the five")
	assert.Equal(t, 1, result.Retried, "and the second pass still runs for the four")

	formed := retryTestClusters(t, residue)
	require.Len(t, formed, 1)

	for id, members := range formed {
		assert.ElementsMatch(t, residue, members)
		assert.NotContains(t, retryTestClusters(t, first), id, "the two passes form different clusters")
	}
}

// TestFaces_MatchNewClusters covers the pass that makes a retry cluster real. Without it the
// cluster holds nothing and DeleteOrphanFaces removes it in the same run.
func TestFaces_MatchNewClusters(t *testing.T) {
	w := isolatedTestFaces(t, "facesmatchnewclusters")

	t.Run("NoClusters", func(t *testing.T) {
		result, err := w.MatchNewClusters(nil)
		require.NoError(t, err)
		assert.Zero(t, result.Updated)
	})
	t.Run("AttachesTheResidue", func(t *testing.T) {
		base := face.RandomEmbedding()
		markers := retryTestMarkers(t, base, 4, 41)

		src := rand.New(rand.NewPCG(42, 43)) //nolint:gosec // deterministic fixtures, not security
		f := entity.NewFace("", entity.SrcAuto, face.Embeddings{
			benchmarkEmbeddingAt(base, 0.02, src),
			benchmarkEmbeddingAt(base, 0.03, src),
		}, face.EmbeddingModelName())
		require.NotNil(t, f)
		require.NoError(t, f.Create())

		result, err := w.MatchNewClusters(entity.Faces{*f})
		require.NoError(t, err)
		assert.Positive(t, result.Updated)

		formed := retryTestClusters(t, markers)
		require.Len(t, formed, 1)
		assert.ElementsMatch(t, markers, formed[f.ID])
	})
	t.Run("LeavesAMarkerWithItsCloserCluster", func(t *testing.T) {
		// The force scan reads every marker, including ones an earlier pass already attached, so
		// this is what stops a retry cluster from taking a face that is closer to another.
		base := face.RandomEmbedding()
		markers := retryTestMarkers(t, base, 4, 44)

		src := rand.New(rand.NewPCG(45, 46)) //nolint:gosec // deterministic fixtures, not security
		near := entity.NewFace("", entity.SrcAuto, face.Embeddings{
			benchmarkEmbeddingAt(base, 0.01, src),
			benchmarkEmbeddingAt(base, 0.02, src),
		}, face.EmbeddingModelName())
		require.NotNil(t, near)
		require.NoError(t, near.Create())

		_, err := w.MatchNewClusters(entity.Faces{*near})
		require.NoError(t, err)
		require.Len(t, retryTestClusters(t, markers), 1)

		// A second cluster of the same person, offered afterwards: every marker already holds a
		// distance, and none of them may move to it.
		far := entity.NewFace("", entity.SrcAuto, face.Embeddings{
			benchmarkEmbeddingAt(base, 0.2, src),
			benchmarkEmbeddingAt(base, 0.21, src),
		}, face.EmbeddingModelName())
		require.NotNil(t, far)
		require.NoError(t, far.Create())

		_, err = w.MatchNewClusters(entity.Faces{*far})
		require.NoError(t, err)

		formed := retryTestClusters(t, markers)
		require.Len(t, formed, 1)
		assert.ElementsMatch(t, markers, formed[near.ID], "the closer cluster keeps every marker")
	})
}

// TestFaces_startClusterRetrySkippedWithoutAPass pins that the retry does not run where the first
// pass did not: the trigger exists so an idle worker does not scan and cluster on every wake, and
// a retry that ignored it would do both.
func TestFaces_startClusterRetrySkippedWithoutAPass(t *testing.T) {
	w := isolatedTestFaces(t, "facesclusterretryskipped")

	restore := face.ClusterCore
	t.Cleanup(func() { face.ClusterCore = restore })
	face.ClusterCore = face.ClusterCoreDefault

	w.conf.Options().FaceClusterCore = face.ClusterCoreDefault
	w.conf.Options().FaceClusterCoreRetry = 4
	require.Equal(t, 4, w.conf.FaceClusterCoreRetry())

	residue := retryTestMarkers(t, face.RandomEmbedding(), 4, 51)

	// A threshold no library reaches, which is the shape of an idle instance: the first pass
	// refuses to run, so there is nothing for a second one to work on either.
	result, err := w.start(FacesOptions{Threshold: 1000000})
	require.NoError(t, err)

	assert.Zero(t, result.Added)
	assert.Zero(t, result.Retried, "the retry inherits the decision not to run")
	assert.Empty(t, retryTestClusters(t, residue))
}

// TestFaces_startClusterRetryFollowsMatching pins the ordering rather than the outcome: the retry
// pass has to see a residue matching has already reduced.
//
// Running it before matching would let the lower core form a new cluster out of markers that an
// existing cluster was about to claim, which no cluster count would reveal.
func TestFaces_startClusterRetryFollowsMatching(t *testing.T) {
	w := isolatedTestFaces(t, "facesclusterretryorder")

	restore := face.ClusterCore
	t.Cleanup(func() { face.ClusterCore = restore })
	face.ClusterCore = face.ClusterCoreDefault

	w.conf.Options().FaceClusterCore = face.ClusterCoreDefault
	w.conf.Options().FaceClusterCoreRetry = 4
	require.Equal(t, 4, w.conf.FaceClusterCoreRetry())

	base := face.RandomEmbedding()

	// A cluster the library already holds, seeded from the same person.
	src := rand.New(rand.NewPCG(21, 22)) //nolint:gosec // deterministic fixtures, not security
	existing := entity.NewFace("", entity.SrcAuto, face.Embeddings{
		benchmarkEmbeddingAt(base, 0.02, src),
		benchmarkEmbeddingAt(base, 0.03, src),
		benchmarkEmbeddingAt(base, 0.04, src),
	}, face.EmbeddingModelName())
	require.NotNil(t, existing)
	require.NoError(t, existing.Create())

	// Four more faces of that person: enough for the retry core on their own, and near enough to
	// the cluster above for matching to claim them first.
	claimed := retryTestMarkers(t, base, 4, 23)

	result, err := w.start(FacesOptions{Force: true, Threshold: 1})
	require.NoError(t, err)

	formed := retryTestClusters(t, claimed)
	require.Len(t, formed, 1, "the four markers end up together either way, so only the cluster tells the order")

	for id, members := range formed {
		assert.Equal(t, existing.ID, id, "matching claimed them, so the retry pass saw nothing left")
		assert.ElementsMatch(t, claimed, members)
	}

	assert.Zero(t, result.Retried, "a pass running before matching would have formed a second cluster here")
}
