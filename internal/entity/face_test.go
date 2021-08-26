package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFace_TableName(t *testing.T) {
	m := &Face{}
	assert.Contains(t, m.TableName(), "faces")
}

func TestFace_Match(t *testing.T) {
	t.Run("1000003-4", func(t *testing.T) {
		m := FaceFixtures.Get("joe-biden")
		match, dist := m.Match(MarkerFixtures.Pointer("1000003-4").Embeddings())

		assert.True(t, match)
		assert.Greater(t, dist, 1.31)
		assert.Less(t, dist, 1.32)
	})

	t.Run("1000003-6", func(t *testing.T) {
		m := FaceFixtures.Get("joe-biden")
		match, dist := m.Match(MarkerFixtures.Pointer("1000003-6").Embeddings())

		assert.True(t, match)
		assert.Greater(t, dist, 1.27)
		assert.Less(t, dist, 1.28)
	})

	t.Run("len(embeddings) == 0", func(t *testing.T) {
		m := FaceFixtures.Get("joe-biden")
		match, dist := m.Match(Embeddings{})

		assert.False(t, match)
		assert.Equal(t, dist, float64(-1))
	})
	t.Run("len(efacEmbeddings) == 0", func(t *testing.T) {
		m := NewFace("12345", SrcAuto, Embeddings{})
		match, dist := m.Match(MarkerFixtures.Pointer("1000003-6").Embeddings())

		assert.False(t, match)
		assert.Equal(t, dist, float64(-1))
	})
	t.Run("jane doe- no match", func(t *testing.T) {
		m := FaceFixtures.Get("jane-doe")
		match, _ := m.Match(MarkerFixtures.Pointer("1000003-5").Embeddings())

		assert.False(t, match)
	})
}

func TestFace_ReportCollision(t *testing.T) {
	m := FaceFixtures.Get("joe-biden")

	assert.Zero(t, m.Collisions)
	assert.Zero(t, m.CollisionRadius)

	if reported, err := m.ReportCollision(MarkerFixtures.Pointer("1000003-4").Embeddings()); err != nil {
		t.Fatal(err)
	} else {
		assert.True(t, reported)
	}

	// Number of collisions must have increased by one.
	assert.Equal(t, 1, m.Collisions)

	// Actual distance is ~1.314040
	assert.Greater(t, m.CollisionRadius, 1.2)
	assert.Less(t, m.CollisionRadius, 1.314)

	if reported, err := m.ReportCollision(MarkerFixtures.Pointer("1000003-6").Embeddings()); err != nil {
		t.Fatal(err)
	} else {
		assert.False(t, reported)
	}

	// Number of collisions must not have increased.
	assert.Equal(t, 1, m.Collisions)

	// Actual distance is ~1.272604
	assert.Greater(t, m.CollisionRadius, 1.1)
	assert.Less(t, m.CollisionRadius, 1.272)
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
	t.Run("success", func(t *testing.T) {
		marker := MarkerFixtures.Get("1000003-4")
		e := marker.Embeddings()

		r := NewFace("123", SrcAuto, e)
		assert.Equal(t, "", r.FaceSrc)
		assert.Equal(t, "123", r.SubjectUID)
	})
}

func TestFace_SetEmbeddings(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		marker := MarkerFixtures.Get("1000003-4")
		e := marker.Embeddings()
		f := FaceFixtures.Get("joe-biden")
		assert.NotEqual(t, e[0][0], f.Embedding()[0])

		err := f.SetEmbeddings(e)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, e[0][0], f.Embedding()[0])
	})
}

func TestFace_Embedding(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		f := FaceFixtures.Get("joe-biden")

		assert.Equal(t, 0.10730543085474682, f.Embedding()[0])
	})
	t.Run("empty embedding", func(t *testing.T) {
		f := NewFace("12345", SrcAuto, Embeddings{})

		assert.Empty(t, f.Embedding())
	})
	t.Run("invalid embedding json", func(t *testing.T) {
		f := NewFace("12345", SrcAuto, Embeddings{})
		f.EmbeddingJSON = []byte("[false]")

		assert.Equal(t, float64(0), f.Embedding()[0])
	})
}

func TestFace_UpdateMatchTime(t *testing.T) {
	f := NewFace("12345", SrcAuto, Embeddings{})
	initialMatchTime := f.MatchedAt
	assert.Equal(t, initialMatchTime, f.MatchedAt)
	f.UpdateMatchTime()
	assert.NotEqual(t, initialMatchTime, f.MatchedAt)
}
