package entity

import (
	"math"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/pkg/rnd"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFace_TableName(t *testing.T) {
	m := &Face{}
	assert.Contains(t, m.TableName(), "faces")
}

func TestFace_Match(t *testing.T) {
	t.Run("Num1000003Four", func(t *testing.T) {
		// The fixture carries a radius from an earlier calibration, so the clamp on read
		// is what keeps it from widening the gate to the stored 2.
		m := FaceFixtures.Get("joe-biden")
		match, dist := m.Match(MarkerFixtures.Pointer("1000003-4").Embeddings(), face.EmbeddingModelName())

		assert.False(t, match)
		assert.Greater(t, dist, m.AcceptDist())
		assert.InDelta(t, face.AcceptDist(face.ClusterRadius), m.AcceptDist(), 1e-9)
	})
	t.Run("Num1000003Six", func(t *testing.T) {
		// Another person's marker, so it is beyond anything a configuration can accept.
		m := FaceFixtures.Get("joe-biden")
		match, dist := m.Match(MarkerFixtures.Pointer("1000003-6").Embeddings(), face.EmbeddingModelName())

		assert.False(t, match)
		assert.Greater(t, dist, float64(face.ConfigDistMax))
	})
	t.Run("ClusterRadiusRaised", func(t *testing.T) {
		// A wider radius reaches the stored row without rewriting it, which is the point
		// of clamping against the live value instead of trusting the column.
		m := FaceFixtures.Get("joe-biden")
		_, dist := m.Match(MarkerFixtures.Pointer("1000003-4").Embeddings(), face.EmbeddingModelName())
		require.Greater(t, dist, m.AcceptDist())

		restore := face.ClusterRadius
		t.Cleanup(func() { face.ClusterRadius = restore })
		face.ClusterRadius = dist - face.MatchDist + face.Epsilon

		match, raised := m.Match(MarkerFixtures.Pointer("1000003-4").Embeddings(), face.EmbeddingModelName())

		assert.True(t, match)
		assert.InDelta(t, dist, raised, 1e-9)
	})
	t.Run("LenEmbeddingsEqualZero", func(t *testing.T) {
		m := FaceFixtures.Get("joe-biden")
		match, dist := m.Match(face.Embeddings{}, face.EmbeddingModelName())

		assert.False(t, match)
		assert.Equal(t, dist, float64(-1))
	})
	t.Run("LenEfacEmbeddingsEqualZero", func(t *testing.T) {
		m := NewFace("12345", SrcAuto, face.Embeddings{}, face.EmbeddingModelName())
		match, dist := m.Match(MarkerFixtures.Pointer("1000003-6").Embeddings(), face.EmbeddingModelName())

		assert.False(t, match)
		assert.Equal(t, dist, float64(-1))
	})
	t.Run("OrderIndependentWithIncomparableVector", func(t *testing.T) {
		// A vector of another width yields -1, which used to win the minimum over every
		// real distance, so the same set matched or did not depending on its order.
		m := NewFace("", SrcAuto, face.Embeddings{face.RandomEmbedding()}, face.EmbeddingModelName())
		require.NotNil(t, m)

		near := m.Embedding()
		short := face.Embedding{0.1, 0.2}

		okShortFirst, distShortFirst := m.Match(face.Embeddings{short, near}, face.EmbeddingModelName())
		okNearFirst, distNearFirst := m.Match(face.Embeddings{near, short}, face.EmbeddingModelName())

		assert.Equal(t, okNearFirst, okShortFirst)
		assert.InDelta(t, distNearFirst, distShortFirst, 1e-9)
		assert.True(t, okShortFirst)
	})
	t.Run("JaneDoeNoMatch", func(t *testing.T) {
		m := FaceFixtures.Get("jane-doe")
		match, _ := m.Match(MarkerFixtures.Pointer("1000003-5").Embeddings(), face.EmbeddingModelName())

		assert.False(t, match)
	})
	t.Run("ClusterWithoutMagnitude", func(t *testing.T) {
		// Such a cluster is 1 away from every unit embedding, so it would accept whatever a
		// model reaching past 1 compares with it.
		m := NewFace("", SrcAuto, face.Embeddings{face.RandomEmbedding()}, face.EmbeddingModelName())
		require.NotNil(t, m)

		m.EmbeddingJSON = make(face.Embedding, len(m.Embedding())).JSON()
		m.embedding = nil

		match, dist := m.Match(face.Embeddings{face.RandomEmbedding()}, face.EmbeddingModelName())

		assert.False(t, match)
		assert.Equal(t, float64(-1), dist)
	})
	t.Run("NonFiniteVector", func(t *testing.T) {
		// A NaN distance is below every threshold it is compared with, so a corrupt vector
		// would match any cluster and be written to markers.face_dist as the best result.
		m := NewFace("", SrcAuto, face.Embeddings{face.RandomEmbedding()}, face.EmbeddingModelName())
		require.NotNil(t, m)

		nan := make(face.Embedding, len(m.Embedding()))
		nan[0] = math.NaN()

		match, dist := m.Match(face.Embeddings{nan}, face.EmbeddingModelName())

		assert.False(t, match)
		assert.Equal(t, float64(-1), dist)
	})
}

