package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisteredUsers(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		users := RegisteredUsers()

		for _, user := range users {
			t.Logf("user: %v, %s, %s, %s", user.ID, user.UserUID, user.UserName(), user.DisplayName)
		}

		t.Logf("user count: %v", len(users))

		assert.GreaterOrEqual(t, len(users), 3)
	})
}
