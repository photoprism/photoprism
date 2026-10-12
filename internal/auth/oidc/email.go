package oidc

import (
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/pkg/authn"
)

// VerifiedEmail returns the email address of the identity if the provider marks it verified, it is valid,
// and it contains only ASCII characters, or an empty string otherwise.
func VerifiedEmail(userInfo *oidc.UserInfo) string {
	if userInfo == nil {
		return ""
	}

	return authn.VerifiedEmail(userInfo.Email, bool(userInfo.EmailVerified))
}
