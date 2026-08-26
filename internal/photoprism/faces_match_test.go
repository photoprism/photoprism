package photoprism

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// setMatchMargin applies an assignment margin for the duration of a test.
func setMatchMargin(t *testing.T, margin float64) {
	restore := face.MatchMargin

	t.Cleanup(func() { face.MatchMargin = restore })

	face.MatchMargin = margin
}

// TestFaces_Match exercises the end-to-end matching flow with a loaded test configuration.
func TestFaces_Match(t *testing.T) {
	c := config.TestConfig()

	m := NewFaces(c)

	opt := FacesOptions{
		Force:     true,
		Threshold: 1,
	}

	r, err := m.Match(opt)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(r)
}

// TestFacesMatchResult_Add covers the totals a run accumulates across its two passes. Ambiguous
// is the one an operator reads to judge the match margin, so a pass that dropped it would report
// a run that merely recognized less.
func TestFacesMatchResult_Add(t *testing.T) {
	r := FacesMatchResult{Updated: 1, Recognized: 2, Unknown: 3, Ambiguous: 4}

	r.Add(FacesMatchResult{Updated: 10, Recognized: 20, Unknown: 30, Ambiguous: 40})

	assert.Equal(t, FacesMatchResult{Updated: 11, Recognized: 22, Unknown: 33, Ambiguous: 44}, r)

	r.Add(FacesMatchResult{})

	assert.Equal(t, FacesMatchResult{Updated: 11, Recognized: 22, Unknown: 33, Ambiguous: 44}, r)
}

// TestRecordFaceMatch covers the per-run statistics a match pass accumulates for each cluster.
func TestRecordFaceMatch(t *testing.T) {
	newFace := func(t *testing.T) *entity.Face {
		t.Helper()
		f := entity.NewFace("", entity.SrcAuto, face.RandomEmbeddings(3, face.RegularFace), face.EmbeddingModelName())
		require.NotNil(t, f)
		require.NotEmpty(t, f.ID)
		return f
	}

	t.Run("Success", func(t *testing.T) {
		stats := make(map[string]*faceMatchStats)
		f := newFace(t)

		recordFaceMatch(stats, f, 0.2)
		recordFaceMatch(stats, f, 0.5)
		recordFaceMatch(stats, f, 0.3)

		require.Len(t, stats, 1)
		assert.Equal(t, 3, stats[f.ID].matched)
		assert.InDelta(t, 0.5, stats[f.ID].maxDist, 1e-9)
		assert.Same(t, f, stats[f.ID].face)
	})
	t.Run("SameRowThroughTwoPointers", func(t *testing.T) {
		// Faces.Match runs two passes over different slices, so one database row reaches this
		// as two pointers. Keyed by pointer, only one of the two entries would be written back.
		stats := make(map[string]*faceMatchStats)
		first := newFace(t)
		second := *first

		recordFaceMatch(stats, first, 0.2)
		recordFaceMatch(stats, &second, 0.4)

		require.Len(t, stats, 1)
		assert.Equal(t, 2, stats[first.ID].matched)
		assert.InDelta(t, 0.4, stats[first.ID].maxDist, 1e-9)
	})
	t.Run("InvalidInput", func(t *testing.T) {
		stats := make(map[string]*faceMatchStats)
		f := newFace(t)
		unsaved := &entity.Face{}

		recordFaceMatch(stats, nil, 0.2)
		recordFaceMatch(stats, unsaved, 0.2)
		recordFaceMatch(stats, f, -1)

		assert.Empty(t, stats)
	})
}

