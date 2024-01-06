package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
)

func TestSessions(t *testing.T) {
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
		if results, err := Sessions(form.SearchSessions{Offset: 1}); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
			//t.Logf("sessions: %#v", results)
		}
	})
	t.Run("SearchAlice", func(t *testing.T) {
		if results, err := Sessions(form.SearchSessions{Count: 100, Query: "alice", Order: "sess_expires DESC, user_name"}); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("sessions: %#v", results)
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", results[0].ID)
				assert.Equal(t, "uqxetse3cy5eo9z2", results[0].UserUID)
				assert.Equal(t, "alice", results[0].UserName)
			}
		}
	})
}
