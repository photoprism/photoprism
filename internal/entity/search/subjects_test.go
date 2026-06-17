package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"

	"github.com/stretchr/testify/assert"
)

func TestSubjects(t *testing.T) {
	t.Run("FindAll", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson})
		assert.NoError(t, err)
		// t.Logf("Subjects: %#v", results)
		assert.LessOrEqual(t, 3, len(results))
	})
	t.Run("FindTwoSubjectsSortByCount", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Count: 2, Order: "count"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.Len(t, results, 2)
		if len(results) > 1 {
			assert.GreaterOrEqual(t, results[0].FileCount, results[1].FileCount)
		}
	})
	t.Run("FindAllSortByName", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Order: "name"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.LessOrEqual(t, 3, len(results))
		if len(results) > 0 {
			assert.Equal(t, "Actor A", results[0].SubjName)
		}
	})
	t.Run("SortByAdded", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Order: "added"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.LessOrEqual(t, 3, len(results))
		if len(results) > 0 {
			assert.Equal(t, "Jane Doe", results[0].SubjName)
		}
	})
	t.Run("SortByRelevance", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Order: "relevance"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.LessOrEqual(t, 3, len(results))
		if len(results) > 0 {
			assert.Equal(t, "John Doe", results[0].SubjName)
		}
	})
	t.Run("SearchFavorite", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Favorite: "yes"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.LessOrEqual(t, 1, len(results))
		if len(results) > 0 {
			assert.Equal(t, "John Doe", results[0].SubjName)
		}
	})
	t.Run("SearchPrivate", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Private: "true"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.Len(t, results, 0)
	})
	t.Run("SearchExcluded", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Excluded: "ja"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.Len(t, results, 0)
	})
	t.Run("SearchFileCountGreaterThanTwo", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Files: 2, Excluded: "no"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.LessOrEqual(t, 1, len(results))
	})
	t.Run("SearchForAlias", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Query: "Powell", Favorite: "no", Private: "no"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.LessOrEqual(t, 1, len(results))
		if len(results) > 0 {
			assert.Equal(t, "Dangling Subject", results[0].SubjName)
		}
	})
	t.Run("SearchForId", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, UID: "js6sg6b2h8njw0sx"})
		if assert.NoError(t, err) {
			if assert.Len(t, results, 1) {
				assert.Equal(t, "Joe Biden", results[0].SubjName)
			}
		}
		//t.Logf("Subjects: %#v", results)
	})
	t.Run("NotNil", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, UID: "jszzzzzzzzzzzzzz"})
		if assert.NoError(t, err) {
			assert.Len(t, results, 0)
			assert.NotNil(t, results)
		}
	})
}

func TestSubjectUIDs(t *testing.T) {
	t.Run("SearchForAlias", func(t *testing.T) {
		results, _, _ := SubjectUIDs("Powell")
		//t.Logf("Subjects: %#v", results)
		//t.Logf("Names: %#v", names)
		assert.Len(t, results, 1)
	})
	t.Run("SearchForNotExistingName", func(t *testing.T) {
		results, _, _ := SubjectUIDs("Anonymous")
		//t.Logf("Subjects: %#v", results)
		//t.Logf("Names: %#v", names)
		assert.Len(t, results, 0)
	})
	t.Run("SearchWithEmptyString", func(t *testing.T) {
		results, _, _ := SubjectUIDs("")
		//t.Logf("Subjects: %#v", results)
		//t.Logf("Names: %#v", names)
		assert.Len(t, results, 0)
	})
}