// TestBuildFaceCandidates validates that we drop non-matchable faces when building the index.
func TestBuildFaceCandidates(t *testing.T) {
	regular := entity.NewFace("", entity.SrcAuto, face.RandomEmbeddings(3, face.RegularFace), face.EmbeddingModelName())
	require.NotNil(t, regular)

	// A cluster from another embedding space must never be compared with the current one.
	stale := *regular
	stale.ID = "stale-model"
	stale.EmbedModel = otherFaceModel(t, face.ConfiguredModel())

	// ResolveCollision raises the kind, which excludes a cluster from automatic matching.
	ambiguous := *regular
	ambiguous.ID = "ambiguous"
	ambiguous.FaceKind = int(face.AmbiguousFace)

	// A cluster with no magnitude is 1 from every marker, so it would capture everything a model
	// accepting past 1 compares with it.
	zero := entity.Face{
		ID:            "zero-magnitude",
		EmbedModel:    regular.EmbedModel,
		EmbeddingJSON: make(face.Embedding, len(regular.Embedding())).JSON(),
	}

	faces := entity.Faces{*regular, stale, ambiguous, zero}

	index := buildFaceIndex(faces)

	require.Len(t, index.candidates, 1)
	require.Equal(t, regular.ID, index.candidates[0].ref.ID)

	// The candidate caches the clamped cutoff, not the raw column, so a stored radius
	// from an earlier calibration cannot widen the gate for a whole match run.
	require.InDelta(t, regular.AcceptDist(), index.candidates[0].acceptDist, 1e-9)
}

// TestFaceCandidateLimit covers the accept distance and collision gates in isolation. The limit
// is what selection bounds each comparison by, so it decides both what a candidate accepts and
// how early one that cannot win is abandoned.
func TestFaceCandidateLimit(t *testing.T) {
	embeddings := face.RandomEmbeddings(1, face.RegularFace)
	ref := entity.NewFace("", entity.SrcAuto, embeddings, face.EmbeddingModelName())
	require.NotNil(t, ref)

	t.Run("AcceptDist", func(t *testing.T) {
		c := faceCandidate{ref: ref, emb: ref.Embedding(), acceptDist: face.AcceptDist(0)}
		assert.InDelta(t, face.AcceptDist(0), c.limit(), 1e-9)
	})
	t.Run("CollisionRadiusBelowAcceptDist", func(t *testing.T) {
		// A cluster that was separated from another by a collision keeps a narrowed radius,
		// and honoring it during matching is what stops the two people re-merging.
		c := faceCandidate{ref: ref, emb: ref.Embedding(), acceptDist: 1.0,
			collisionRadius: face.CollisionDist * 2}
		require.Greater(t, c.collisionRadius, face.CollisionDist, "the gate only applies above the collision distance")
		assert.InDelta(t, c.collisionRadius, c.limit(), 1e-9)
	})
	t.Run("CollisionRadiusAboveAcceptDist", func(t *testing.T) {
		// The wider of the two never wins: a collision radius may only narrow the gate.
		c := faceCandidate{ref: ref, emb: ref.Embedding(), acceptDist: 0.5, collisionRadius: 1.0}
		assert.InDelta(t, 0.5, c.limit(), 1e-9)
	})
	t.Run("CollisionRadiusUnmeasured", func(t *testing.T) {
		// Below the collision distance the radius is not yet meaningful and is ignored.
		c := faceCandidate{ref: ref, emb: ref.Embedding(), acceptDist: 1.0,
			collisionRadius: face.CollisionDist / 2}
		assert.InDelta(t, 1.0, c.limit(), 1e-9)
	})
}

