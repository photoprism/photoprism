package entity

import (
	"testing"

	"github.com/pquerna/otp"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewAuthKey(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		uid := "us7gqkzx1g9a82h4"
		keyUrl := "otpauth://totp/Example:alice@google.com?secret=JBSWY3DPEHPK3PXP&issuer=Example&algorithm=sha256&digits=8"
		recoveryCode := rnd.RecoveryCode()

		m, err := NewPasscode(uid, keyUrl, recoveryCode)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("NewPasscode/Valid: %#v", m)

		assert.NotNil(t, m)
		assert.Equal(t, uid, m.UID)
		assert.Equal(t, authn.KeyTOTP.String(), m.KeyType)
		assert.True(t, authn.KeyTOTP.Equal(m.KeyType))
		assert.Equal(t, recoveryCode, m.RecoveryCode)
	})
	t.Run("Invalid", func(t *testing.T) {
		m, err := NewPasscode("foo", "", "")

		t.Logf("TestNewAuthKey/Invalid: %#v", m)

		assert.Error(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, "foo", m.UID)
		assert.True(t, authn.KeyTOTP.NotEqual(m.KeyType))
		assert.Equal(t, "", m.RecoveryCode)
	})
}

func TestAuthKey_Key(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		uid := "us7gqkzx1g9a82h4"
		keyUrl := "otpauth://totp/Example:alice@google.com?secret=JBSWY3DPEHPK3PXP&issuer=Example&algorithm=sha256&digits=8"
		recoveryCode := rnd.RecoveryCode()

		m, err := NewPasscode(uid, keyUrl, recoveryCode)

		if err != nil {
			t.Fatal(err)
		}

		key := m.Key()

		t.Logf("TestAuthKey_Key/Valid: %#v", m)

		assert.NotNil(t, m)
		assert.Equal(t, authn.KeyTOTP.String(), key.Type())
		assert.Equal(t, keyUrl, key.URL())
		assert.Equal(t, uint64(30), key.Period())
		assert.Equal(t, "JBSWY3DPEHPK3PXP", key.Secret())
		assert.Equal(t, keyUrl, key.String())
		assert.Equal(t, "alice@google.com", key.AccountName())
		assert.Equal(t, "SHA256", key.Algorithm().String())
		assert.Equal(t, otp.Digits(8), key.Digits())
		assert.Equal(t, 8, key.Digits().Length())
		assert.Equal(t, recoveryCode, m.RecoveryCode)

	})
	t.Run("Invalid", func(t *testing.T) {
		m, err := NewPasscode("foo", "", "")

		if err == nil {
			t.Fatal("error expected")
		}

		key := m.Key()

		t.Logf("TestAuthKey_Key/Invalid: %#v", m)

		assert.Error(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, "", key.Type())
		assert.Equal(t, "", key.URL())
		assert.Equal(t, uint64(30), key.Period())
		assert.Equal(t, "", key.Secret())
		assert.Equal(t, "", key.String())
		assert.Equal(t, "", key.AccountName())
		assert.Equal(t, "SHA1", key.Algorithm().String())
		assert.Equal(t, otp.DigitsSix, key.Digits())
		assert.Equal(t, 6, key.Digits().Length())
		assert.Equal(t, "", m.RecoveryCode)
	})

}
