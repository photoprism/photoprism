package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSessions(t *testing.T) {
	expectedUserUid := "uqxetse3cy5eo9z2"
	expectedUserName := "alice"
	// First of alice's sessions ordered by ClientName: the "alice_app_password"
	// app-password fixture (application provider, default method).
	expectedSessionId := "84ced4d6b1b9e5db31d4ad5364d9b7da6b2a0e85372a46df9784852d8ada6e48"

	t.Run("Default", func(t *testing.T) {
		if results, err := Sessions(form.SearchSessions{}); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
			//t.Logf("sessions: %#v", results)
		}
	})
	t.Run("Limit", func(t *testing.T) {
		if results, err := Sessions(form.SearchSessions{Count: 1}); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 1, len(results))
			//t.Logf("sessions: %#v", results)
		}
	})
	t.Run("Offset", func(t *testing.T) {
		if results, err := Sessions(form.SearchSessions{Offset: 1, Order: sortby.LastActive}); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
			//t.Logf("sessions: %#v", results)
		}
	})
	t.Run("Search", func(t *testing.T) {
		if results, err := Sessions(form.SearchSessions{Count: 100, Query: expectedUserName, Order: sortby.SessExpires}); err != nil {
			t.Fatal(err)
		} else {
			// t.Logf("sessions: %#v", results)
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"), results[0].ID)
				assert.Equal(t, expectedUserUid, results[0].UserUID)
				assert.Equal(t, expectedUserName, results[0].UserName)
			}
		}
	})
	t.Run("UID", func(t *testing.T) {
		if results, err := Sessions(form.SearchSessions{Count: 100, UID: expectedUserUid, Order: sortby.SessExpires}); err != nil {
			t.Fatal(err)
		} else {
			// t.Logf("sessions: %#v", results)
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"), results[0].ID)
				assert.Equal(t, expectedUserUid, results[0].UserUID)
				assert.Equal(t, expectedUserName, results[0].UserName)
			}
		}
	})
	t.Run("Providers", func(t *testing.T) {
		if results, err := Sessions(form.SearchSessions{Count: 100, UID: expectedUserUid, Provider: "default,application,client,local,access_token", Order: sortby.ClientName}); err != nil {
			t.Fatal(err)
		} else {
			// t.Logf("sessions: %#v", results)
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, expectedSessionId, results[0].ID)
				assert.Equal(t, expectedUserUid, results[0].UserUID)
				assert.Equal(t, expectedUserName, results[0].UserName)
			}
		}
	})
	t.Run("Methods", func(t *testing.T) {
		if results, err := Sessions(form.SearchSessions{Count: 100, UID: expectedUserUid, Method: "default,oauth2,session,2fa", Order: sortby.ClientName}); err != nil {
			t.Fatal(err)
		} else {
			// t.Logf("sessions: %#v", results)
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, expectedSessionId, results[0].ID)
				assert.Equal(t, expectedUserUid, results[0].UserUID)
				assert.Equal(t, expectedUserName, results[0].UserName)
			}
		}
	})
	t.Run("SortByLoginAt", func(t *testing.T) {
		if results, err := Sessions(form.SearchSessions{Count: 100, Order: sortby.LoginAt}); err != nil {
			t.Fatal(err)
		} else {
			// t.Logf("sessions: %#v", results)
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, "friend", results[0].UserName)
				assert.Equal(t, "bob", results[1].UserName)
			}
		}
	})
	t.Run("SortByLoginAtReverse", func(t *testing.T) {
		if results, err := Sessions(form.SearchSessions{Count: 100, Offset: 2, Order: sortby.LoginAt, Reverse: true}); err != nil {
			t.Fatal(err)
		} else {
			// t.Logf("sessions: %#v", results)
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, "Analytics", results[0].ClientName)
				assert.Equal(t, "Invalid", results[1].ClientName)
			}
		}
	})
	t.Run("SortByCreatedAt", func(t *testing.T) {
		if results, err := Sessions(form.SearchSessions{Count: 100, Order: sortby.CreatedAt}); err != nil {
			t.Fatal(err)
		} else {
			// t.Logf("sessions: %#v", results)
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, "bob", results[0].UserName)
				assert.Equal(t, "friend", results[1].UserName)
			}
		}
	})
	t.Run("InvalidSortOrder", func(t *testing.T) {
		results, err := Sessions(form.SearchSessions{Count: 100, Order: "invalid"})
		assert.Error(t, err)
		assert.Empty(t, results)
	})
}

func TestSessions_Literal(t *testing.T) {
	base := "zzl" + rnd.Base36(5)

	for _, name := range []string{base + "_a", base + "Xa"} {
		s := entity.NewSession(3600, 0)
		s.UserName = name
		require.NoError(t, s.Create())
		t.Cleanup(func() { _ = entity.UnscopedDb().Delete(s).Error })
	}

	result, err := Sessions(form.SearchSessions{Query: base + "_a", Count: 100})
	require.NoError(t, err)

	if assert.Len(t, result, 1) {
		assert.Equal(t, base+"_a", result[0].UserName)
	}
}

func TestSessions_Bindable(t *testing.T) {
	base := "zzb" + rnd.Base36(5)

	s := entity.NewSession(3600, 0)
	s.UserName = base + "name"
	require.NoError(t, s.Create())
	t.Cleanup(func() { _ = entity.UnscopedDb().Delete(s).Error })

	result, err := Sessions(form.SearchSessions{Query: base + "\x00na\xffme", Count: 100})
	require.NoError(t, err)

	if assert.Len(t, result, 1) {
		assert.Equal(t, base+"name", result[0].UserName)
	}

	uid := entity.UserFixtures.Get("alice").UserUID
	expected, err := Sessions(form.SearchSessions{UID: uid, Count: 100})
	require.NoError(t, err)
	require.NotEmpty(t, expected)

	result, err = Sessions(form.SearchSessions{UID: uid[:4] + "\x00" + uid[4:] + "\xff", Count: 100})
	require.NoError(t, err)
	assert.Len(t, result, len(expected))
}

func TestSessions_NoMatch(t *testing.T) {
	// A filter value of only NUL bytes and invalid UTF-8 matches nothing, while an empty one stays unfiltered.
	all, err := Sessions(form.SearchSessions{Count: 100})
	require.NoError(t, err)
	require.NotEmpty(t, all)

	for _, frm := range []form.SearchSessions{{Query: "\x00", Count: 100}, {UID: "\xff", Count: 100},
		{Provider: "\xff", Count: 100}, {Method: "\x00", Count: 100}, {Query: "\x00 ", Count: 100}, {Query: "\xff|", Count: 100}} {
		result, err := Sessions(frm)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result, "%+v", frm)
	}
}
