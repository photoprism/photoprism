package oidc

import (
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
)

// Username returns the username for a new account based on the userinfo and the preferred username claim.
// Claims with characters other than ASCII are skipped, since new usernames must be ASCII.
func Username(userInfo *oidc.UserInfo, preferredClaim string) string {
	return username(userInfo, preferredClaim, true)
}

// SkipsClaim reports whether skipping claims with characters other than ASCII changes the username.
func SkipsClaim(userInfo *oidc.UserInfo, preferredClaim string) bool {
	return username(userInfo, preferredClaim, false) != Username(userInfo, preferredClaim)
}

// claimOrder returns the name claims in the order the preferred claim selects.
func claimOrder(preferredClaim string) []string {
	switch preferredClaim {
	case authn.OidcClaimName:
		return []string{authn.OidcClaimName, authn.OidcClaimPreferredUsername, authn.OidcClaimNickname}
	case authn.OidcClaimNickname:
		return []string{authn.OidcClaimNickname, authn.OidcClaimPreferredUsername, authn.OidcClaimName}
	default:
		return []string{authn.OidcClaimPreferredUsername, authn.OidcClaimName, authn.OidcClaimNickname}
	}
}

// claimValue returns the value of a name claim.
func claimValue(userInfo *oidc.UserInfo, claim string) string {
	switch claim {
	case authn.OidcClaimName:
		return userInfo.Name
	case authn.OidcClaimNickname:
		return userInfo.Nickname
	default:
		return userInfo.PreferredUsername
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

// username returns the username for the preferred claim: with the email claim, the verified address, else the
// first name claim that yields a handle, and with the other claims the handle of the verified address as the
// last resort. If ascii is true, values with characters other than ASCII are skipped.
func username(userInfo *oidc.UserInfo, preferredClaim string, ascii bool) string {
	usable := func(value string) bool {
		return !ascii || authn.IsASCIIUsername(value)
	}

	emailClaim := preferredClaim == authn.OidcClaimEmail

	// With the email claim, a verified address is the username, and one with other characters is refused.
	if emailClaim && bool(userInfo.EmailVerified) {
		if !usable(userInfo.Email) {
			return ""
		} else if name := verifiedEmailUsername(userInfo); name != "" {
			return name
		}
	}

	for _, claim := range claimOrder(preferredClaim) {
		if value := claimValue(userInfo, claim); !usable(value) {
			continue
		} else if name := clean.Handle(value); name != "" {
			return name
		}
	}

	if !emailClaim && usable(userInfo.Email) {
		if name := verifiedEmailUsername(userInfo); name != "" {
			return clean.Handle(name)
		}
	}

	return ""
}
