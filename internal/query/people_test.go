package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/stretchr/testify/assert"
)

func TestPeople(t *testing.T) {
	results, err := People(3, 0, false)

	if err != nil {
		t.Fatal(err)
	}

	assert.GreaterOrEqual(t, 1, len(results))

	for _, val := range results {
		assert.IsType(t, entity.Person{}, val)
	}
}

func TestFaces(t *testing.T) {
	results, err := Faces()

	if err != nil {
		t.Fatal(err)
	}

	assert.GreaterOrEqual(t, 1, len(results))

	for _, val := range results {
		assert.IsType(t, entity.Face{}, val)
	}
}
