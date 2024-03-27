package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pquerna/otp"
)

func TestAuthKey(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuer := "Foo"
		accountName := "Bar Baz"
		result, err := AuthKey(issuer, accountName)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// Example String URL:
		// otpauth://totp/Foo:Bar%20Baz?algorithm=SHA1&digits=6&issuer=Foo&period=30&secret=25I6JAIZRNSOPN2WMADP6HX5GTCYTF36
		assert.Len(t, result.String(), 113)
		assert.Equal(t, issuer, result.Issuer())
		assert.Equal(t, accountName, result.AccountName())
		assert.Equal(t, "SHA1", result.Algorithm().String())
		assert.Equal(t, otp.Digits(6), result.Digits())
		assert.Equal(t, uint64(30), result.Period())
	})
	t.Run("EmptyIssuer", func(t *testing.T) {
		issuer := ""
		accountName := "Bar Baz"
		result, err := AuthKey(issuer, accountName)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
	t.Run("EmptyAccountName", func(t *testing.T) {
		issuer := "foo"
		accountName := ""
		result, err := AuthKey(issuer, accountName)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestRecoveryCode(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result := RecoveryCode()

		t.Logf("Recovery Code: %s", result)
		assert.Len(t, result, 12)
	})
}