// TestSelectBestFaceGates covers the cases where no candidate may be returned at all.
func TestSelectBestFaceGates(t *testing.T) {
	embeddings := face.RandomEmbeddings(1, face.RegularFace)
	ref := entity.NewFace("", entity.SrcAuto, embeddings, face.EmbeddingModelName())
	require.NotNil(t, ref)

	t.Run("TooFar", func(t *testing.T) {
		idx := faceIndex{candidates: []faceCandidate{{ref: ref, emb: ref.Embedding(), acceptDist: -1}}}
		best, dist, _ := selectBestFace(embeddings, idx)
		assert.Nil(t, best)
		assert.InDelta(t, -1.0, dist, 1e-9)
	})
	t.Run("OutsideCollisionRadius", func(t *testing.T) {
		other := face.RandomEmbeddings(1, face.RegularFace)
		d := other.Dist(ref.Embedding())
		require.Positive(t, d, "the two random faces must be far enough apart to test with")

		idx := faceIndex{candidates: []faceCandidate{{
			ref:             ref,
			emb:             ref.Embedding(),
			acceptDist:      d + 1,
			collisionRadius: d / 2,
		}}}

		best, _, _ := selectBestFace(other, idx)
		assert.Nil(t, best, "a face beyond the collision radius must be refused")
	})
	t.Run("NoEmbeddings", func(t *testing.T) {
		idx := faceIndex{candidates: []faceCandidate{{ref: ref, emb: ref.Embedding(), acceptDist: face.AcceptDist(0)}}}
		best, dist, _ := selectBestFace(face.Embeddings{}, idx)
		assert.Nil(t, best)
		assert.InDelta(t, -1.0, dist, 1e-9)
	})
	t.Run("FirstWinsOnTie", func(t *testing.T) {
		// Selection bounds each comparison by the best distance found so far, so an equally
		// close candidate must not displace the one already chosen. The margin would refuse
		// both, which is a separate rule and is covered by TestSelectBestFaceMargin.
		setMatchMargin(t, 0)

		first := &entity.Face{ID: "first"}
		second := &entity.Face{ID: "second"}
		emb := ref.Embedding()

		idx := faceIndex{candidates: []faceCandidate{
			{ref: first, emb: emb, acceptDist: face.AcceptDist(0)},
			{ref: second, emb: emb, acceptDist: face.AcceptDist(0)},
		}}

		best, _, _ := selectBestFace(embeddings, idx)
		require.NotNil(t, best)
		assert.Equal(t, "first", best.ID)
	})
}

// TestSelectBestFace ensures the best candidate is returned after indexing.
func TestSelectBestFace(t *testing.T) {
	markerEmb := face.RandomEmbeddings(1, face.RegularFace)

	matchFace := entity.NewFace("", entity.SrcAuto, markerEmb, face.EmbeddingModelName())
	require.NotNil(t, matchFace)

	// Force a different face that should not be a better match.
	otherEmb := face.RandomEmbeddings(4, face.RegularFace)
	otherFace := entity.NewFace("", entity.SrcAuto, otherEmb, face.EmbeddingModelName())
	require.NotNil(t, otherFace)

	faces := entity.Faces{*matchFace, *otherFace}

	index := buildFaceIndex(faces)
	require.Len(t, index.candidates, 2)

	best, dist, _ := selectBestFace(markerEmb, index)
	require.NotNil(t, best)
	require.Equal(t, matchFace.ID, best.ID)
	require.InDelta(t, 0.0, dist, 1e-9)
}

// TestBenchmarkEmbeddingAt checks the fixture helper places a vector where it claims. The
// closest-wins test reads distances off these shells, so a helper that drifted would weaken
// that test without failing it.
func TestBenchmarkEmbeddingAt(t *testing.T) {
	rnd := rand.New(rand.NewPCG(21, 22)) //nolint:gosec // deterministic fixtures, not security
	base := face.RandomEmbedding()

	for _, d := range []float64{0.05, 0.15, 0.39, 0.6, 1.0, 1.41} {
		e := benchmarkEmbeddingAt(base, d, rnd)
		assert.InDelta(t, d, base.Dist(e), 1e-9, "must land at the requested distance")
		assert.InDelta(t, 1.0, e.Magnitude(), 1e-9, "must stay a unit vector")
	}

	t.Run("BeyondDiameter", func(t *testing.T) {
		// Two unit vectors cannot be more than 2 apart, so a larger request clamps.
		assert.InDelta(t, 2.0, base.Dist(benchmarkEmbeddingAt(base, 2.5, rnd)), 1e-9)
	})
}

