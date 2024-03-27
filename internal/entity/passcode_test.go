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

func TestPasscode_Secret(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		uid := "uqxc08w3d0ej2283"
		keyUrl := "otpauth://totp/Example:alice?secret=JBSWY3DPEHPK3PXP"
		recoveryCode := "123"

		m, err := NewPasscode(uid, keyUrl, recoveryCode)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "JBSWY3DPEHPK3PXP", m.Secret())
	})
	t.Run("NoSecret", func(t *testing.T) {
		m := &Passcode{
			UID:          "uqxc08w3d0ej2283",
			KeyURL:       "otpauth://totp/Example:alice",
			RecoveryCode: "123",
		}

		assert.Equal(t, "", m.Secret())
	})
}

func TestPasscode_GenerateCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := &Passcode{
			UID:          "uqxc08w3d0ej2283",
			KeyURL:       "otpauth://totp/Example:alice",
			RecoveryCode: "123",
		}

		code, err := m.GenerateCode()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 6, len(code))
	})
	t.Run("InvalidType", func(t *testing.T) {
		m := &Passcode{
			UID:          "uqxc08w3d0ej2283",
			KeyURL:       "otpauth://xxx/Example:alice",
			KeyType:      "xxx",
			RecoveryCode: "123",
		}

		code, err := m.GenerateCode()

		assert.Error(t, err)

		assert.Equal(t, 0, len(code))
	})
}

func TestPasscode_Verify(t *testing.T) {
	t.Run("ValidCode", func(t *testing.T) {
		m := &Passcode{
			UID:          "uqxc08w3d0ej2283",
			KeyURL:       "otpauth://totp/Example:alice",
			RecoveryCode: "123",
		}

		assert.Nil(t, m.VerifiedAt)

		code, err := m.GenerateCode()

		if err != nil {
			t.Fatal(err)
		}

		valid, recoveryCode, err := m.Verify(code)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, valid)
		assert.False(t, recoveryCode)
		assert.NotNil(t, m.VerifiedAt)
	})
	t.Run("InvalidCode", func(t *testing.T) {
		m := &Passcode{
			UID:          "uqxc08w3d0ej2283",
			KeyURL:       "otpauth://totp/Example:alice",
			RecoveryCode: "123",
		}

		assert.Nil(t, m.VerifiedAt)

		valid, recoveryCode, err := m.Verify("123456")

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, valid)
		assert.False(t, recoveryCode)
		assert.Nil(t, m.VerifiedAt)
	})
	t.Run("CodeTooShort", func(t *testing.T) {
		m := &Passcode{
			UID:          "uqxc08w3d0ej2283",
			KeyURL:       "otpauth://totp/Example:alice",
			RecoveryCode: "123",
		}

		assert.Nil(t, m.VerifiedAt)

		valid, recoveryCode, err := m.Verify("111")

		assert.Error(t, err)
		assert.False(t, valid)
		assert.False(t, recoveryCode)
		assert.Nil(t, m.VerifiedAt)
	})
	t.Run("ValidRecoveryCode", func(t *testing.T) {
		m := &Passcode{
			UID:          "uqxc08w3d0ej2283",
			KeyURL:       "otpauth://totp/Example:alice",
			RecoveryCode: "123",
		}

		assert.Nil(t, m.VerifiedAt)

		valid, recoveryCode, err := m.Verify("123")

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, valid)
		assert.True(t, recoveryCode)
		assert.Nil(t, m.VerifiedAt)
	})
	t.Run("ErrPasscodeRequired", func(t *testing.T) {
		m := &Passcode{
			UID:          "uqxc08w3d0ej2283",
			KeyURL:       "otpauth://totp/Example:alice",
			RecoveryCode: "123",
		}

		assert.Nil(t, m.VerifiedAt)

		valid, recoveryCode, err := m.Verify("")

		assert.Error(t, err)
		assert.False(t, valid)
		assert.False(t, recoveryCode)
		assert.Nil(t, m.VerifiedAt)
	})
	t.Run("ErrInvalidPasscodeFormat", func(t *testing.T) {
		m := &Passcode{
			UID:          "uqxc08w3d0ej2283",
			KeyURL:       "otpauth://totp/Example:alice",
			RecoveryCode: "123",
		}

		assert.Nil(t, m.VerifiedAt)

		valid, recoveryCode, err := m.Verify("123456789112345678911234567891123456789112345678911234567891123456789112345678911234567891123456789112345678911234567891123456789112345678911234567891123456789112345678911234567891123456789112345678911234567891123456789112345678911234567891123456789112345678911234567891123456789112345678911234567891")

		assert.Error(t, err)
		assert.False(t, valid)
		assert.False(t, recoveryCode)
		assert.Nil(t, m.VerifiedAt)
	})
	t.Run("ErrInvalidPasscodeType", func(t *testing.T) {
		m := &Passcode{
			UID:          "uqxc08w3d0ej2283",
			KeyURL:       "otpauth://xxx/Example:alice",
			RecoveryCode: "123",
			KeyType:      "xxx",
		}

		assert.Nil(t, m.VerifiedAt)

		valid, recoveryCode, err := m.Verify("123456")

		assert.Error(t, err)
		assert.False(t, valid)
		assert.False(t, recoveryCode)
		assert.Nil(t, m.VerifiedAt)
	})
}

func TestPasscode_Activate(t *testing.T) {
	m := &Passcode{
		UID:          "uqxc08w3d0ej2283",
		KeyURL:       "otpauth://totp/Example:alice",
		RecoveryCode: "123",
	}

	assert.Nil(t, m.VerifiedAt)
	assert.Nil(t, m.ActivatedAt)

	err := m.Activate()

	assert.Equal(t, authn.ErrPasscodeNotVerified, err)
	assert.Nil(t, m.ActivatedAt)

	code, err := m.GenerateCode()

	if err != nil {
		t.Fatal(err)
	}

	_, _, err = m.Verify(code)

	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, m.VerifiedAt)
	assert.Nil(t, m.ActivatedAt)

	err = m.Activate()

	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, m.ActivatedAt)

	err = m.Activate()

	assert.Equal(t, authn.ErrPasscodeAlreadyActivated, err)
}
