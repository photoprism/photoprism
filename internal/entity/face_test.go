package entity

import (
	"testing"

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
