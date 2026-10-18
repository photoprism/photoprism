package entity

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Authorization code types.
const (
	AuthCodeRequest = "request"
	AuthCodeCode    = "code"
)

// AuthCode represents a pending OAuth authorization request or an authorization code issued for it.
// Only the hash of the request ID or code is stored.
type AuthCode struct {
	ID                  uint      `gorm:"primaryKey;size:32;" json:"-" yaml:"-"`
	CodeHash            string    `gorm:"type:bytes;size:64;uniqueIndex:uix_auth_codes_code_hash;default:'';" json:"-" yaml:"-"`
	CodeType            string    `gorm:"type:bytes;size:16;default:'';" json:"CodeType" yaml:"CodeType"`
	ClientID            string    `gorm:"type:bytes;size:512;default:'';" json:"ClientID" yaml:"ClientID"`
	ClientName          string    `gorm:"size:200;default:'';" json:"ClientName" yaml:"ClientName,omitempty"`
	UserUID             string    `gorm:"type:bytes;size:42;index;default:'';" json:"UserUID" yaml:"UserUID,omitempty"`
	RedirectURI         string    `gorm:"type:bytes;size:2048;default:'';" json:"RedirectURI" yaml:"RedirectURI"`
	Scope               string    `gorm:"size:1024;default:'';" json:"Scope" yaml:"Scope"`
	Resource            string    `gorm:"type:bytes;size:512;default:'';" json:"Resource" yaml:"Resource"`
	State               string    `gorm:"size:1024;default:'';" json:"-" yaml:"-"`
	CodeChallenge       string    `gorm:"type:bytes;size:128;default:'';" json:"-" yaml:"-"`
	CodeChallengeMethod string    `gorm:"type:bytes;size:16;default:'';" json:"-" yaml:"-"`
	ExpiresAt           time.Time `gorm:"index;" json:"ExpiresAt" yaml:"ExpiresAt"`
	CreatedAt           time.Time `json:"CreatedAt" yaml:"-"`
}

// TableName returns the entity table name.
func (AuthCode) TableName() string {
	return "auth_codes"
}

// Create inserts a new row into the database.
func (m *AuthCode) Create() error {
	if m == nil {
		return errors.New("auth code must not be nil")
	} else if m.CodeHash == "" {
		return errors.New("auth code hash must not be empty")
	}

	return Db().Create(m).Error
}

// Expired reports whether the request or code has expired.
func (m *AuthCode) Expired() bool {
	if m == nil {
		return true
	}

	return !time.Now().UTC().Before(m.ExpiresAt)
}

// Consume deletes the row and reports whether this call removed it, so that only one of
// several concurrent redemptions succeeds.
func (m *AuthCode) Consume() (bool, error) {
	if m == nil || m.ID == 0 {
		return false, errors.New("auth code id must not be empty")
	}

	result := UnscopedDb().Where("id = ? AND code_hash = ?", m.ID, m.CodeHash).Delete(&AuthCode{})

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}

// Replace atomically replaces the hash and type of an unexpired row and sets the other fields that are passed,
// and reports whether this call changed it, so that a pending request can be completed only once.
func (m *AuthCode) Replace(codeHash, codeType string, values Values) (bool, error) {
	if m == nil || m.ID == 0 {
		return false, errors.New("auth code id must not be empty")
	} else if codeHash == "" || codeType == "" {
		return false, errors.New("auth code hash and type must not be empty")
	}

	update := Values{}

	for k, v := range values {
		update[k] = v
	}

	update["code_hash"] = codeHash
	update["code_type"] = codeType

	result := UnscopedDb().Model(&AuthCode{}).
		Where("id = ? AND code_hash = ? AND code_type = ? AND expires_at > ?", m.ID, m.CodeHash, m.CodeType, time.Now().UTC()).
		UpdateColumns(update)

	if result.Error != nil {
		return false, result.Error
	} else if result.RowsAffected != 1 {
		return false, nil
	}

	m.CodeHash = codeHash
	m.CodeType = codeType

	return true, nil
}

// FindAuthCode returns the row with the specified hash and type, or nil if none was found.
func FindAuthCode(codeHash, codeType string) (*AuthCode, error) {
	if codeHash == "" || codeType == "" {
		return nil, nil
	}

	m := &AuthCode{}

	if err := UnscopedDb().Where("code_hash = ? AND code_type = ?", codeHash, codeType).First(m).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return m, nil
}

// DeleteExpiredAuthCodes deletes expired authorization requests and codes and returns the number of deleted rows.
func DeleteExpiredAuthCodes() (int64, error) {
	result := UnscopedDb().Where("expires_at < ?", time.Now().UTC()).Delete(&AuthCode{})
	return result.RowsAffected, result.Error
}
