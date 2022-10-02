package entity

import (
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
	Tokens []string `json:"tokens"` // Share Tokens.
	Shares UIDs     `json:"shares"` // Share UIDs.
}

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

	for _, share := range data.Shares {
		if share == uid {
			return true
		}
	}

	return false
}

// SharedUIDs returns shared entity UIDs.
func (data SessionData) SharedUIDs() UIDs {
	if len(data.Tokens) > 0 && len(data.Shares) == 0 {
		data.RefreshShares()
	}

	return data.Shares
}