// TestSelectBestFaceReturnsClosest pins the invariant that a search narrowed by anything other
// than distance cannot hold: the marker gets the closest cluster that accepts it, not merely an
// acceptable one. A merely acceptable cluster is then widened to the inflated distance by
// UpdateMatchStats, so a near miss here does not stay a near miss.
func TestSelectBestFaceReturnsClosest(t *testing.T) {
	// The margin refuses a winner rather than choosing a different one, so it is switched off
	// here and pinned separately: with it on, the reference and the answer agree by returning
	// nothing, which would weaken this test without failing it.
	setMatchMargin(t, 0)

	index, base := benchmarkFaceIndex([]float64{0.15, 0.3, 0.45, 0.6, 0.9, 1.1, 1.3})
	require.Len(t, index.candidates, benchmarkCandidateCount)

	// Clusters built from one sample all share the same gate, which would make the bound
	// selection lowers to the running best indistinguishable from one that widens it. Real
	// clusters differ, so spread the gates and narrow a few by a measured collision.
	for i := range index.candidates {
		c := &index.candidates[i]
		c.acceptDist = 0.2 + float64(i%9)*0.1

		if i%7 == 0 {
			c.collisionRadius = face.CollisionDist + float64(i%5)*0.05
		}
	}

	limits := make(map[float64]struct{}, len(index.candidates))
	for i := range index.candidates {
		limits[index.candidates[i].limit()] = struct{}{}
	}
	require.Greater(t, len(limits), 5, "the candidates must not all share one gate")

	rnd := rand.New(rand.NewPCG(7, 8)) //nolint:gosec // deterministic fixtures, not security

	var accepted int

	for i := range 200 {
		markerEmb := face.Embeddings{benchmarkEmbeddingAt(base, 0.05+float64(i%12)*0.05, rnd)}

		// The reference scans every candidate to completion and applies no bound at all.
		want, wantDist := legacySelectBestFace(markerEmb, index)
		got, gotDist, _ := selectBestFace(markerEmb, index)

		if want == nil {
			assert.Nil(t, got, "nothing accepts the marker, so nothing may be returned")
			continue
		}

		accepted++

		require.NotNil(t, got)
		assert.Equal(t, want.ID, got.ID, "the closest accepting cluster must win")
		assert.InDelta(t, wantDist, gotDist, 1e-9)
	}

	require.Positive(t, accepted, "the fixture must produce matches for the comparison to mean anything")
}

// TestSelectBestFaceMargin covers the assignment margin, which is what a face lying between two
// people runs into: it is large, sharp and confidently detected, so nothing on the detection side
// selects against it and only the gap between the two competing distances does.
func TestSelectBestFaceMargin(t *testing.T) {
	setMatchMargin(t, face.MatchMarginDefault)

	// One cluster per person, and a marker placed a chosen distance from each.
	base := face.FixtureEmbedding(4001)
	near := &entity.Face{ID: "near", EmbedModel: face.EmbeddingModelName()}
	far := &entity.Face{ID: "far", EmbedModel: face.EmbeddingModelName()}

	selectAt := func(t *testing.T, nearDist, farDist float64) (*entity.Face, float64, bool) {
		t.Helper()

		marker := face.Embeddings{base}
		idx := faceIndex{candidates: []faceCandidate{
			{ref: near, emb: face.FixtureEmbeddingAt(base, nearDist, 4002), acceptDist: face.AcceptDistMax},
			{ref: far, emb: face.FixtureEmbeddingAt(base, farDist, 4003), acceptDist: face.AcceptDistMax},
		}}

		return selectBestFace(marker, idx)
	}

	t.Run("BetweenTwoClusters", func(t *testing.T) {
		best, dist, ambiguous := selectAt(t, 0.70, 0.72)
		assert.Nil(t, best, "a marker equidistant between two clusters must not be assigned to either")
		assert.InDelta(t, -1.0, dist, 1e-9)
		assert.True(t, ambiguous, "the caller has to tell this apart from a marker nothing accepted")
	})
	t.Run("ClearlyNearerClusterStillWins", func(t *testing.T) {
		// The companion direction, so the margin cannot regress into refusing everything.
		best, dist, ambiguous := selectAt(t, 0.30, 0.90)
		require.NotNil(t, best)
		assert.Equal(t, "near", best.ID)
		assert.InDelta(t, 0.30, dist, 0.01)
		assert.False(t, ambiguous)
	})
	t.Run("RunnerUpOutsideItsOwnGate", func(t *testing.T) {
		// A cluster that would refuse the marker anyway is no competitor for it.
		marker := face.Embeddings{base}
		idx := faceIndex{candidates: []faceCandidate{
			{ref: near, emb: face.FixtureEmbeddingAt(base, 0.70, 4002), acceptDist: face.AcceptDistMax},
			{ref: far, emb: face.FixtureEmbeddingAt(base, 0.72, 4003), acceptDist: 0.5},
		}}

		best, _, _ := selectBestFace(marker, idx)
		require.NotNil(t, best)
		assert.Equal(t, "near", best.ID)
	})
	t.Run("SameSubjectIsExempt", func(t *testing.T) {
		// A person may own several clusters, so whichever of two of theirs wins names the
		// same face and the marker must not be refused for a difference that does not matter.
		near.SubjUID = "ps6sg6be2lvl0y11"
		far.SubjUID = "ps6sg6be2lvl0y11"

		t.Cleanup(func() { near.SubjUID, far.SubjUID = "", "" })

		best, _, _ := selectAt(t, 0.70, 0.72)
		require.NotNil(t, best)
		assert.Equal(t, "near", best.ID)
	})
	t.Run("DifferentSubjects", func(t *testing.T) {
		near.SubjUID = "ps6sg6be2lvl0y11"
		far.SubjUID = "ps6sg6be2lvl0y12"

		t.Cleanup(func() { near.SubjUID, far.SubjUID = "", "" })

		best, _, _ := selectAt(t, 0.70, 0.72)
		assert.Nil(t, best)
	})
	t.Run("Disabled", func(t *testing.T) {
		setMatchMargin(t, 0)

		best, _, _ := selectAt(t, 0.70, 0.72)
		require.NotNil(t, best)
		assert.Equal(t, "near", best.ID)
	})
}

