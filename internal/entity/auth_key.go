package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/pquerna/otp"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// AuthKey represents a two-factor authentication key.
type AuthKey struct {
	UID           string    `gorm:"type:VARBINARY(255);primary_key;" json:"UID"`
	KeyType       string    `gorm:"size:64;default:'';primary_key;" json:"KeyType" yaml:"KeyType"`
	KeyURL        string    `gorm:"size:2048;default:'';column:key_url;" json:"-" yaml:"-"`
	key           *otp.Key  `gorm:"-" yaml:"-"`
	RecoveryCodes string    `gorm:"size:2048;default:'';" json:"-" yaml:"-"`
	RecoveryEmail string    `gorm:"size:255;" json:"RecoveryEmail" yaml:"RecoveryEmail,omitempty"`
	CreatedAt     time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt     time.Time `json:"UpdatedAt" yaml:"-"`
}

// TableName returns the entity table name.
func (AuthKey) TableName() string {
	return "auth_keys"
}

// NewAuthKey returns a new two-factor authentication key or nil if no valid entity UID was provided.
func NewAuthKey(uid string, keyUrl string) (*AuthKey, error) {
	// Create new authentication key.
	m := &AuthKey{
		UID:           uid,
		KeyURL:        keyUrl,
		RecoveryCodes: "",
		RecoveryEmail: "",
	}

	// Return an error if the uid or key are invalid.
	if rnd.InvalidUID(uid, 0) {
		return m, errors.New("auth: invalid uid")
	} else if keyUrl == "" {
		return m, errors.New("auth: invalid url")
	} else if err := m.SetKeyURL(keyUrl); err != nil {
		return m, err
	}

	return m, nil
}

// SetUID assigns a valid entity UID.
func (m *AuthKey) SetUID(uid string) *AuthKey {
	if rnd.IsUID(uid, 0) {
		m.UID = uid
	}

	return m
}

// InvalidUID checks if the entity UID is invalid.
func (m *AuthKey) InvalidUID() bool {
	if m == nil {
		return true
	}

	return !rnd.IsUID(m.UID, 0)
}

// Key returns the parsed two-factor authentication key or nil if the KeyURL is invalid.
func (m *AuthKey) Key() *otp.Key {
	if m == nil {
		return nil
	} else if m.key != nil {
		return m.key
	}

	key, err := otp.NewKeyFromURL(m.KeyURL)

	if err != nil {
		return nil
	}

	m.key = key

	return m.key
}

// SetKey sets a new two-factor authentication key.
func (m *AuthKey) SetKey(key *otp.Key) error {
	if key == nil {
		return errors.New("auth: key is nil")
	}

	if keyType := key.Type(); authn.MethodTOTP.NotEqual(keyType) {
		return fmt.Errorf("auth: invalid key type %s", clean.Log(keyType))
	} else if key.Secret() == "" {
		return errors.New("auth: invalid key secret")
	}

	m.KeyType = key.Type()
	m.KeyURL = key.URL()
	m.key = key

	return nil
}

// SetKeyURL sets a new two-factor authentication key based on the URL provided.
func (m *AuthKey) SetKeyURL(keyUrl string) error {
	key, err := otp.NewKeyFromURL(keyUrl)

	if err != nil {
		return fmt.Errorf("auth: %s", err)
	} else if key == nil {
		return errors.New("auth: failed to parse url")
	}

	return m.SetKey(key)
}
