package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/stretchr/testify/assert"
)

func TestFaces(t *testing.T) {
	results, err := Faces(true)

	if err != nil {
		t.Fatal(err)
	}

	assert.GreaterOrEqual(t, len(results), 1)

	for _, val := range results {
		assert.IsType(t, entity.Face{}, val)
	}
}

func TestMatchKnownFaces(t *testing.T) {
	const faceFixtureId = uint(6)

	if m, err := MarkerByID(faceFixtureId); err != nil {
		t.Fatal(err)
	} else {
		assert.Empty(t, m.RefUID)
	}

	affected, err := MatchKnownFaces()

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(1), affected)

	if m, err := MarkerByID(faceFixtureId); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(t, "rqu0xs11qekk9jx8", m.RefUID)
	}
}
