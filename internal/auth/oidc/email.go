package oidc

import (
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/pkg/clean"
)

// VerifiedEmail returns the email address of the identity if the provider marks it verified and it is valid,
// or an empty string otherwise.
func VerifiedEmail(userInfo *oidc.UserInfo) string {
	if userInfo == nil || !bool(userInfo.EmailVerified) {
		return ""
	}

	return clean.Email(userInfo.Email)
}
