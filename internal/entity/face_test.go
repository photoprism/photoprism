package entity

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/ai/face"

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
		// is what keeps it from widening the gate to the stored 2.4.
		m := FaceFixtures.Get("joe-biden")
		match, dist := m.Match(MarkerFixtures.Pointer("1000003-4").Embeddings())

		assert.False(t, match)
		assert.Greater(t, dist, 1.31)
		assert.Less(t, dist, 1.32)
	})
	t.Run("Num1000003Six", func(t *testing.T) {
		m := FaceFixtures.Get("joe-biden")
		match, dist := m.Match(MarkerFixtures.Pointer("1000003-6").Embeddings())

		assert.False(t, match)
		assert.Greater(t, dist, 1.27)
		assert.Less(t, dist, 1.28)
	})
	t.Run("ClusterRadiusRaised", func(t *testing.T) {
		// A wider radius reaches the stored row without rewriting it, which is the point
		// of clamping against the live value instead of trusting the column.
		restore := face.ClusterRadius
		t.Cleanup(func() { face.ClusterRadius = restore })
		face.ClusterRadius = 0.95

		m := FaceFixtures.Get("joe-biden")
		match, dist := m.Match(MarkerFixtures.Pointer("1000003-4").Embeddings())

		assert.True(t, match)
		assert.Greater(t, dist, 1.31)
		assert.Less(t, dist, 1.32)
	})
	t.Run("LenEmbeddingsEqualZero", func(t *testing.T) {
		m := FaceFixtures.Get("joe-biden")
		match, dist := m.Match(face.Embeddings{})

		assert.False(t, match)
		assert.Equal(t, dist, float64(-1))
	})
	t.Run("LenEfacEmbeddingsEqualZero", func(t *testing.T) {
		m := NewFace("12345", SrcAuto, face.Embeddings{})
		match, dist := m.Match(MarkerFixtures.Pointer("1000003-6").Embeddings())

		assert.False(t, match)
		assert.Equal(t, dist, float64(-1))
	})
	t.Run("JaneDoeNoMatch", func(t *testing.T) {
		m := FaceFixtures.Get("jane-doe")
		match, _ := m.Match(MarkerFixtures.Pointer("1000003-5").Embeddings())

		assert.False(t, match)
	})
}

func TestFace_ResolveCollision(t *testing.T) {
	t.Run("Collision", func(t *testing.T) {
		// Both markers sit beyond the accept distance the shipped radius allows, so the
		// gate has to be widened for the collision bookkeeping below to be reachable.
		restore := face.ClusterRadius
		t.Cleanup(func() { face.ClusterRadius = restore })
		face.ClusterRadius = 0.95

		m := FaceFixtures.Get("joe-biden")

		assert.Zero(t, m.Collisions)
		assert.Zero(t, m.CollisionRadius)

		if reported, err := m.ResolveCollision(MarkerFixtures.Pointer("1000003-4").Embeddings()); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, reported)
		}

		// Number of collisions must have increased by one.
		assert.Equal(t, 1, m.Collisions)

		// Actual distance is ~1.314040
		assert.Greater(t, m.CollisionRadius, 1.2)
		assert.Less(t, m.CollisionRadius, 1.314)

		if reported, err := m.ResolveCollision(MarkerFixtures.Pointer("1000003-6").Embeddings()); err != nil {
			t.Fatal(err)
		} else {
			assert.True(t, reported)
		}

		// Number of collisions must not have increased.
		assert.Equal(t, 2, m.Collisions)

		// Actual distance is ~1.272604
		assert.Greater(t, m.CollisionRadius, 1.1)
		assert.Less(t, m.CollisionRadius, 1.272)
	})
	t.Run("SubjectIdEmpty", func(t *testing.T) {
		m := NewFace("", SrcAuto, face.RandomEmbeddings(2, face.RegularFace))
		if reported, err := m.ResolveCollision(MarkerFixtures.Pointer("1000003-4").Embeddings()); err != nil {
			t.Fatal(err)
		} else {
			assert.False(t, reported)
		}
	})
	t.Run("InvalidFaceId", func(t *testing.T) {
		m := NewFace("123", SrcAuto, face.Embeddings{})
		m.ID = ""
		if reported, err := m.ResolveCollision(MarkerFixtures.Pointer("1000003-4").Embeddings()); err == nil {
			t.Fatal(err)
		} else {
			assert.False(t, reported)
			assert.Equal(t, "invalid face id", err.Error())
		}
	})
	t.Run("EmbeddingEmpty", func(t *testing.T) {
		m := NewFace("123", SrcAuto, face.Embeddings{})
		m.EmbeddingJSON = []byte("")
		m.ID = "foo"
		if reported, err := m.ResolveCollision(MarkerFixtures.Pointer("1000003-4").Embeddings()); err == nil {
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

		r := NewFace("123", SrcAuto, e)
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

		err := m.SetEmbeddings(e)
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

		require.NoError(t, m.SetEmbeddings(embeddings))
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
		err := m.SetEmbeddings(face.Embeddings{make(face.Embedding, 8)})

		require.Error(t, err)
		assert.Contains(t, err.Error(), face.ModelFaceNet)
		assert.Contains(t, err.Error(), "faces migrate")
	})
}

