package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/stretchr/testify/assert"
)

func TestSubjects(t *testing.T) {
	results, err := Subjects(3, 0)

	if err != nil {
		t.Fatal(err)
	}

	assert.GreaterOrEqual(t, len(results), 1)

	for _, val := range results {
		assert.IsType(t, entity.Subject{}, val)
	}
}

func TestSubjectMap(t *testing.T) {
	results, err := SubjectMap()

	if err != nil {
		t.Fatal(err)
	}

	assert.GreaterOrEqual(t, len(results), 1)

	for _, val := range results {
		assert.IsType(t, entity.Subject{}, val)
	}
}

func TestRemoveDanglingMarkerSubjects(t *testing.T) {
	affected, err := RemoveDanglingMarkerSubjects()

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(1), affected)
}

func TestCreateMarkerSubjects(t *testing.T) {
	affected, err := CreateMarkerSubjects()

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, affected, int64(2))
}

func TestSubjectUIDs(t *testing.T) {
	result, remaining := SubjectUIDs("john & his | cats")

	if len(result) != 1 {
		t.Fatal("expected one result")
	} else {
		assert.Equal(t, "jqu0xs11qekk9jx8", result[0])
		assert.Equal(t, "his | cats", remaining)
	}
}
