package entity

import (
	"slices"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
)

// UIDs represents a slice of unique ID strings.
type UIDs []string

// String returns all UIDs as comma separated string.
func (u UIDs) String() string {
	return u.Join(",")
}

// Join returns all UIDs as custom separated string.
func (u UIDs) Join(s string) string {
	return strings.Join(u, s)
}

// SessionData represents User Session data.
type SessionData struct {
	Tokens []string `json:"tokens"`           // Share Tokens.
	Shares UIDs     `json:"shares"`           // Share UIDs.
	Links  UIDs     `json:"links,omitempty"`  // UIDs of the links the tokens were redeemed through.
	Groups []string `json:"groups,omitempty"` // Normalized login-time group identifiers (OIDC/LDAP).
}

// SessionGroupsByteLimit caps the serialized size of SessionData.Groups so the
// session data always fits its 16384-byte database column with room to spare
// for share tokens redeemed later in the session's lifetime. The budget holds
// roughly 300 GUID-sized identifiers — beyond Entra's 200-group overage
// threshold, where the IdP stops emitting groups in tokens altogether.
const SessionGroupsByteLimit = 12288

// NewSessionData creates a new session data struct and returns a pointer to it.
func NewSessionData() *SessionData {
	return &SessionData{}
}

// RedeemedLinks returns the links a token resolves that the session still holds: those within their
// expiration time that either admit a redemption or already admitted this session.
func (data *SessionData) RedeemedLinks(token string) (found Links) {
	for _, link := range FindRedeemedLinksByToken(token, "") {
		if !link.Redeemable() && !slices.Contains(data.Links, link.LinkUID) {
			continue
		}

		found = append(found, link)
	}

	return found
}

// RefreshShares updates the list of shared UIDs in the session data. A redeemed link keeps its share
// until it is removed or reaches its expiration time, so the view limit bounds how many sessions a
// link admits rather than how long each of them lasts.
func (data *SessionData) RefreshShares() *SessionData {
	var shares []string

	for _, token := range data.Tokens {
		for _, link := range data.RedeemedLinks(token) {
			shares = append(shares, link.ShareUID)
		}
	}

	data.Shares = shares

	return data
}

// RedeemToken appends a new token and updates the list of shared UIDs in the session data.
// The token is stored in its sanitized form so it still resolves on the next lookup, and a value
// that cannot be sanitized is refused before the query runs.
func (data *SessionData) RedeemToken(token string) (n int) {
	if token = clean.ShareToken(token); token == "" {
		return 0
	}

	// A token the session already holds needs no new redemption, as the sharing page redeems on
	// every load. It reports what the session still holds through it, without counting another view.
	if slices.Contains(data.Tokens, token) {
		return len(data.RedeemedLinks(token))
	}

	links := FindRedeemableLinksByToken(token, "")

	// No redeemable links found?
	if n = len(links); n == 0 {
		return n
	}

	// Append new token.
	data.Tokens = append(data.Tokens, token)

	// Append the shares and the links they were redeemed through.
	for _, link := range links {
		data.Shares = append(data.Shares, link.ShareUID)
		data.Links = append(data.Links, link.LinkUID)
		link.Redeem()
	}

	return n
}

// SetGroups stores the user's normalized login-time group identifiers,
// dropping trailing entries once their serialized size would exceed
// SessionGroupsByteLimit so the session data cannot outgrow its column.
func (data *SessionData) SetGroups(groups []string) *SessionData {
	size := 0

	for i, g := range groups {
		// Each entry serializes as a quoted string plus a separator.
		if size += len(g) + 3; size > SessionGroupsByteLimit {
			log.Warnf("auth: session group set truncated to %d of %d entries", i, len(groups))
			groups = groups[:i]
			break
		}
	}

	if len(groups) == 0 {
		data.Groups = nil
	} else {
		data.Groups = groups
	}

	return data
}

// Redacted returns a copy of the session data without server-side fields (the
// login-time group set), so API session responses never disclose the user's
// upstream group memberships to clients.
func (data *SessionData) Redacted() *SessionData {
	if data == nil {
		return nil
	}

	redacted := *data
	redacted.Groups = nil

	return &redacted
}

// NoShares checks if the session has no shares yet.
func (data SessionData) NoShares() bool {
	return len(data.Shares) == 0
}

// HasShares checks if the session has any shares.
func (data SessionData) HasShares() bool {
	return len(data.Shares) > 0
}

// HasShare if the session includes the specified share
func (data SessionData) HasShare(uid string) bool {
	if uid == "" || data.NoShares() {
		return false
	}

	return slices.Contains(data.Shares, uid)
}

// SharedUIDs returns shared entity UIDs.
func (data SessionData) SharedUIDs() UIDs {
	if len(data.Tokens) > 0 && len(data.Shares) == 0 {
		data.RefreshShares()
	}

	return data.Shares
}
