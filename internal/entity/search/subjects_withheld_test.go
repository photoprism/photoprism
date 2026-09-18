package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// TestUserSubjects_OmitsHiddenPeople is the named regression for the hidden flag on the subject
// search. frm.All only decides whether the filter block runs and frm.Hidden decides inside it, so
// the session check sets both.
func TestUserSubjects_OmitsHiddenPeople(t *testing.T) {
	hidden := newWithheldSearchSubject(t, "Hidden Hilda", true)
	visible := entity.SubjectFixtures.Pointer("john-doe")

	holds := func(results SubjectResults, uid string) bool {
		for i := range results {
			if results[i].SubjUID == uid {
				return true
			}
		}

		return false
	}

	frm := func() form.SearchSubjects {
		return form.SearchSubjects{Hidden: "yes", Count: 1000}
	}

	t.Run("Unscoped", func(t *testing.T) {
		results, err := Subjects(frm())
		require.NoError(t, err)
		assert.True(t, holds(results, hidden.SubjUID), "internal and CLI use is not scoped")
	})
	t.Run("Admin", func(t *testing.T) {
		results, err := UserSubjects(frm(), entity.SessionFixtures.Pointer("alice"))
		require.NoError(t, err)
		assert.True(t, holds(results, hidden.SubjUID))
	})
	t.Run("WithoutPrivatePeople", func(t *testing.T) {
		results, err := UserSubjects(frm(), filteredSession())
		require.NoError(t, err)
		assert.True(t, holds(results, visible.SubjUID), "everyone else is still listed")
		assert.False(t, holds(results, hidden.SubjUID))
	})
	t.Run("WithoutPrivatePeopleAll", func(t *testing.T) {
		results, err := UserSubjects(form.SearchSubjects{All: true, Hidden: "yes", Count: 1000}, filteredSession())
		require.NoError(t, err)
		assert.False(t, holds(results, hidden.SubjUID), "all does not skip the visibility filters")
	})
}
