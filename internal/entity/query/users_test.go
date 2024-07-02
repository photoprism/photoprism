package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisteredUsers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		users := RegisteredUsers()

		for _, user := range users {
			t.Logf("user: %v, %s, %s, %s", user.ID, user.UserUID, user.Username(), user.DisplayName)
			assert.NotEmpty(t, user.UserUID)
		}

		assert.GreaterOrEqual(t, len(users), 3)
	})
}

func TestUsers(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		if results, err := Users(0, 0, "", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
		}
	})
	t.Run("Limit", func(t *testing.T) {
		if results, err := Users(1, 0, "", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 1, len(results))
		}
	})
	t.Run("Offset", func(t *testing.T) {
		if results, err := Users(0, 1, "", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
		}
	})
	t.Run("SearchAlice", func(t *testing.T) {
		if results, err := Users(100, 0, "", "alice"); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 1, len(results))
			if len(results) > 0 {
				assert.Equal(t, 5, results[0].ID)
				assert.Equal(t, "uqxetse3cy5eo9z2", results[0].UserUID)
				assert.Equal(t, "alice", results[0].UserName)
			}
		}
	})
	t.Run("SortByID", func(t *testing.T) {
		if results, err := Users(100, 0, "id", ""); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 2, len(results))
		}
	})
	t.Run("SearchAliceSortByID", func(t *testing.T) {
		if results, err := Users(100, 0, "id", "alice"); err != nil {
			t.Fatal(err)
		} else {
			assert.LessOrEqual(t, 1, len(results))
		}
	})
}
