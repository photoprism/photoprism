package query

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestMarkers(t *testing.T) {
	t.Run("find umatched", func(t *testing.T) {
		results, err := Markers(3, 0, entity.MarkerFace, false, false, entity.Timestamp())

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(results), 1)

		for _, val := range results {
			assert.IsType(t, entity.Marker{}, val)
		}
	})
	t.Run("find all", func(t *testing.T) {
		results, err := Markers(3, 0, entity.MarkerFace, false, false, time.Time{})

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(results), 1)

		for _, val := range results {
			assert.IsType(t, entity.Marker{}, val)
		}
	})
	t.Run("find embeddings", func(t *testing.T) {
		results, err := Markers(3, 0, entity.MarkerFace, true, false, time.Time{})

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(results), 1)

		for _, val := range results {
			assert.IsType(t, entity.Marker{}, val)
		}
	})
	t.Run("find false", func(t *testing.T) {
		results, err := Markers(3, 0, entity.MarkerFace, false, true, time.Time{})

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(results))
	})
}

func TestEmbeddings(t *testing.T) {
	results, err := Embeddings(false, false, 0)

	if err != nil {
		t.Fatal(err)
	}

	assert.GreaterOrEqual(t, len(results), 1)

	for _, val := range results {
		assert.IsType(t, entity.Embedding{}, val)
	}
}

func TestRemoveInvalidMarkerReferences(t *testing.T) {
	affected, err := RemoveInvalidMarkerReferences()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, affected, int64(1))
}

func TestMarkersWithInvalidReferences(t *testing.T) {
	f, s, err := MarkersWithInvalidReferences()

	assert.NoError(t, err)

	assert.GreaterOrEqual(t, len(f), 0)
	assert.GreaterOrEqual(t, len(s), 0)
}

func TestCountUnmatchedFaceMarkers(t *testing.T) {
	n, threshold := CountUnmatchedFaceMarkers()

	assert.False(t, threshold.IsZero())
	assert.GreaterOrEqual(t, n, 0)
}

func TestCountMarkers(t *testing.T) {
	n := CountMarkers(entity.MarkerFace)

	assert.GreaterOrEqual(t, n, 1)
}