func TestFace_Embedding(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := FaceFixtures.Get("joe-biden")

		assert.Equal(t, 0.10730543085474682, m.Embedding()[0])
	})
	t.Run("EmptyEmbedding", func(t *testing.T) {
		m := NewFace("12345", SrcAuto, face.Embeddings{})
		m.EmbeddingJSON = []byte("")

		assert.Empty(t, m.Embedding())
	})
	t.Run("InvalidEmbeddingJson", func(t *testing.T) {
		m := NewFace("12345", SrcAuto, face.Embeddings{})
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
		face.ClusterRadius, face.MatchDist = face.ThresholdMax, face.ThresholdMax

		m := &Face{SampleRadius: face.ThresholdMax}
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
		m := NewFace("uds5ttbeu5yj2sqf", SrcAuto, face.RandomEmbeddings(1, face.RegularFace))
		require.NoError(t, m.Create())
		require.NoError(t, m.UpdateMatchStats(4, 0.1))
		assert.Equal(t, 4, m.Samples)
		assert.InDelta(t, 0.1+face.Epsilon, m.SampleRadius, 1e-9)
	})
	t.Run("ClampsToClusterRadius", func(t *testing.T) {
		// The slack must not be able to lift the stored radius past the configured cap.
		m := NewFace("uds5ttbeu5yj2sqg", SrcAuto, face.RandomEmbeddings(1, face.RegularFace))
		require.NoError(t, m.Create())
		require.NoError(t, m.UpdateMatchStats(4, face.ClusterRadius))
		assert.InDelta(t, face.ClusterRadius, m.SampleRadius, 1e-9)
	})
	t.Run("NegativeDistance", func(t *testing.T) {
		m := NewFace("uds5ttbeu5yj2sqh", SrcAuto, face.RandomEmbeddings(1, face.RegularFace))
		require.NoError(t, m.Create())
		require.NoError(t, m.UpdateMatchStats(4, -1))
		assert.Zero(t, m.SampleRadius)
	})
}

func TestFace_UpdateMatchTime(t *testing.T) {
	m := NewFace("12345", SrcAuto, face.RandomEmbeddings(1, face.RegularFace))
	initialMatchTime := m.MatchedAt
	assert.Equal(t, initialMatchTime, m.MatchedAt)
	if err := m.Matched(); err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, initialMatchTime, m.MatchedAt)
}

func TestFace_Save(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		m := NewFace("dhsthrdst", SrcAuto, face.RandomEmbeddings(1, face.RegularFace))

		assert.Nil(t, FindFace(m.ID))

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, FindFace(m.ID))
		assert.Equal(t, "dhsthrdst", FindFace(m.ID).SubjUID)
	})
	t.Run("Error", func(t *testing.T) {
		m := NewFace("12345fde", SrcAuto, face.Embeddings{face.Embedding{1}, face.Embedding{2}})
		assert.Nil(t, FindFace(m.ID))
		assert.Error(t, m.Create())
		assert.Nil(t, FindFace(m.ID))
	})
}

func TestFace_Update(t *testing.T) {
	m := NewFace("12345fdef", SrcAuto, face.RandomEmbeddings(2, face.RegularFace))
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
		m := NewFace("12345unique", SrcAuto, face.RandomEmbeddings(1, face.RegularFace))
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
	m := NewFace("", SrcAuto, embeddings)
	require.NotNil(t, m)

	t.Run("SameModelMatches", func(t *testing.T) {
		match, dist := m.Match(embeddings)
		assert.True(t, match)
		assert.InDelta(t, 0, dist, 0.0001)
	})
	t.Run("OtherModelRefused", func(t *testing.T) {
		other := *m
		other.EmbedModel = face.ModelArcFaceR50
		match, dist := other.Match(embeddings)
		assert.False(t, match)
		assert.InDelta(t, -1, dist, 0.0001)
	})
}
