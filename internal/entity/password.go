package entity

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

var (
	PasswordCost = 14
)

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
		return fmt.Errorf("password is too short")
	} else if len(pw) > txt.ClipPassword {
		return fmt.Errorf("password must have less than %d characters", txt.ClipPassword)
	}

	// Check if string already is a bcrypt hash.
	if allowHash {
		if cost, err := bcrypt.Cost([]byte(pw)); err == nil && cost >= bcrypt.MinCost {
			m.Hash = pw
			return nil
		}
	}

	// Generate hash from plain text string.
	if bytes, err := bcrypt.GenerateFromPassword([]byte(pw), PasswordCost); err != nil {
		return err
	} else {
		m.Hash = string(bytes)
		return nil
	}
}

// IsValid checks if the password is correct.
func (m *Password) IsValid(s string) bool {
	return !m.IsWrong(s)
}

// IsWrong checks if the specified password is incorrect.
func (m *Password) IsWrong(s string) bool {
	if m.IsEmpty() {
		// No password set.
		return true
	} else if s = clean.Password(s); s == "" {
		// No password provided.
		return true
	} else if err := bcrypt.CompareHashAndPassword([]byte(m.Hash), []byte(s)); err != nil {
		// Wrong password.
		return true
	}

	// Ok.
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
	if m.IsEmpty() {
		return 0, fmt.Errorf("password is empty")
	}

	return bcrypt.Cost([]byte(m.Hash))
}

// IsEmpty returns true if the password is not set.
func (m *Password) IsEmpty() bool {
	return m.Hash == ""
}

// String returns the password hash.
func (m *Password) String() string {
	return m.Hash
}