func TestFace_ResolveCollision(t *testing.T) {
	t.Run("Collision", func(t *testing.T) {
		m := FaceFixtures.Get("joe-biden")

		// Resolving a collision narrows the cluster and revises what it holds, so the row
		// goes back to what the fixture says before another test reads it.
		t.Cleanup(func() {
			f := FaceFixtures.Get("joe-biden")
			assert.NoError(t, m.Updates(Values{"collisions": f.Collisions, "collision_radius": f.CollisionRadius}))
		})

		far := MarkerFixtures.Pointer("1000003-4").Embeddings()
		farDist := far.Dist(m.Embedding())

		// The nearer collision has to stay outside the marker this cluster holds, or
		// revising its matches unlinks that marker and leaves the cluster an orphan.
		nearDist := 0.5 * face.AcceptDist(m.SampleRadius)
		near := face.Embeddings{face.FixtureEmbeddingAt(m.Embedding(), nearDist, 9001)}
		require.Greater(t, farDist, nearDist)
		require.Greater(t, nearDist, MarkerFixtures.Pointer("ms6sg6b14ahkyd24").Embeddings().Dist(m.Embedding()))

		// A collision is only reported for an embedding the cluster still accepts, and the
		// farther of the two sits outside what the shipped radius allows.
		restore := face.ClusterRadius
		t.Cleanup(func() { face.ClusterRadius = restore })
		face.ClusterRadius = farDist - face.MatchDist + face.Epsilon

		assert.Zero(t, m.Collisions)
		assert.Zero(t, m.CollisionRadius)

		if reported, err := m.ResolveCollision(far, face.EmbeddingModelName()); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, reported)
		}

		// Number of collisions must have increased by one.
		assert.Equal(t, 1, m.Collisions)
		assert.InDelta(t, farDist-face.Epsilon, m.CollisionRadius, 1e-9)

		if reported, err := m.ResolveCollision(near, face.EmbeddingModelName()); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, reported)
		}

		// A nearer collision narrows the radius rather than widening it.
		assert.Equal(t, 2, m.Collisions)
		assert.InDelta(t, nearDist-face.Epsilon, m.CollisionRadius, 1e-9)
		assert.Less(t, m.CollisionRadius, farDist-face.Epsilon)
	})
	t.Run("SubjectIdEmpty", func(t *testing.T) {
		m := NewFace("", SrcAuto, face.RandomEmbeddings(2, face.RegularFace), face.EmbeddingModelName())
		if reported, err := m.ResolveCollision(MarkerFixtures.Pointer("1000003-4").Embeddings(), face.EmbeddingModelName()); err != nil {
			t.Fatal(err)
		} else {
			assert.False(t, reported)
		}
	})
	t.Run("InvalidFaceId", func(t *testing.T) {
		m := NewFace("123", SrcAuto, face.Embeddings{}, face.EmbeddingModelName())
		m.ID = ""
		if reported, err := m.ResolveCollision(MarkerFixtures.Pointer("1000003-4").Embeddings(), face.EmbeddingModelName()); err == nil {
			t.Fatal(err)
		} else {
			assert.False(t, reported)
			assert.Equal(t, "invalid face id", err.Error())
		}
	})
	t.Run("EmbeddingEmpty", func(t *testing.T) {
		m := NewFace("123", SrcAuto, face.Embeddings{}, face.EmbeddingModelName())
		m.EmbeddingJSON = []byte("")
		m.ID = "foo"
		if reported, err := m.ResolveCollision(MarkerFixtures.Pointer("1000003-4").Embeddings(), face.EmbeddingModelName()); err == nil {
			t.Fatal(err)
		} else {
			assert.False(t, reported)
			assert.Equal(t, "embedding must not be empty", err.Error())
		}
	})
}

