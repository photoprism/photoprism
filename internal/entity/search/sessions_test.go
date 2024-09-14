package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSessions(t *testing.T) {
	expectedUserUid := "uqxetse3cy5eo9z2"
	expectedUserName := "alice"
	expectedSessionId := "fde4d5154d5383370c9f0c21fd51655d54a185a26dc043d1866fc4678e7ecb62"

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
}
