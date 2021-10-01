package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/face"

	"github.com/stretchr/testify/assert"
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
		match, dist := m.Match(face.Embeddings{})

		assert.False(t, match)
		assert.Equal(t, dist, float64(-1))
	})
	t.Run("len(efacEmbeddings) == 0", func(t *testing.T) {
		m := NewFace("12345", SrcAuto, face.Embeddings{})
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

func TestFace_ResolveCollision(t *testing.T) {
	t.Run("collision", func(t *testing.T) {
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
	t.Run("subject id empty", func(t *testing.T) {
		m := NewFace("", SrcAuto, face.Embeddings{})
		if reported, err := m.ResolveCollision(MarkerFixtures.Pointer("1000003-4").Embeddings()); err != nil {
			t.Fatal(err)
		} else {
			assert.False(t, reported)
		}
	})
	t.Run("invalid face id", func(t *testing.T) {
		m := NewFace("123", SrcAuto, face.Embeddings{})
		m.ID = ""
		if reported, err := m.ResolveCollision(MarkerFixtures.Pointer("1000003-4").Embeddings()); err == nil {
			t.Fatal(err)
		} else {
			assert.False(t, reported)
			assert.Equal(t, "invalid face id", err.Error())
		}
	})
	t.Run("embedding empty", func(t *testing.T) {
		m := NewFace("123", SrcAuto, face.Embeddings{})
		m.EmbeddingJSON = []byte("")
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
	t.Run("success", func(t *testing.T) {
		marker := MarkerFixtures.Get("1000003-4")
		e := marker.Embeddings()

		r := NewFace("123", SrcAuto, e)
		assert.Equal(t, "", r.FaceSrc)
		assert.Equal(t, "123", r.SubjUID)
	})
}

func TestFace_SetEmbeddings(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
}

func TestFace_Embedding(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := FaceFixtures.Get("joe-biden")

		assert.Equal(t, 0.10730543085474682, m.Embedding()[0])
	})
	t.Run("empty embedding", func(t *testing.T) {
		m := NewFace("12345", SrcAuto, face.Embeddings{})
		m.EmbeddingJSON = []byte("")

		assert.Empty(t, m.Embedding())
	})
	t.Run("invalid embedding json", func(t *testing.T) {
		m := NewFace("12345", SrcAuto, face.Embeddings{})
		m.EmbeddingJSON = []byte("[false]")

		assert.Equal(t, float64(0), m.Embedding()[0])
	})
}

func TestFace_UpdateMatchTime(t *testing.T) {
	m := NewFace("12345", SrcAuto, face.Embeddings{})
	initialMatchTime := m.MatchedAt
	assert.Equal(t, initialMatchTime, m.MatchedAt)
	m.Matched()
	assert.NotEqual(t, initialMatchTime, m.MatchedAt)
}

func TestFace_Save(t *testing.T) {
	m := NewFace("12345fde", SrcAuto, face.Embeddings{face.Embedding{1}, face.Embedding{2}})
	assert.Nil(t, FindFace(m.ID))
	m.Save()
	assert.NotNil(t, FindFace(m.ID))
	assert.Equal(t, "12345fde", FindFace(m.ID).SubjUID)
}

func TestFace_Update(t *testing.T) {
	m := NewFace("12345fdef", SrcAuto, face.Embeddings{face.Embedding{8}, face.Embedding{16}})
	assert.Nil(t, FindFace(m.ID))
	m.Save()
	assert.NotNil(t, FindFace(m.ID))
	assert.Equal(t, "12345fdef", FindFace(m.ID).SubjUID)

	m2 := FindFace(m.ID)
	m2.Update("SubjUID", "new")
	assert.Equal(t, "new", FindFace(m.ID).SubjUID)
}

func TestFace_RefreshPhotos(t *testing.T) {
	f := FaceFixtures.Get("joe-biden")

	if err := f.RefreshPhotos(); err != nil {
		t.Fatal(err)
	}
}

func TestFirstOrCreateFace(t *testing.T) {
	t.Run("create new face", func(t *testing.T) {
		m := NewFace("12345unique", SrcAuto, face.Embeddings{face.Embedding{99}, face.Embedding{2}})
		r := FirstOrCreateFace(m)
		assert.Equal(t, "12345unique", r.SubjUID)
	})
	t.Run("return existing entity", func(t *testing.T) {
		m := FaceFixtures.Pointer("joe-biden")
		r := FirstOrCreateFace(m)
		assert.Equal(t, "jqy3y652h8njw0sx", r.SubjUID)
		assert.Equal(t, 33, r.Samples)
	})
}

func TestFindFace(t *testing.T) {
	t.Run("existing face", func(t *testing.T) {
		assert.NotNil(t, FindFace("VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG7"))
		assert.Equal(t, 3, FindFace("VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG7").Samples)
	})
	t.Run("empty id", func(t *testing.T) {
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
