package form

// OAuthConsent represents a decision on a pending OAuth authorization request.
type OAuthConsent struct {
	Request string   `json:"request" binding:"required"`
	Scopes  []string `json:"scopes,omitempty"`
}
