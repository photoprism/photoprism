package oidc

import (
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
)

// Username returns the preferred username based on the userinfo and the preferred username OIDC claim.
// Once a claim with characters other than ASCII is skipped, only the verified email address may follow,
// since new usernames must be ASCII and must not come from a claim the user can freely edit.
func Username(userInfo *oidc.UserInfo, preferredClaim string) string {
	return username(userInfo, preferredClaim, true)
}

// SkipsClaim reports whether skipping claims with characters other than ASCII changes the username.
func SkipsClaim(userInfo *oidc.UserInfo, preferredClaim string) bool {
	return username(userInfo, preferredClaim, false) != Username(userInfo, preferredClaim)
}

// claimOrder returns the claims in the order the preferred claim selects.
func claimOrder(preferredClaim string) []string {
	switch preferredClaim {
	case authn.OidcClaimName:
		return []string{authn.OidcClaimName, authn.OidcClaimPreferredUsername, authn.OidcClaimNickname, authn.OidcClaimEmail}
	case authn.OidcClaimNickname:
		return []string{authn.OidcClaimNickname, authn.OidcClaimPreferredUsername, authn.OidcClaimName, authn.OidcClaimEmail}
	case authn.OidcClaimEmail:
		return []string{authn.OidcClaimEmail, authn.OidcClaimPreferredUsername, authn.OidcClaimName, authn.OidcClaimNickname}
	default:
		return []string{authn.OidcClaimPreferredUsername, authn.OidcClaimName, authn.OidcClaimNickname, authn.OidcClaimEmail}
	}
}

// verifiedEmailUsername returns the verified email address if it is a valid username, or an empty string.
// Callers check that the address is ASCII as provided, since sanitizing it lowercases first.
func verifiedEmailUsername(userInfo *oidc.UserInfo) string {
	if name := clean.Email(userInfo.Email); userInfo.EmailVerified && len(name) > 4 && clean.Handle(name) != "" {
		return name
	}

	return ""
}

// username returns the first usable claim in the order the preferred claim selects. If ascii is true,
// a claim with characters other than ASCII, other than an unverified email address, ends the search,
// and only an ASCII verified email address may be used instead.
func username(userInfo *oidc.UserInfo, preferredClaim string, ascii bool) string {
	for _, claim := range claimOrder(preferredClaim) {
		var value string

		switch claim {
		case authn.OidcClaimEmail:
			if !userInfo.EmailVerified {
				continue
			}

			value = userInfo.Email
		case authn.OidcClaimName:
			value = userInfo.Name
		case authn.OidcClaimNickname:
			value = userInfo.Nickname
		default:
			value = userInfo.PreferredUsername
		}

		if ascii && !authn.IsASCIIUsername(value) {
			if claim != authn.OidcClaimEmail && authn.IsASCIIUsername(userInfo.Email) {
				return verifiedEmailUsername(userInfo)
			}

			return ""
		} else if claim == authn.OidcClaimEmail {
			if name := verifiedEmailUsername(userInfo); name != "" {
				return name
			}
		} else if name := clean.Handle(value); name != "" {
			return name
		}
	}

	return ""
}
