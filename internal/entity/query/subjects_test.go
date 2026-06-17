package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/stretchr/testify/assert"
)

func TestPeople(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		results, err := People()
		if assert.Nil(t, err) {
			assert.LessOrEqual(t, 3, len(results))
			t.Logf("people: %#v", results)
		}
	})
	t.Run("NotNil", func(t *testing.T) {
		t.Cleanup(func() {
			entity.Entities.Truncate(entity.Db())
			entity.CreateDefaultFixtures()
			entity.CreateTestFixtures()
			entity.File{}.RegenerateIndex()
		})
		// Clean the database as if it's brand new
		entity.Entities.Truncate(entity.Db())
		entity.CreateDefaultFixtures()
		entity.File{}.RegenerateIndex()

		results, err := People()

		if assert.Nil(t, err) {
			assert.NotNil(t, results)
			assert.Len(t, results, 0)
		}
	})
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

func TestRemoveOrphanSubjects(t *testing.T) {
	affected, err := RemoveOrphanSubjects()

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
