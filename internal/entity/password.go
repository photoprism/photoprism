package entity

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Password represents a password hash.
type Password struct {
	UID       string    `gorm:"type:VARBINARY(255);primary_key;" json:"UID"`
	Hash      string    `deepcopier:"skip" gorm:"type:VARBINARY(255);" json:"Hash"`
	CreatedAt time.Time `deepcopier:"skip" json:"CreatedAt"`
	UpdatedAt time.Time `deepcopier:"skip" json:"UpdatedAt"`
}

// NewPassword creates a new password instance.
func NewPassword(uid, password string) Password {
	if uid == "" {
		panic("auth: cannot set password without uid")
	}

	m := Password{UID: uid}

	if password != "" {
		if err := m.SetPassword(password); err != nil {
			log.Errorf("auth: failed setting password for %s", uid)
		}
	}

	return m
}

// SetPassword sets a new password stored as hash.
func (m *Password) SetPassword(password string) error {
	if bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14); err != nil {
		return err
	} else {
		m.Hash = string(bytes)
		return nil
	}
}

// InvalidPassword returns true if the given password does not match the hash.
func (m *Password) InvalidPassword(password string) bool {
	if m.Hash == "" && password == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(m.Hash), []byte(password))
	return err != nil
}

// Create inserts a new row to the database.
func (m *Password) Create() error {
	return Db().Create(m).Error
}

// Save inserts a new row to the database or updates a row if the primary key already exists.
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

// String returns the password hash.
func (m *Password) String() string {
	return m.Hash
}

// Unknown returns true if the password is an empty string.
func (m *Password) Unknown() bool {
	return m.Hash == ""
}
