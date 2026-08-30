package search

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		require.Len(t, results, 2)
		assert.GreaterOrEqual(t, results[0].FileCount, results[1].FileCount)
	})
	t.Run("FindAllSortByName", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Order: "name"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		require.LessOrEqual(t, 3, len(results))
		assert.Equal(t, "Actor A", results[0].SubjName)
	})
	t.Run("SortByAdded", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Order: "added"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		require.LessOrEqual(t, 3, len(results))
		assert.Equal(t, "Jane Doe", results[0].SubjName)
	})
	t.Run("SortByRelevance", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Order: "relevance"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		require.LessOrEqual(t, 3, len(results))
		assert.Equal(t, "John Doe", results[0].SubjName)
	})
	t.Run("SearchFavorite", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, Favorite: "yes"})
		assert.NoError(t, err)
		//t.Logf("Subjects: %#v", results)
		require.LessOrEqual(t, 1, len(results))
		assert.Equal(t, "John Doe", results[0].SubjName)
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
		require.LessOrEqual(t, 1, len(results))
		assert.Equal(t, "Dangling Subject", results[0].SubjName)
	})
	t.Run("SearchForId", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, UID: "js6sg6b2h8njw0sx"})
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, "Joe Biden", results[0].SubjName)
		//t.Logf("Subjects: %#v", results)
	})
	t.Run("NotNil", func(t *testing.T) {
		results, err := Subjects(form.SearchSubjects{Type: entity.SubjPerson, UID: "jszzzzzzzzzzzzzz"})
		require.NoError(t, err)
		assert.Len(t, results, 0)
		assert.NotNil(t, results)
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

// TestSubjects_Birthday covers the field through the search projection, which the People page reads
// and the edit dialog is seeded from: a column the result struct does not map scans as nil in
// silence, and the dialog would then offer to clear a date it never showed.
func TestSubjects_Birthday(t *testing.T) {
	m := entity.NewSubject("Birthday Search Subject", entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, m)

	born := time.Date(1990, 8, 1, 0, 0, 0, 0, time.UTC)
	changed, err := m.SetBirthday(&born)
	require.NoError(t, err)
	require.True(t, changed)
	require.NoError(t, m.Create())

	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid = ?", m.SubjUID) })

	results, err := Subjects(form.SearchSubjects{UID: m.SubjUID})
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.NotNil(t, results[0].SubjBirthday)
	assert.Equal(t, born, results[0].SubjBirthday.UTC())
}

// TestSubjects_BirthdayKeyAlwaysPresent pins the absence of omitempty on the field.
//
// The client tracks only the keys a response carried, and its live list refresh copies a field by
// name: an omitted key is never seen, so clearing a date would stop reaching the loaded row. Nothing
// else fails if a sweep tidies the tag, which is why this asserts on the tag rather than on a value.
func TestSubjects_BirthdayKeyAlwaysPresent(t *testing.T) {
	b, err := json.Marshal(Subject{SubjUID: "js6sg6b1qekk9jx8"})

	require.NoError(t, err)
	assert.Contains(t, string(b), `"Birthday":null`, "an unset date of birth is null, never absent")
}
