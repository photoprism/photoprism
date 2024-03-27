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
	t.Run("InvalidUID", func(t *testing.T) {
		m, err := NewPasscode("foo", "", "")

		t.Logf("TestNewAuthKey/Invalid: %#v", m)

		assert.Error(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, "foo", m.UID)
		assert.True(t, authn.KeyTOTP.NotEqual(m.KeyType))
		assert.Equal(t, "", m.RecoveryCode)
	})
	t.Run("EmptyUrl", func(t *testing.T) {
		m, err := NewPasscode("us7gqkzx1g9axxxx", "", "")

		t.Logf("TestNewAuthKey/Invalid: %#v", m)

		assert.Error(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, "us7gqkzx1g9axxxx", m.UID)
		assert.True(t, authn.KeyTOTP.NotEqual(m.KeyType))
		assert.Equal(t, "", m.RecoveryCode)
	})
	t.Run("InvalidUrl", func(t *testing.T) {
		m, err := NewPasscode("us7gqkzx1g9axxxx", "avcgo6842485%^^&", "")

		t.Logf("TestNewAuthKey/Invalid: %#v", m)

		assert.Error(t, err)
		assert.NotNil(t, m)
		assert.Equal(t, "us7gqkzx1g9axxxx", m.UID)
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

func TestPasscode_Delete(t *testing.T) {
	t.Run("UidNotSet", func(t *testing.T) {
		m := &Passcode{
			UID:          "",
			KeyURL:       "otpauth://totp/PhotoPrism:bob?algorithm=SHA1&digits=6&issuer=PhotoPrism%20Pro&period=30",
			RecoveryCode: "123",
		}

		err := m.Delete()

		assert.Error(t, err)
	})
}

func TestPasscode_SetUID(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		m := &Passcode{
			UID:          "123",
			KeyURL:       "",
			RecoveryCode: "123",
		}

		assert.True(t, m.InvalidUID())

		passcode := m.SetUID("uqxc08w3d0ej2283")

		assert.False(t, passcode.InvalidUID())
	})

	t.Run("Invalid", func(t *testing.T) {
		m := &Passcode{
			UID:          "uqxc08w3d0ej2283",
			KeyURL:       "",
			RecoveryCode: "123",
		}

		assert.False(t, m.InvalidUID())

		passcode := m.SetUID("xxx")

		assert.False(t, passcode.InvalidUID())
	})
}

func TestPasscode_SetKey(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		uid := "uqxc08w3d0ej2283"
		keyUrl := "otpauth://totp/Example:alice?secret=JBSWY3DPEHPK3PXP"
		recoveryCode := "123"

		m, err := NewPasscode(uid, keyUrl, recoveryCode)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "otpauth://totp/Example:alice?secret=JBSWY3DPEHPK3PXP", m.Key().String())

		key, err := otp.NewKeyFromURL("otpauth://totp/Example:bob?secret=JBSWY3DPEHPK3PXP")

		if err != nil {
			t.Fatal(err)
		}

		err = m.SetKey(key)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "otpauth://totp/Example:bob?secret=JBSWY3DPEHPK3PXP", m.Key().String())
	})
	t.Run("InvalidKeyType", func(t *testing.T) {
		uid := "uqxc08w3d0ej2283"
		keyUrl := "otpauth://totp/Example:alice?secret=JBSWY3DPEHPK3PXP"
		recoveryCode := "123"

		m, err := NewPasscode(uid, keyUrl, recoveryCode)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "otpauth://totp/Example:alice?secret=JBSWY3DPEHPK3PXP", m.Key().String())

		key, err := otp.NewKeyFromURL("otpauth://xxx/Example:bob?secret=JBSWY3DPEHPK3PXP")

		if err != nil {
			t.Fatal(err)
		}

		err = m.SetKey(key)

		assert.Error(t, err)
		assert.Equal(t, "otpauth://totp/Example:alice?secret=JBSWY3DPEHPK3PXP", m.Key().String())
	})
	t.Run("NoSecret", func(t *testing.T) {
		uid := "uqxc08w3d0ej2283"
		keyUrl := "otpauth://totp/Example:alice?secret=JBSWY3DPEHPK3PXP"
		recoveryCode := "123"

		m, err := NewPasscode(uid, keyUrl, recoveryCode)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "otpauth://totp/Example:alice?secret=JBSWY3DPEHPK3PXP", m.Key().String())

		key, err := otp.NewKeyFromURL("otpauth://totp/Example:bob")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", key.Secret())

		err = m.SetKey(key)

		assert.Error(t, err)
		assert.Equal(t, "otpauth://totp/Example:alice?secret=JBSWY3DPEHPK3PXP", m.Key().String())
	})
}
