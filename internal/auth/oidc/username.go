package oidc

import (
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
)

// Username returns the preferred username based on the userinfo and the preferred username OIDC claim.
// A claim with characters other than ASCII is skipped, since new usernames must be ASCII.
func Username(userInfo *oidc.UserInfo, preferredClaim string) string {
	return username(userInfo, preferredClaim, true)
}

// SkipsClaim reports whether Username skips a claim with characters other than ASCII that it would
// otherwise use.
func SkipsClaim(userInfo *oidc.UserInfo, preferredClaim string) bool {
	return username(userInfo, preferredClaim, false) != Username(userInfo, preferredClaim)
}

// asciiHandle returns the handle of a claim, or an empty string if the claim contains characters other
// than ASCII.
func asciiHandle(claim string) string {
	if !authn.IsASCIIUsername(claim) {
		return ""
	}

	return clean.Handle(claim)
}

// asciiEmail returns the sanitized email address, or an empty string if it contains characters other
// than ASCII.
func asciiEmail(email string) string {
	if !authn.IsASCIIUsername(email) {
		return ""
	}

	return clean.Email(email)
}

// username returns the first non-empty handle of the claims in the order the preferred claim selects,
// skipping claims with characters other than ASCII if ascii is true.
func username(userInfo *oidc.UserInfo, preferredClaim string, ascii bool) (userName string) {
	handle, email := clean.Handle, clean.Email

	if ascii {
		handle, email = asciiHandle, asciiEmail
	}

	switch preferredClaim {
	case authn.OidcClaimName:
		if name := handle(userInfo.Name); len(name) > 0 {
			userName = name
		} else if name = handle(userInfo.PreferredUsername); len(name) > 0 {
			userName = name
		} else if name = handle(userInfo.Nickname); len(name) > 0 {
			userName = name
		} else if name = email(userInfo.Email); userInfo.EmailVerified && len(name) > 4 && clean.Handle(name) != "" {
			userName = name
		}
	case authn.OidcClaimNickname:
		if name := handle(userInfo.Nickname); len(name) > 0 {
			userName = name
		} else if name = handle(userInfo.PreferredUsername); len(name) > 0 {
			userName = name
		} else if name = handle(userInfo.Name); len(name) > 0 {
			userName = name
		} else if name = email(userInfo.Email); userInfo.EmailVerified && len(name) > 4 && clean.Handle(name) != "" {
			userName = name
		}
	case authn.OidcClaimEmail:
		if name := email(userInfo.Email); userInfo.EmailVerified && len(name) > 4 && clean.Handle(name) != "" {
			userName = name
		} else if name = handle(userInfo.PreferredUsername); len(name) > 0 {
			userName = name
		} else if name = handle(userInfo.Name); len(name) > 0 {
			userName = name
		} else if name = handle(userInfo.Nickname); len(name) > 0 {
			userName = name
		}
	default:
		if name := handle(userInfo.PreferredUsername); len(name) > 0 {
			userName = name
		} else if name = handle(userInfo.Name); len(name) > 0 {
			userName = name
		} else if name = handle(userInfo.Nickname); len(name) > 0 {
			userName = name
		} else if name = email(userInfo.Email); userInfo.EmailVerified && len(name) > 4 && clean.Handle(name) != "" {
			userName = name
		}
	}

	return userName
}