// TestAmbiguousBestFace covers the exemptions the margin applies before it refuses an assignment.
func TestAmbiguousBestFace(t *testing.T) {
	setMatchMargin(t, face.MatchMarginDefault)

	anon := &entity.Face{ID: "anon"}
	named := &entity.Face{ID: "named", SubjUID: "ps6sg6be2lvl0y11"}
	other := &entity.Face{ID: "other", SubjUID: "ps6sg6be2lvl0y12"}
	same := &entity.Face{ID: "same", SubjUID: "ps6sg6be2lvl0y11"}

	t.Run("NoRunnerUp", func(t *testing.T) {
		assert.False(t, ambiguousBestFace(anon, nil, 0.7, -1))
	})
	t.Run("ClearWinner", func(t *testing.T) {
		assert.False(t, ambiguousBestFace(anon, named, 0.3, 0.9))
	})
	t.Run("Anonymous", func(t *testing.T) {
		// Nothing says the two hold the same person, so the coin toss stands.
		assert.True(t, ambiguousBestFace(anon, named, 0.7, 0.72))
	})
	t.Run("DifferentSubjects", func(t *testing.T) {
		assert.True(t, ambiguousBestFace(named, other, 0.7, 0.72))
	})
	t.Run("SameSubject", func(t *testing.T) {
		assert.False(t, ambiguousBestFace(named, same, 0.7, 0.72))
	})
}