func TestFace_ReviseMatches(t *testing.T) {
	m := FaceFixtures.Get("joe-biden")
	removed, err := m.ReviseMatches()

	if err != nil {
		t.Fatal(err)
	}

	assert.Empty(t, removed)
}

func TestNewFace(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		marker := MarkerFixtures.Get("1000003-4")
		e := marker.Embeddings()

		r := NewFace("123", SrcAuto, e, face.EmbeddingModelName())
		assert.Equal(t, "", r.FaceSrc)
		assert.Equal(t, "123", r.SubjUID)
	})
}

func TestFace_MatchId(t *testing.T) {
	t.Run("ANum123BNum456", func(t *testing.T) {
		f1 := Face{ID: "A123"}
		f2 := Face{ID: "B456"}
		f3 := Face{ID: ""}

		assert.Equal(t, "A123-B456", f1.MatchId(f2))
		assert.Equal(t, "A123-B456", f2.MatchId(f1))
		assert.Equal(t, "", f3.MatchId(f1))
	})
}

func TestFace_SkipMatching(t *testing.T) {
	t.Run("Regular", func(t *testing.T) {
		m := FaceFixtures.Get("joe-biden")
		assert.False(t, m.SkipMatching())
	})
	t.Run("Ambiguous", func(t *testing.T) {
		// ResolveCollision is the only thing that raises the kind above RegularFace.
		m := FaceFixtures.Get("joe-biden")
		m.FaceKind = int(face.AmbiguousFace)
		assert.True(t, m.SkipMatching())
	})
}

func TestFace_SetEmbeddings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		marker := MarkerFixtures.Get("1000003-4")
		e := marker.Embeddings()
		m := FaceFixtures.Get("joe-biden")
		assert.NotEqual(t, e[0][0], m.Embedding()[0])

		err := m.SetEmbeddings(e, face.EmbeddingModelName())
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, e[0][0], m.Embedding()[0])
	})
	t.Run("CapsSampleRadius", func(t *testing.T) {
		embeddings := make(face.Embeddings, 2)
		for i := range embeddings {
			embeddings[i] = make(face.Embedding, len(face.NullEmbedding))
		}
		embeddings[0][0] = 1
		embeddings[1][0] = -1

		m := &Face{}

		require.NoError(t, m.SetEmbeddings(embeddings, face.EmbeddingModelName()))
		require.Equal(t, 2, m.Samples)
		assert.InDelta(t, face.ClusterRadius, m.SampleRadius, 1e-9)
	})
	t.Run("DimensionMismatch", func(t *testing.T) {
		restore := face.ConfiguredModel()

		t.Cleanup(func() {
			_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
		})

		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
			Name:  face.ModelFaceNet,
			Model: face.FindEmbeddingModel(face.ModelFaceNet),
		}))

		m := &Face{}
		err := m.SetEmbeddings(face.Embeddings{make(face.Embedding, 8)}, face.EmbeddingModelName())

		require.Error(t, err)
		assert.Contains(t, err.Error(), face.ModelFaceNet)
		assert.Contains(t, err.Error(), "faces migrate")
	})
}

