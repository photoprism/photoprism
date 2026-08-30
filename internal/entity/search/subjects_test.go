package search

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"

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

// TestUserSubjects covers the session scoping, which decides what a role without private access is
// allowed to reach. Exercised here rather than through the API because CE grants people to admin and
// client only, both of which hold private access - the roles this guard exists for are edition ones.
func TestUserSubjects(t *testing.T) {
	m := entity.NewSubject("Private Search Subject", entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, m)
	m.SubjPrivate = true
	require.NoError(t, m.Create())

	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid = ?", m.SubjUID) })

	holds := func(t *testing.T, results SubjectResults) bool {
		t.Helper()
		for _, r := range results {
			if r.SubjUID == m.SubjUID {
				return true
			}
		}
		return false
	}

	// Every shape that reaches a row: the default list, the two filters that name private people,
	// and the uid lookup, which answers before the content filters.
	forms := map[string]form.SearchSubjects{
		"Default": {Type: entity.SubjPerson, Count: 1000},
		"All":     {Type: entity.SubjPerson, Count: 1000, All: true},
		"Private": {Type: entity.SubjPerson, Count: 1000, Private: "yes"},
		"UID":     {UID: m.SubjUID, Count: 1000},
	}

	t.Run("DeniedRole", func(t *testing.T) {
		sess := entity.SessionFixtures.Pointer("visitor")
		require.True(t, acl.Rules.Deny(acl.ResourcePeople, sess.GetUser().AclRole(), acl.AccessPrivate))

		for name, frm := range forms {
			t.Run(name, func(t *testing.T) {
				results, err := UserSubjects(frm, sess)
				require.NoError(t, err)
				assert.False(t, holds(t, results), "a private person must not be reachable")
			})
		}
	})
	t.Run("AllowedRole", func(t *testing.T) {
		sess := entity.SessionFixtures.Pointer("alice")
		require.False(t, acl.Rules.Deny(acl.ResourcePeople, sess.GetUser().AclRole(), acl.AccessPrivate))

		for name, frm := range forms {
			t.Run(name, func(t *testing.T) {
				results, err := UserSubjects(frm, sess)
				require.NoError(t, err)
				// The default list excludes nothing here: private is only filtered when asked for.
				assert.True(t, holds(t, results), "and must stay reachable for a role that may see it")
			})
		}
	})
	t.Run("NoSession", func(t *testing.T) {
		// The unscoped entry point is what internal callers use; it must not start filtering.
		results, err := Subjects(form.SearchSubjects{UID: m.SubjUID, Count: 1000})
		require.NoError(t, err)
		assert.True(t, holds(t, results))
	})
}

// TestSubjectSessionSeesPrivate pins the role the scoping resolves. A client session carries no
// user, so reading the user role resolves to RoleNone and refuses what the client was authorized
// with - the handler admits the request on the client role and the scoping has to agree with it.
func TestSubjectSessionSeesPrivate(t *testing.T) {
	t.Run("NoSession", func(t *testing.T) {
		assert.True(t, SubjectSessionSeesPrivate(nil), "internal and CLI use is not scoped")
	})
	t.Run("Admin", func(t *testing.T) {
		assert.True(t, SubjectSessionSeesPrivate(entity.SessionFixtures.Pointer("alice")))
	})
	t.Run("DeniedRole", func(t *testing.T) {
		assert.False(t, SubjectSessionSeesPrivate(entity.SessionFixtures.Pointer("visitor")))
	})
	t.Run("ClientWithoutUser", func(t *testing.T) {
		sess := &entity.Session{ClientName: "subject-scope-probe", AuthProvider: authn.ProviderClient.String(),
			AuthMethod: authn.MethodOAuth2.String(), AuthScope: "people"}

		require.True(t, sess.IsClient())
		require.True(t, sess.NoUser())
		assert.True(t, SubjectSessionSeesPrivate(sess), "the client role holds full access to people")
	})
}

// TestSubjects_OmitsDeleted pins that the uid lookup drops a soft-deleted person like the list does.
// The uid is the one branch that answers before the other filters, so it had to be told separately.
func TestSubjects_OmitsDeleted(t *testing.T) {
	m := entity.NewSubject("Deleted Search Subject", entity.SubjPerson, entity.SrcManual)
	require.NotNil(t, m)
	require.NoError(t, m.Create())

	t.Cleanup(func() { entity.UnscopedDb().Delete(&entity.Subject{}, "subj_uid = ?", m.SubjUID) })

	results, err := Subjects(form.SearchSubjects{UID: m.SubjUID, Count: 1000})
	require.NoError(t, err)
	require.Len(t, results, 1)

	require.NoError(t, m.Delete())

	results, err = Subjects(form.SearchSubjects{UID: m.SubjUID, Count: 1000})
	require.NoError(t, err)
	assert.Empty(t, results)
}
