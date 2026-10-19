package entity

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// UserValidationError wraps an error caused by the submitted account values rather than the database.
type UserValidationError struct {
	Err error
}

// Error returns the message of the wrapped error.
func (e UserValidationError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the wrapped error.
func (e UserValidationError) Unwrap() error {
	return e.Err
}

// IsUserValidationError reports whether err was caused by the submitted account values.
func IsUserValidationError(err error) bool {
	var v UserValidationError
	return errors.As(err, &v)
}

// AddUser creates a new user record and sets the password in a single transaction.
func AddUser(frm form.User) error {
	// The auth ID is set after validation, since setting it clears the same ID on other accounts.
	authForm := frm
	frm.AuthID = ""
	user := NewUser().SetFormValues(frm)

	// Check auth id and password.
	if authId := clean.Auth(authForm.AuthID); frm.Provider().Is(authn.ProviderOIDC) && len(authId) < 4 {
		return UserValidationError{Err: authn.ErrAuthIDRequired}
	} else if len(frm.Password) > txt.ClipPassword {
		return UserValidationError{Err: fmt.Errorf("password must have less than %d characters", txt.ClipPassword)}
	} else if (frm.Provider().RequiresLocalPassword() || frm.Password != "") && len([]rune(frm.Password)) < PasswordLength {
		return UserValidationError{Err: fmt.Errorf("password must have at least %d characters", PasswordLength)}
	}

	// Check the username as entered, since the form value is sanitized, and then role and email.
	if !authn.IsASCIIUsername(frm.UserName) {
		return UserValidationError{Err: user.usernameASCIIError(frm.UserName)}
	} else if err := user.Validate(); err != nil {
		return err
	}

	user.SetAuthID(authForm.AuthID, authForm.AuthIssuer)

	err := Db().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// Associations are not saved with the user, so its settings and details are saved here.
		if err := tx.Save(user.Settings()).Error; err != nil {
			return err
		} else if err = tx.Save(user.Details()).Error; err != nil {
			return err
		}

		if frm.Password != "" {
			pw := NewPassword(user.UserUID, frm.Password, false)

			if err := tx.Create(&pw).Error; err != nil {
				return err
			}
		}

		log.Infof("successfully added user %s", clean.LogQuote(user.Username()))

		return nil
	})

	if err == nil {
		user.WarnExistingBasePath()
	}

	return err
}