func TestFace_Embedding(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// The fixtures are generated for whichever model a run resolves to, so what the
		// vector has to be is its width, not a particular value.
		m := FaceFixtures.Get("joe-biden")

		assert.Len(t, m.Embedding(), face.ExpectedDims())
		assert.InDelta(t, 0.0, m.Embedding().Dist(m.Embedding()), 1e-9)
	})
	t.Run("EmptyEmbedding", func(t *testing.T) {
		m := NewFace("12345", SrcAuto, face.Embeddings{}, face.EmbeddingModelName())
		m.EmbeddingJSON = []byte("")

		assert.Empty(t, m.Embedding())
	})
	t.Run("InvalidEmbeddingJson", func(t *testing.T) {
		m := NewFace("12345", SrcAuto, face.Embeddings{}, face.EmbeddingModelName())
		m.EmbeddingJSON = []byte("[false]")

		assert.Equal(t, float64(0), m.Embedding()[0])
	})
}

func TestFace_MatchMarkersEmpty(t *testing.T) {
	m := FaceFixtures.Get("joe-biden")
	require.NoError(t, m.MatchMarkers(nil))
	require.NoError(t, m.MatchMarkers([]string{}))
}

func TestFace_AcceptDist(t *testing.T) {
	t.Run("WithinClusterRadius", func(t *testing.T) {
		m := &Face{SampleRadius: 0.2}
		assert.InDelta(t, 0.2+face.MatchDist, m.AcceptDist(), 1e-9)
	})
	t.Run("StoredRadiusClamped", func(t *testing.T) {
		m := &Face{SampleRadius: 2}
		assert.InDelta(t, face.ClusterRadius+face.MatchDist, m.AcceptDist(), 1e-9)
	})
	t.Run("CappedAtCeiling", func(t *testing.T) {
		restoreRadius, restoreDist := face.ClusterRadius, face.MatchDist
		t.Cleanup(func() { face.ClusterRadius, face.MatchDist = restoreRadius, restoreDist })
		face.ClusterRadius, face.MatchDist = face.AcceptDistMax, face.AcceptDistMax

		m := &Face{SampleRadius: face.AcceptDistMax}
		assert.InDelta(t, face.AcceptDistMax, m.AcceptDist(), 1e-9)
	})
}

func TestFace_UpdateMatchStats(t *testing.T) {
	t.Run("NoFaceId", func(t *testing.T) {
		m := &Face{}
		require.NoError(t, m.UpdateMatchStats(3, 0.2))
		assert.Zero(t, m.Samples)
		assert.Zero(t, m.SampleRadius)
	})
	t.Run("NoSamples", func(t *testing.T) {
		m := FaceFixtures.Pointer("jane-doe")
		radius := m.SampleRadius
		require.NoError(t, m.UpdateMatchStats(0, 0.2))
		assert.Equal(t, radius, m.SampleRadius)
	})
	t.Run("AddsEpsilonSlack", func(t *testing.T) {
		m := NewFace("uds5ttbeu5yj2sqf", SrcAuto, face.RandomEmbeddings(1, face.RegularFace), face.EmbeddingModelName())
		require.NoError(t, m.Create())
		require.NoError(t, m.UpdateMatchStats(4, 0.1))
		assert.Equal(t, 4, m.Samples)
		assert.InDelta(t, 0.1+face.Epsilon, m.SampleRadius, 1e-9)
	})
	t.Run("ClampsToClusterRadius", func(t *testing.T) {
		// The slack must not be able to lift the stored radius past the configured cap.
		m := NewFace("uds5ttbeu5yj2sqg", SrcAuto, face.RandomEmbeddings(1, face.RegularFace), face.EmbeddingModelName())
		require.NoError(t, m.Create())
		require.NoError(t, m.UpdateMatchStats(4, face.ClusterRadius))
		assert.InDelta(t, face.ClusterRadius, m.SampleRadius, 1e-9)
	})
	t.Run("NeverNarrowsTheRadius", func(t *testing.T) {
		// A run visits only the markers that were unmatched when it started, so one newly
		// indexed face arriving near the centroid would otherwise rewrite the radius to its
		// own distance and refuse every member beyond it on the next pass.
		m := NewFace("uds5ttbeu5yj2sqi", SrcAuto, face.RandomEmbeddings(1, face.RegularFace), face.EmbeddingModelName())
		require.NoError(t, m.Create())
		require.NoError(t, m.UpdateMatchStats(20, 0.30))

		wide := m.SampleRadius
		accept := m.AcceptDist()
		require.InDelta(t, 0.30+face.Epsilon, wide, 1e-9)

		require.NoError(t, m.UpdateMatchStats(1, 0.05))

		assert.InDelta(t, wide, m.SampleRadius, 1e-9, "a single close match must not shrink the cluster")
		assert.InDelta(t, accept, m.AcceptDist(), 1e-9, "so the accept distance holds")
		assert.Equal(t, 20, m.Samples, "and the sample count is not replaced by the subset")
	})
	t.Run("StillWidens", func(t *testing.T) {
		// Growing is the whole point of the statistic: a farther member must still be able
		// to widen the cluster toward its clamp.
		m := NewFace("uds5ttbeu5yj2sqj", SrcAuto, face.RandomEmbeddings(1, face.RegularFace), face.EmbeddingModelName())
		require.NoError(t, m.Create())
		require.NoError(t, m.UpdateMatchStats(3, 0.10))
		require.NoError(t, m.UpdateMatchStats(4, 0.25))

		assert.InDelta(t, 0.25+face.Epsilon, m.SampleRadius, 1e-9)
		assert.Equal(t, 4, m.Samples)
	})
	t.Run("NegativeDistance", func(t *testing.T) {
		m := NewFace("uds5ttbeu5yj2sqh", SrcAuto, face.RandomEmbeddings(1, face.RegularFace), face.EmbeddingModelName())
		require.NoError(t, m.Create())
		require.NoError(t, m.UpdateMatchStats(4, -1))
		assert.Zero(t, m.SampleRadius)
	})
}

