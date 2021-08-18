package query

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/txt"
)

// DeleteUserByName deletes an existing user or returns error if not found.
func DeleteUserByName(userName string) error {
	if userName == "" {
		return fmt.Errorf("can not delete user from db: empty username")
	}
	user := entity.FindUserByName(userName)
	if err := Db().Where("user_name = ?", userName).Delete(&entity.User{}).Error; user == nil || err != nil {
		return fmt.Errorf("user %s not found", txt.Quote(userName))
	}
	if err := Db().Where("uid = ?", user.UserUID).Delete(&entity.Password{}).Error; err != nil {
		log.Debug(err)
	}
	return nil
}

// AllUsers Returns a list of all registered Users.
func AllUsers() []entity.User {
	var users []entity.User
	if err := Db().Find(&users).Error; err != nil {
		log.Error(err)
		return []entity.User{}
	}
	return users
}