func TestFacesMatchRespectsVeto(t *testing.T) {
	c := config.TestConfig()
	w := NewFaces(c)

	var marker entity.Marker
	require.NoError(t, entity.Db().Where("marker_type = ? AND marker_invalid = 0 AND face_id <> ''", entity.MarkerFace).Take(&marker).Error)

	origFaceID := marker.FaceID
	require.NotEqual(t, "", origFaceID)

	var f entity.Face
	require.NoError(t, entity.Db().Where("id = ?", origFaceID).Take(&f).Error)

	_, err := marker.ClearFace()
	require.NoError(t, err)

	stats := make(map[string]*faceMatchStats)
	faces := entity.Faces{f}

	// The cluster has to be eligible and actually accept this marker, or the veto is not what
	// keeps it unassigned. The fixture radius is narrower than the marker it holds, so widen
	// it to what a cluster containing that marker would really have recorded.
	require.NotEmpty(t, buildFaceIndex(faces).candidates, "the fixture cluster must be matchable")

	// The fixture marker carries the same vector as its cluster, so the distance is zero.
	reach := marker.Embeddings().Dist(f.Embedding())
	require.GreaterOrEqual(t, reach, 0.0, "the marker must be comparable with its cluster")

	faces[0].SampleRadius = face.ClampSampleRadius(reach + face.Epsilon)
	require.LessOrEqual(t, reach, faces[0].AcceptDist(), "the cluster must accept its own marker")

	// ClearFace stamps matched_at, while a non-force run only visits markers that carry none,
	// so the marker has to look unmatched again or MatchFaces never reaches the veto and the
	// assertion below passes on a marker it never saw.
	unmatch := func(t *testing.T) {
		t.Helper()
		require.NoError(t, entity.UnscopedDb().Model(&entity.Marker{}).
			Where("marker_uid = ?", marker.MarkerUID).
			UpdateColumn("matched_at", nil).Error)
	}

	unmatch(t)

	w.rememberVeto(marker.MarkerUID)
	_, err = w.MatchFaces(faces, false, nil, stats)
	require.NoError(t, err)

	require.NoError(t, entity.Db().Where("marker_uid = ?", marker.MarkerUID).Take(&marker).Error)
	require.Equal(t, "", marker.FaceID, "a vetoed marker must not be re-assigned")

	// Positive control: the same call without the veto has to re-assign the marker, or the
	// assertion above holds for a reason that has nothing to do with the veto.
	w.clearVeto(marker.MarkerUID)
	unmatch(t)

	_, err = w.MatchFaces(faces, false, nil, stats)
	require.NoError(t, err)

	require.NoError(t, entity.Db().Where("marker_uid = ?", marker.MarkerUID).Take(&marker).Error)
	assert.Equal(t, origFaceID, marker.FaceID, "a marker that is not vetoed must be re-assigned")

	// restore original assignment to keep fixtures consistent
	dist := marker.Embeddings().Dist(f.Embedding())
	_, err = marker.SetFace(&f, dist)
	require.NoError(t, err)
}

// TestStampMatchedFaces pins that the stamping loop leaves a cluster a collision reopened during
// the pass unmatched.
//
// Stamping it would end the only route back: the next run reads clusters that are still
// unmatched, so a cluster stamped here is not re-examined, and the markers ReviseMatches dropped
// have nothing to be compared against. Every cluster the loop iterates started out unmatched, so
// the timestamp cannot tell the two apart and the flag has to.
func TestStampMatchedFaces(t *testing.T) {
	stamped := entity.Face{ID: "REOPENTESTSTAMPED", MatchedAt: entity.TimeStamp()}
	reopened := entity.Face{ID: "REOPENTESTREOPEN", MatchedAt: entity.TimeStamp()}

	require.NoError(t, entity.UnscopedDb().Create(&stamped).Error)
	require.NoError(t, entity.UnscopedDb().Create(&reopened).Error)
	t.Cleanup(func() {
		entity.UnscopedDb().Delete(&entity.Face{}, "id IN (?)", []string{stamped.ID, reopened.ID})
	})

	// What a pass looks like on the way out: one cluster untouched, one reopened by a collision.
	faces := entity.Faces{stamped, reopened}
	require.NoError(t, entity.UnscopedDb().Model(&faces[0]).UpdateColumn("matched_at", nil).Error)
	require.NoError(t, entity.UnscopedDb().Model(&faces[1]).UpdateColumn("matched_at", nil).Error)
	faces[0].MatchedAt = nil
	faces[1].ReopenForTest()

	stampMatchedFaces(faces)

	var after entity.Faces
	require.NoError(t, entity.UnscopedDb().Find(&after, "id IN (?)", []string{stamped.ID, reopened.ID}).Error)
	require.Len(t, after, 2)

	for _, f := range after {
		switch f.ID {
		case stamped.ID:
			assert.NotNil(t, f.MatchedAt, "an untouched cluster is stamped as compared")
		case reopened.ID:
			assert.Nil(t, f.MatchedAt, "a reopened cluster stays unmatched so the next run reads it")
		}
	}
}