func TestFace_UpdateMatchTime(t *testing.T) {
	m := NewFace("12345", SrcAuto, face.RandomEmbeddings(1, face.RegularFace), face.EmbeddingModelName())
	initialMatchTime := m.MatchedAt
	assert.Equal(t, initialMatchTime, m.MatchedAt)
	if err := m.Matched(); err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, initialMatchTime, m.MatchedAt)
}

func TestFace_Save(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		m := NewFace("dhsthrdst", SrcAuto, face.RandomEmbeddings(1, face.RegularFace), face.EmbeddingModelName())

		assert.Nil(t, FindFace(m.ID))

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, FindFace(m.ID))
		assert.Equal(t, "dhsthrdst", FindFace(m.ID).SubjUID)
	})
	t.Run("Error", func(t *testing.T) {
		m := NewFace("12345fde", SrcAuto, face.Embeddings{face.Embedding{1}, face.Embedding{2}}, face.EmbeddingModelName())
		assert.Nil(t, FindFace(m.ID))
		assert.Error(t, m.Create())
		assert.Nil(t, FindFace(m.ID))
	})
}

func TestFace_Update(t *testing.T) {
	m := NewFace("12345fdef", SrcAuto, face.RandomEmbeddings(2, face.RegularFace), face.EmbeddingModelName())
	id := m.ID

	m.CreatedAt = time.Now()
	t.Logf("FaceID: %s", id)

	assert.Nil(t, FindFace(id))

	if err := m.Create(); err != nil {
		t.Fatal(err)
		return
	}

	assert.NotNil(t, FindFace(id))
	assert.Equal(t, "12345fdef", FindFace(m.ID).SubjUID)

	m2 := FindFace(m.ID)

	if err := m2.Update("SubjUID", "new"); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "new", FindFace(m.ID).SubjUID)
}

func TestFace_RefreshPhotos(t *testing.T) {
	f := FaceFixtures.Get("joe-biden")

	if err := f.RefreshPhotos(); err != nil {
		t.Fatal(err)
	}
}

func TestFirstOrCreateFace(t *testing.T) {
	t.Run("CreateNewFace", func(t *testing.T) {
		m := NewFace("12345unique", SrcAuto, face.RandomEmbeddings(1, face.RegularFace), face.EmbeddingModelName())
		r := FirstOrCreateFace(m)
		assert.Equal(t, "12345unique", r.SubjUID)
	})
	t.Run("ReturnExistingEntity", func(t *testing.T) {
		m := FaceFixtures.Pointer("joe-biden")
		r := FirstOrCreateFace(m)
		assert.Equal(t, "js6sg6b2h8njw0sx", r.SubjUID)
		assert.Equal(t, 33, r.Samples)
	})
}

