package oidc

import (
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
)

// Username returns the preferred username based on the userinfo and the preferred username OIDC claim.
func Username(userInfo *oidc.UserInfo, preferredClaim string) (userName string) {
	switch preferredClaim {
	case authn.ClaimName:
		if name := clean.Handle(userInfo.Name); len(name) > 0 {
			userName = name
		} else if name = clean.Handle(userInfo.PreferredUsername); len(name) > 0 {
			userName = name
		} else if name = clean.Handle(userInfo.Nickname); len(name) > 0 {
			userName = name
		} else if name = clean.Email(userInfo.Email); userInfo.EmailVerified && len(name) > 4 {
			userName = name
		}
	case authn.ClaimNickname:
		if name := clean.Handle(userInfo.Nickname); len(name) > 0 {
			userName = name
		} else if name = clean.Handle(userInfo.PreferredUsername); len(name) > 0 {
			userName = name
		} else if name = clean.Handle(userInfo.Name); len(name) > 0 {
			userName = name
		} else if name = clean.Email(userInfo.Email); userInfo.EmailVerified && len(name) > 4 {
			userName = name
		}
	case authn.ClaimEmail:
		if name := clean.Email(userInfo.Email); userInfo.EmailVerified && len(name) > 4 {
			userName = name
		} else if name = clean.Handle(userInfo.PreferredUsername); len(name) > 0 {
			userName = name
		} else if name = clean.Handle(userInfo.Name); len(name) > 0 {
			userName = name
		} else if name = clean.Handle(userInfo.Nickname); len(name) > 0 {
			userName = name
		}
	default:
		if name := clean.Handle(userInfo.PreferredUsername); len(name) > 0 {
			userName = name
		} else if name = clean.Handle(userInfo.Name); len(name) > 0 {
			userName = name
		} else if name = clean.Handle(userInfo.Nickname); len(name) > 0 {
			userName = name
		} else if name = clean.Email(userInfo.Email); userInfo.EmailVerified && len(name) > 4 {
			userName = name
		}
	}

	return userName
}
