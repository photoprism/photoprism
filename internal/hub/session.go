package hub

import "time"

// Session represents backend api session data.
type Session struct {
	Tier      int
	MapKey    string
	Customer  string `json:",omitempty"`
	ExpiresAt string
}

// Expired tests if the api session is expired.
func (p *Session) Expired() bool {
	if p.ExpiresAt == "" {
		return true
	} else if date, err := time.Parse("2006-01-02T15:04:05", p.ExpiresAt); err != nil {
		return true
	} else if date.Before(time.Now()) {
		return true
	}

	return false
}
