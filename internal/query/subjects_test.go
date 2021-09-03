package query

import (
	"strings"
	"testing"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/stretchr/testify/assert"
)

func TestPeople(t *testing.T) {
	if results, err := People(); err != nil {
		t.Fatal(err)
	} else {
		assert.LessOrEqual(t, 3, len(results))
		t.Logf("people: %#v", results)
	}
}

func TestPeopleCount(t *testing.T) {
	if result, err := PeopleCount(); err != nil {
		t.Fatal(err)
	} else {
		assert.LessOrEqual(t, 3, result)
		t.Logf("there are %d people", result)
	}
}

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
	assert.LessOrEqual(t, int64(0), affected)
}

func TestSearchSubjectUIDs(t *testing.T) {
	t.Run("john & his | cats", func(t *testing.T) {
		result, names, remaining := SearchSubjectUIDs("john & his | cats")

		if len(result) != 1 {
			t.Fatal("expected one result")
		} else {
			assert.Equal(t, "jqu0xs11qekk9jx8", result[0])
			assert.Equal(t, "his | cats", remaining)
			assert.Equal(t, "John Doe", strings.Join(names, ", "))
		}
	})
	t.Run("xxx", func(t *testing.T) {
		result, _, _ := SearchSubjectUIDs("xxx")
		assert.Empty(t, result)
	})
	t.Run("empty string", func(t *testing.T) {
		result, _, _ := SearchSubjectUIDs("")
		assert.Empty(t, result)
	})
}
