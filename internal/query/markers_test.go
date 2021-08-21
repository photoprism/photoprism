package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestMarkers(t *testing.T) {
	t.Run("find all", func(t *testing.T) {
		results, err := Markers(3, 0, entity.MarkerFace, false, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(results), 1)

		for _, val := range results {
			assert.IsType(t, entity.Marker{}, val)
		}
	})
	t.Run("find embeddings", func(t *testing.T) {
		results, err := Markers(3, 0, entity.MarkerFace, true, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(results), 1)

		for _, val := range results {
			assert.IsType(t, entity.Marker{}, val)
		}
	})
	t.Run("find false", func(t *testing.T) {
		results, err := Markers(3, 0, entity.MarkerFace, false, true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(results))
	})
}

func TestEmbeddings(t *testing.T) {
	results, err := Embeddings(false)

	if err != nil {
		t.Fatal(err)
	}

	assert.GreaterOrEqual(t, len(results), 1)

	for _, val := range results {
		assert.IsType(t, entity.Embedding{}, val)
	}
}

func TestAddMarkerSubjects(t *testing.T) {
	affected, err := AddMarkerSubjects()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, affected, int64(1))
}

func TestTidyMarkers(t *testing.T) {
	assert.NoError(t, TidyMarkers())
}
