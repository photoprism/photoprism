package entity

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
)

// AddUser creates a new user record and sets the password in a single transaction.
func AddUser(frm form.User) error {
	user := NewUser().SetFormValues(frm)

	if len(frm.Password) < PasswordLength {
		return fmt.Errorf("password must have at least %d characters", PasswordLength)
	}

	if err := user.Validate(); err != nil {
		return err
	}

	return Db().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		pw := NewPassword(user.UserUID, frm.Password, false)

		if err := tx.Create(&pw).Error; err != nil {
			return err
		}

		log.Infof("successfully added user %s", clean.LogQuote(user.Username()))

		return nil
	})
}
