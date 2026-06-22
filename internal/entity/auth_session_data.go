package entity

import (
	"slices"
	"strings"
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

// RefreshShares updates the list of shared UIDs in the session data.
func (data *SessionData) RefreshShares() *SessionData {
	var shares []string

	for _, token := range data.Tokens {
		links := FindValidLinks(token, "")

		if len(links) == 0 {
			continue
		}

		for _, link := range links {
			shares = append(shares, link.ShareUID)
		}
	}

	data.Shares = shares

	return data
}

// RedeemToken appends a new token and updates the list of shared UIDs in the session data.
func (data *SessionData) RedeemToken(token string) (n int) {
	links := FindValidLinks(token, "")

	// No valid links found?
	if n = len(links); n == 0 {
		return n
	}

	// Append new token.
	data.Tokens = append(data.Tokens, token)

	// Append new shares.
	for _, link := range links {
		data.Shares = append(data.Shares, link.ShareUID)
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
