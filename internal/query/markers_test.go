package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestMarkers(t *testing.T) {
	results, err := Markers(3, 0, entity.MarkerFace, false, false)

	if err != nil {
		t.Fatal(err)
	}

	assert.GreaterOrEqual(t, len(results), 1)

	for _, val := range results {
		assert.IsType(t, entity.Marker{}, val)
	}
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

func TestMatchMarkersWithPeople(t *testing.T) {
	affected, err := MatchMarkersWithPeople()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, affected, 1)
}