func TestFindFace(t *testing.T) {
	t.Run("ExistingFace", func(t *testing.T) {
		assert.NotNil(t, FindFace("VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG7"))
		assert.Equal(t, 3, FindFace("VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG7").Samples)
	})
	t.Run("EmptyId", func(t *testing.T) {
		assert.Nil(t, FindFace(""))
	})
}

func TestFace_HideAndShow(t *testing.T) {
	f := FaceFixtures.Get("joe-biden")

	if err := f.Hide(); err != nil {
		t.Fatal(err)
	} else if err = f.Show(); err != nil {
		t.Fatal(err)
	}
}

func TestFace_SetSubjectUID(t *testing.T) {
	f := FindFace(FaceFixtures.Get("joe-biden").ID)
	assert.NotEmpty(t, f)

	if !assert.Empty(t, f.SetSubjectUID(SubjectFixtures.Get("jane-doe").SubjUID)) {
		return
	}

	f = FindFace(FaceFixtures.Get("joe-biden").ID)
	assert.NotEmpty(t, f)

	if !assert.Empty(t, f.SetSubjectUID(SubjectFixtures.Get("joe-biden").SubjUID)) {
		return
	}
}

func TestFace_SameEmbeddingModel(t *testing.T) {
	restore := face.ConfiguredModel()

	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})

	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
		Name:  face.ModelFaceNet,
		Model: face.FindEmbeddingModel(face.ModelFaceNet),
	}))

	t.Run("SameModel", func(t *testing.T) {
		m := &Face{EmbedModel: face.ModelFaceNet}
		assert.True(t, m.SameEmbeddingModel())
	})
	t.Run("NotRecorded", func(t *testing.T) {
		// Rows created before provenance was tracked are FaceNet-compatible.
		m := &Face{EmbedModel: ""}
		assert.True(t, m.SameEmbeddingModel())
	})
	t.Run("OtherModel", func(t *testing.T) {
		m := &Face{EmbedModel: face.ModelSFace}
		assert.False(t, m.SameEmbeddingModel())
	})
	t.Run("LegacyOtherModel", func(t *testing.T) {
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelSFace}))
		assert.False(t, (&Face{}).SameEmbeddingModel())
	})
}

func TestFace_MatchOtherModel(t *testing.T) {
	restore := face.ConfiguredModel()

	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})

	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
		Name:  face.ModelFaceNet,
		Model: face.FindEmbeddingModel(face.ModelFaceNet),
	}))

	embeddings := face.Embeddings{face.RandomEmbedding()}
	m := NewFace("", SrcAuto, embeddings, face.EmbeddingModelName())
	require.NotNil(t, m)

	t.Run("SameModelMatches", func(t *testing.T) {
		match, dist := m.Match(embeddings, face.EmbeddingModelName())
		assert.True(t, match)
		assert.InDelta(t, 0, dist, 0.0001)
	})
	t.Run("OtherModelRefused", func(t *testing.T) {
		other := *m
		other.EmbedModel = face.ModelArcFaceR50
		match, dist := other.Match(embeddings, face.EmbeddingModelName())
		assert.False(t, match)
		assert.InDelta(t, -1, dist, 0.0001)
	})
	// The argument carries its own provenance, so a vector from another 512-dim model must
	// be refused even though this cluster matches the configured one.
	t.Run("OtherModelArgumentRefused", func(t *testing.T) {
		match, dist := m.Match(embeddings, face.ModelArcFaceR50)
		assert.False(t, match)
		assert.InDelta(t, -1, dist, 0.0001)
	})
	t.Run("LegacyArgumentMatchesFaceNet", func(t *testing.T) {
		match, _ := m.Match(embeddings, "")
		assert.True(t, match)
	})
}

