package entity

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/pquerna/otp"
)

// Passcode represents a two-factor authentication key.
type Passcode struct {
	UID          string     `gorm:"type:VARBINARY(255);primary_key;" json:"UID"`
	KeyType      string     `gorm:"size:64;default:'';primary_key;" json:"Type" yaml:"Type"`
	KeyURL       string     `gorm:"size:2048;default:'';column:key_url;" json:"-" yaml:"-"`
	key          *otp.Key   `gorm:"-" yaml:"-"`
	RecoveryCode string     `gorm:"size:255;default:'';" json:"-" yaml:"-"`
	VerifiedAt   *time.Time `json:"VerifiedAt" yaml:"-"`
	ActivatedAt  *time.Time `json:"ActivatedAt" yaml:"-"`
	CreatedAt    time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt    time.Time  `json:"UpdatedAt" yaml:"-"`
}

// TableName returns the entity table name.
func (Passcode) TableName() string {
	return "passcodes"
}

// NewPasscode returns a new two-factor authentication key or nil if no valid entity UID was provided.
func NewPasscode(uid string, keyUrl, recoveryCode string) (*Passcode, error) {
	// Create new authentication key.
	m := &Passcode{
		UID:          uid,
		KeyURL:       keyUrl,
		RecoveryCode: clean.Token(recoveryCode),
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
		VerifiedAt:   nil,
		ActivatedAt:  nil,
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

// FindPasscode returns the matching key or nil if it was not found.
func FindPasscode(find Passcode) *Passcode {
	m := &Passcode{}

	keyType := authn.Key(find.KeyType)

	// Build query.
	stmt := UnscopedDb()
	if rnd.IsUID(find.UID, 0) {
		stmt = stmt.Where("uid = ? AND key_type = ?", find.UID, keyType.String())
	} else {
		return nil
	}

	// Find matching record.
	if err := stmt.First(m).Error; err != nil {
		return nil
	}

	return m
}

// Create new entity in the database.
func (m *Passcode) Create() (err error) {
	return Db().Create(m).Error
}

// Save updates the record in the database or inserts a new record if it does not exist yet.
func (m *Passcode) Save() (err error) {
	return UnscopedDb().Save(m).Error
}

// Delete deletes the entity record.
func (m *Passcode) Delete() (err error) {
	if m == nil {
		return fmt.Errorf("entity is nil")
	} else if m.UID == "" {
		return fmt.Errorf("uid not set")
	}

	err = UnscopedDb().Delete(m).Error

	return err
}

// Updates multiple properties in the database.
func (m *Passcode) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// SetUID assigns a valid entity UID.
func (m *Passcode) SetUID(uid string) *Passcode {
	if rnd.IsUID(uid, 0) {
		m.UID = uid
	}

	return m
}

// InvalidUID checks if the entity UID is invalid.
func (m *Passcode) InvalidUID() bool {
	if m == nil {
		return true
	}

	return !rnd.IsUID(m.UID, 0)
}

// Key returns the parsed two-factor authentication key or nil if the KeyURL is invalid.
func (m *Passcode) Key() *otp.Key {
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
func (m *Passcode) SetKey(key *otp.Key) error {
	if key == nil {
		return errors.New("auth: key is nil")
	}

	if keyType := key.Type(); authn.KeyTOTP.NotEqual(keyType) {
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
func (m *Passcode) SetKeyURL(keyUrl string) error {
	key, err := otp.NewKeyFromURL(keyUrl)

	if err != nil {
		return fmt.Errorf("auth: %s", err)
	} else if key == nil {
		return errors.New("auth: failed to parse url")
	}

	return m.SetKey(key)
}

// Secret returns the key secret or an empty string if none is set.
func (m *Passcode) Secret() string {
	if m == nil {
		return ""
	}

	key := m.Key()

	if key == nil {
		return ""
	}

	return key.Secret()
}

// Type returns the normalized key type.
func (m *Passcode) Type() authn.KeyType {
	if m == nil {
		return ""
	}

	return authn.Key(m.KeyType)
}

// GenerateCode returns a valid passcode for testing.
func (m *Passcode) GenerateCode() (code string, err error) {
	if m == nil {
		return "", errors.New("passcode is nil")
	}

	// Get authentication key.
	key := m.Key()

	if key == nil {
		return "", authn.ErrInvalidPasscodeKey
	}

	// Generate code depending on key type.
	switch m.Type() {
	case authn.KeyTOTP:
		code, err = totp.GenerateCodeCustom(
			key.Secret(),
			time.Now().UTC(),
			totp.ValidateOpts{
				Period:    uint(key.Period()),
				Skew:      1,
				Digits:    key.Digits(),
				Algorithm: key.Algorithm(),
			},
		)
	default:
		return "", authn.ErrInvalidPasscodeType
	}

	// Return result.
	return code, err
}

// Valid checks if the passcode provided is valid.
func (m *Passcode) Valid(code string) (valid bool, recovery bool, err error) {
	// Validate arguments.
	if m == nil {
		return false, false, errors.New("passcode is nil")
	} else if code == "" {
		return false, false, authn.ErrPasscodeRequired
	} else if len(code) > 255 {
		return false, false, authn.ErrInvalidPasscodeFormat
	}

	// Get authentication key.
	key := m.Key()

	if key == nil {
		return false, false, authn.ErrInvalidPasscodeKey
	}

	// Check if recovery code has been used.
	if m.RecoveryCode == code {
		return true, true, nil
	}

	// Verify passcode.
	switch m.Type() {
	case authn.KeyTOTP:
		valid, err = totp.ValidateCustom(
			code,
			key.Secret(),
			time.Now().UTC(),
			totp.ValidateOpts{
				Period:    uint(key.Period()),
				Skew:      1,
				Digits:    key.Digits(),
				Algorithm: key.Algorithm(),
			},
		)
	default:
		return false, false, authn.ErrInvalidPasscodeType
	}

	// Check if an error has been returned.
	if err != nil {
		return valid, false, err
	}

	// Set verified timestamp if nil.
	if valid && m.VerifiedAt == nil {
		m.VerifiedAt = TimePointer()
		err = m.Updates(Map{"VerifiedAt": m.VerifiedAt})
	}

	// Return result.
	return valid, false, err
}

// Activate activates the passcode.
func (m *Passcode) Activate() (err error) {
	if m == nil {
		return errors.New("passcode is nil")
	}

	if m.VerifiedAt == nil {
		return authn.ErrPasscodeNotVerified
	} else if m.ActivatedAt != nil {
		return authn.ErrPasscodeAlreadyActivated
	} else {
		m.ActivatedAt = TimePointer()
		err = m.Updates(Map{"ActivatedAt": m.ActivatedAt})
	}

	return err
}

// Image returns an image with a QR Code that can be used to initialize compatible authenticator apps.
func (m *Passcode) Image(size int) (image.Image, error) {
	if m == nil {
		return nil, errors.New("key is nil")
	}

	key := m.Key()

	if key == nil {
		return nil, authn.ErrPasscodeNotSetUp
	}

	return key.Image(size, size)
}

// PNG returns a PNG image buffer with a QR Code that can be used to initialize compatible authenticator apps.
func (m *Passcode) PNG(size int) *bytes.Buffer {
	if m == nil {
		return nil
	}

	img, err := m.Image(size)

	if err != nil {
		return nil
	}

	var buf bytes.Buffer

	err = png.Encode(&buf, img)

	if err != nil {
		return nil
	}

	return &buf
}
