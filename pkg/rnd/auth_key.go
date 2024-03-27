package rnd

import (
	"fmt"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// AuthKey returns a randomly initialized, time-based one-time password key, e.g. for 2-factor authentication.
func AuthKey(issuer, accountName string) (key *otp.Key, err error) {
	if issuer = clip(issuer, 64); issuer == "" {
		return nil, fmt.Errorf("invalid site name")
	}

	if accountName = clip(accountName, 64); accountName == "" {
		return nil, fmt.Errorf("invalid user name")
	}

	key, err = totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})

	return key, err
}

// RecoveryCode returns a randomly created recovery code for e.g. for 2-factor authentication.
func RecoveryCode() (code string) {
	return Base36(12)
}
