package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSession(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result, err := Session("")
		t.Logf("session: %#v", result)
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "", result.ID)
		assert.Equal(t, "", result.UserUID)
		assert.Equal(t, "", result.UserName)
	})
	t.Run("Alice", func(t *testing.T) {
		if result, err := Session("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("session: %#v", result)
			assert.NotNil(t, result)
			assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", result.ID)
			assert.Equal(t, "uqxetse3cy5eo9z2", result.UserUID)
			assert.Equal(t, "alice", result.UserName)
		}
	})
	t.Run("Bob", func(t *testing.T) {
		if result, err := Session("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("session: %#v", result)
			assert.NotNil(t, result)
			assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1", result.ID)
			assert.Equal(t, "uqxc08w3d0ej2283", result.UserUID)
			assert.Equal(t, "bob", result.UserName)
		}
	})
}

func TestSessions(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		if results, err := Sessions(0, 0, "", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
			//t.Logf("sessions: %#v", results)
		}
	})
	t.Run("Limit", func(t *testing.T) {
		if results, err := Sessions(1, 0, "", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 1, len(results))
			//t.Logf("sessions: %#v", results)
		}
	})
	t.Run("Offset", func(t *testing.T) {
		if results, err := Sessions(0, 1, "", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
			//t.Logf("sessions: %#v", results)
		}
	})
	t.Run("SearchAlice", func(t *testing.T) {
		if results, err := Sessions(100, 0, "sess_expires DESC, user_name", "alice"); err != nil {
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
	t.Run("SortByID", func(t *testing.T) {
		if results, err := Sessions(100, 0, "id", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
			//t.Logf("sessions: %#v", results)
		}
	})
	t.Run("SearchAliceSortByID", func(t *testing.T) {
		if results, err := Sessions(100, 0, "id", "alice"); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 1, len(results))
			//t.Logf("sessions: %#v", results)
		}
	})
}
