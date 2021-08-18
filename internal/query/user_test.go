package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestDeleteUserByName(t *testing.T) {
	u := entity.FirstOrCreateUser(&entity.User{
		ID:           877,
		AddressID:    1,
		UserName:     "delete",
		FullName:     "Delete",
		PrimaryEmail: "delete@example.com",
	})

	t.Run("delete empty username", func(t *testing.T) {
		err := DeleteUserByName("")
		assert.Error(t, err)
	})
	t.Run("delete fail", func(t *testing.T) {
		err := DeleteUserByName("notmatching")
		assert.Error(t, err)
	})
	t.Run("delete success", func(t *testing.T) {
		err := DeleteUserByName(u.UserName)
		assert.Nil(t, err)
	})
}

func TestAllUsers(t *testing.T) {
	t.Run("list all", func(t *testing.T) {
		users := AllUsers()
		for _, user := range users {
			log.Infof("user: %v, %s, %s, %s", user.ID, user.UserUID, user.UserName, user.FullName)
		}
		log.Infof("user count: %v", len(users))
		assert.Greater(t, len(users), 3)
	})
}