func TestFace_ReviseMatchesSkipsOtherModels(t *testing.T) {
	restore := face.ConfiguredModel()

	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})

	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{
		Name:  face.ModelFaceNet,
		Model: face.FindEmbeddingModel(face.ModelFaceNet),
	}))

	m := NewFace("", SrcAuto, face.Embeddings{face.RandomEmbedding()}, face.EmbeddingModelName())
	require.NotNil(t, m)
	require.NoError(t, m.Create())

	t.Cleanup(func() {
		UnscopedDb().Delete(&Face{}, "id = ?", m.ID)
	})

	// A marker from another embedding space, far from the cluster in any case.
	other := Marker{
		MarkerUID:      rnd.GenerateUID('m'),
		MarkerType:     MarkerFace,
		MarkerSrc:      SrcImage,
		FaceID:         m.ID,
		EmbeddingsJSON: face.Embeddings{face.RandomEmbedding()}.JSON(),
		EmbedModel:     face.ModelArcFaceR50,
	}

	require.NoError(t, Db().Create(&other).Error)

	t.Cleanup(func() {
		UnscopedDb().Delete(&Marker{}, "marker_uid = ?", other.MarkerUID)
	})

	revised, err := m.ReviseMatches()
	require.NoError(t, err)

	for _, r := range revised {
		assert.NotEqual(t, other.MarkerUID, r.MarkerUID, "an incomparable marker must not be cleared")
	}

	stored := Marker{}
	require.NoError(t, UnscopedDb().First(&stored, "marker_uid = ?", other.MarkerUID).Error)
	assert.Equal(t, m.ID, stored.FaceID, "the assignment must survive a revision it could not evaluate")
}

func TestFace_MatchMarkers(t *testing.T) {
	cluster := FaceFixtures.Pointer("joe-biden")

	// newFacelessMarker persists an unassigned marker well inside what the cluster accepts, so
	// only the size bound can keep it out.
	newFacelessMarker := func(t *testing.T, size int, seed uint64) *Marker {
		t.Helper()

		m := &Marker{
			FileUID:    "fs6sg6bw45bnlqdw",
			MarkerType: MarkerFace,
			MarkerSrc:  SrcImage,
			Size:       size,
			Score:      50,
			X:          0.1,
			Y:          0.1,
			W:          0.1,
			H:          0.1,
		}

		at := face.FixtureEmbeddingAt(cluster.Embedding(), 0.2*cluster.AcceptDist(), seed)
		m.SetEmbeddings(face.Embeddings{at}, cluster.EmbedModel, face.DetectorYuNet)

		require.NoError(t, Db().Create(m).Error)
		t.Cleanup(func() { Db().Delete(m) })

		return m
	}
	t.Run("AdmitsAnOrdinaryMarker", func(t *testing.T) {
		m := newFacelessMarker(t, face.SizeThreshold, 9101)

		require.NoError(t, cluster.MatchMarkers(Faceless))

		found := FindMarker(m.MarkerUID)
		require.NotNil(t, found)
		assert.Equal(t, cluster.ID, found.FaceID)
	})
	t.Run("RepointsASmallMarkerThatIsAlreadyClustered", func(t *testing.T) {
		// The merge path calls this to move markers off clusters it is about to purge, so the
		// size bound must not reach them: one left behind would point at a deleted cluster.
		m := newFacelessMarker(t, face.SizeThreshold-1, 9104)

		other := NewFace(cluster.SubjUID, SrcAuto, face.Embeddings{cluster.Embedding()}, cluster.EmbedModel)
		require.NotNil(t, other)
		other = FirstOrCreateFace(other)
		require.NotNil(t, other)

		require.NoError(t, m.Update("FaceID", other.ID))
		require.NoError(t, cluster.MatchMarkers([]string{other.ID}))

		found := FindMarker(m.MarkerUID)
		require.NotNil(t, found)
		assert.Equal(t, cluster.ID, found.FaceID, "a clustered marker is re-pointed whatever its size")
	})
	t.Run("RefusesAMarkerBelowTheDetectionFloor", func(t *testing.T) {
		// Only the second detection pass produces one, and it exists to mark a face a crowd
		// photograph would otherwise lose rather than to name a person from it.
		m := newFacelessMarker(t, face.SizeThreshold-1, 9102)

		require.NoError(t, cluster.MatchMarkers(Faceless))

		found := FindMarker(m.MarkerUID)
		require.NotNil(t, found)
		assert.Empty(t, found.FaceID)
	})
}
