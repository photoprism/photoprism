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
	t.Run("Find 2 subjects, sort by count", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Count: 2, Order: "count"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.Greater(t, results[0].FileCount, results[1].FileCount)
		assert.Equal(t, 2, len(results))
	})
	t.Run("FindAll sort by name", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Order: "name"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.Equal(t, "Actor A", results[0].SubjName)
		assert.LessOrEqual(t, 3, len(results))
	})
	t.Run("sort by added", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Order: "added"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.Equal(t, "Jane Doe", results[0].SubjName)
		assert.LessOrEqual(t, 3, len(results))
	})
	t.Run("sort by relevance", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Order: "relevance"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.Equal(t, "John Doe", results[0].SubjName)
		assert.LessOrEqual(t, 3, len(results))
	})
	t.Run("search favorite", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Favorite: "yes"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.Equal(t, "John Doe", results[0].SubjName)
		assert.LessOrEqual(t, 1, len(results))
	})
	t.Run("search private", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Private: "true"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.Equal(t, 0, len(results))
	})
	t.Run("search excluded", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Excluded: "ja"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.Equal(t, 0, len(results))
	})
	t.Run("search file count >2", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Files: 2, Excluded: "no"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.LessOrEqual(t, 1, len(results))
	})
	t.Run("search for alias", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Query: "Powell", Favorite: "no", Private: "no"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.Equal(t, "Dangling Subject", results[0].SubjName)
		assert.LessOrEqual(t, 1, len(results))
	})
	t.Run("search for ID", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, UID: "js6sg6b2h8njw0sx"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		assert.Equal(t, "Joe Biden", results[0].SubjName)
		assert.Equal(t, 1, len(results))
	})
}

func TestSubjectUIDs(t *testing.T) {
	t.Run("search for alias", func(t *testing.T) {
		results, _, _ := SubjectUIDs("Powell")
		//t.Logf("Subjects: %#v", results)
		//t.Logf("Names: %#v", names)
		assert.Equal(t, 1, len(results))
	})
	t.Run("search for not existing name", func(t *testing.T) {
		results, _, _ := SubjectUIDs("Anonymous")
		//t.Logf("Subjects: %#v", results)
		//t.Logf("Names: %#v", names)
		assert.Equal(t, 0, len(results))
	})
	t.Run("search with empty string", func(t *testing.T) {
		results, _, _ := SubjectUIDs("")
		//t.Logf("Subjects: %#v", results)
		//t.Logf("Names: %#v", names)
		assert.Equal(t, 0, len(results))
	})
}
