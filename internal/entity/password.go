package entity

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// DefaultPasswordCost specifies the cost of the BCrypt Password Hash,
// see https://github.com/photoprism/photoprism/issues/3718.
var DefaultPasswordCost = 12

// Password represents a password hash.
type Password struct {
	UID       string    `gorm:"type:VARBINARY(255);primary_key;" json:"UID"`
	Hash      string    `deepcopier:"skip" gorm:"type:VARBINARY(255);" json:"Hash"`
	CreatedAt time.Time `deepcopier:"skip" json:"CreatedAt"`
	UpdatedAt time.Time `deepcopier:"skip" json:"UpdatedAt"`
}

// TableName returns the entity table name.
func (Password) TableName() string {
	return "passwords"
}

// NewPassword creates a new password instance.
func NewPassword(uid, pw string, allowHash bool) Password {
	if uid == "" {
		panic("auth: cannot set password without uid")
	}

	m := Password{UID: uid}

	if pw != "" {
		if err := m.SetPassword(pw, allowHash); err != nil {
			log.Errorf("auth: %s", err)
		}
	}

	return m
}

// SetPassword sets a new password stored as hash.
func (m *Password) SetPassword(pw string, allowHash bool) error {
	// Remove leading and trailing white space.
	pw = clean.Password(pw)

	// Check if password is too short or too long.
	if len([]rune(pw)) < 1 {
		return authn.ErrPasswordTooShort
	} else if len(pw) > txt.ClipPassword {
		return authn.ErrPasswordTooLong
	}

	// Check if string already is a bcrypt hash.
	if allowHash {
		if cost, err := bcrypt.Cost([]byte(pw)); err == nil && cost >= bcrypt.MinCost {
			m.Hash = pw
			return nil
		}
	}

	// Generate hash from plain text string using the default password cost.
	if bytes, err := bcrypt.GenerateFromPassword([]byte(pw), DefaultPasswordCost); err != nil {
		return err
	} else {
		m.Hash = string(bytes)
		return nil
	}
}

// Valid checks if the password is correct.
func (m *Password) Valid(s string) bool {
	return !m.Invalid(s)
}

// Invalid checks if the specified password is incorrect.
func (m *Password) Invalid(s string) bool {
	if m.Empty() {
		// Invalid, no password set.
		return true
	} else if s = clean.Password(s); s == "" {
		// Invalid, no password provided.
		return true
	} else if err := bcrypt.CompareHashAndPassword([]byte(m.Hash), []byte(s)); err != nil {
		// Invalid, does not match.
		return true
	}

	// Not invalid.
	return false
}

// Create inserts a new row to the database.
func (m *Password) Create() error {
	return Db().Create(m).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *Password) Save() error {
	return Db().Save(m).Error
}

// Delete removes the password record from the database.
func (m *Password) Delete() error {
	if m.UID == "" {
		return fmt.Errorf("missing password uid")
	}

	return Db().Delete(m).Error
}

// FindPassword returns an entity pointer if exists.
func FindPassword(uid string) *Password {
	result := Password{}

	if err := Db().Where("uid = ?", uid).First(&result).Error; err == nil {
		return &result
	}

	return nil
}

// Cost returns the hashing cost of the currently set password.
func (m *Password) Cost() (int, error) {
	if m.Empty() {
		return 0, authn.ErrPasswordRequired
	}

	return bcrypt.Cost([]byte(m.Hash))
}

// Empty checks if a password has not been set yet.
func (m *Password) Empty() bool {
	return m.Hash == ""
}

// String returns the BCrypt Password Hash.
func (m *Password) String() string {
	return m.Hash
}
