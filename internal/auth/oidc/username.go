package oidc

import (
	"strings"
	"unicode"

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
// last resort. Handles are derived with runs of dots collapsed, those of names and nicknames without punctuation
// at their ends, and names that are not portable usernames are skipped; if ascii is true, also values with
// characters other than ASCII.
func username(userInfo *oidc.UserInfo, preferredClaim string, ascii bool) string {
	usable := func(value string) bool {
		return !ascii || authn.IsASCIIUsername(value)
	}

	emailClaim := preferredClaim == authn.OidcClaimEmail

	// With the email claim, a verified address is the username, and one with other characters is refused.
	if emailClaim && bool(userInfo.EmailVerified) {
		if !usable(userInfo.Email) || !authn.PortableUsername(userInfo.Email) {
			return ""
		} else if name := verifiedEmailUsername(userInfo); name != "" {
			return name
		}
	}

	for _, claim := range claimOrder(preferredClaim) {
		value := claimValue(userInfo, claim)

		if !usable(value) {
			continue
		}

		// Names and nicknames are display text, so punctuation at their edges is not part of the username, also
		// where the handle maps it to a dot.
		var name string

		if claim == authn.OidcClaimName || claim == authn.OidcClaimNickname {
			name = trimText(handle(trimText(value)))
		} else {
			name = handle(value)
		}

		if name != "" && authn.PortableUsername(name) {
			return name
		}
	}

	if !emailClaim && usable(userInfo.Email) {
		if name := handle(verifiedEmailUsername(userInfo)); name != "" && authn.PortableUsername(name) {
			return name
		}
	}

	return ""
}

// handle returns the handle of a claim value with runs of dots collapsed, so that e.g. "John F. Kennedy" gives
// "john.f.kennedy" rather than a name with "..".
func handle(value string) string {
	name := clean.Handle(value)

	for strings.Contains(name, "..") {
		name = strings.ReplaceAll(name, "..", ".")
	}

	return name
}

// trimText removes punctuation, symbols, and whitespace from both ends of a name derived from display text.
func trimText(name string) string {
	return strings.TrimFunc(name, func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsSpace(r)
	})
}
