package oauth

import (
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Parameter bounds of authorization requests, matching the column sizes of auth_codes.
const (
	MaxClientIDLength    = 512
	MaxRedirectURILength = 2048
	MaxScopeLength       = 1024
	MaxStateLength       = 1024
	MaxResourceLength    = 512
)

// RequestError represents an OAuth error with its error code and description.
type RequestError struct {
	Code        string
	Description string
}

// Error returns the error description.
func (e *RequestError) Error() string {
	if e == nil {
		return ""
	}

	return e.Description
}

// NewRequestError returns a new OAuth error with the specified code and description.
func NewRequestError(code, description string) *RequestError {
	return &RequestError{Code: code, Description: description}
}

// AuthorizeRequest represents the parameters of an authorization request.
type AuthorizeRequest struct {
	ResponseType        string
	ClientID            string
	RedirectURI         string
	Scope               string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
	Resource            string
	duplicates          []string
}

// authorizeParams lists the parameters of an authorization request that must not be repeated.
var authorizeParams = []string{"response_type", "client_id", "redirect_uri", "scope", "state", "code_challenge", "code_challenge_method", "resource"}

// ParseAuthorizeRequest returns the authorization request parameters contained in the query values.
func ParseAuthorizeRequest(values url.Values) *AuthorizeRequest {
	r := &AuthorizeRequest{
		ResponseType:        strings.TrimSpace(values.Get("response_type")),
		ClientID:            strings.TrimSpace(values.Get("client_id")),
		RedirectURI:         strings.TrimSpace(values.Get("redirect_uri")),
		Scope:               strings.Join(strings.Fields(values.Get("scope")), " "),
		State:               values.Get("state"),
		CodeChallenge:       strings.TrimSpace(values.Get("code_challenge")),
		CodeChallengeMethod: strings.TrimSpace(values.Get("code_challenge_method")),
		Resource:            strings.TrimSpace(values.Get("resource")),
	}

	for _, name := range authorizeParams {
		if len(values[name]) > 1 {
			r.duplicates = append(r.duplicates, name)
		}
	}

	return r
}

// ValidateClient checks the syntax of the client ID and redirect URI. Errors may be redirected to the client
// only after the redirect URI has also been matched against the client's registration with MatchRedirectURI.
func (r *AuthorizeRequest) ValidateClient() *RequestError {
	switch {
	case r == nil:
		return NewRequestError(ErrCodeInvalidRequest, "invalid request")
	case r.repeated("client_id"):
		return NewRequestError(ErrCodeInvalidRequest, "client_id must not be repeated")
	case r.repeated("redirect_uri"):
		return NewRequestError(ErrCodeInvalidRequest, "redirect_uri must not be repeated")
	case r.ClientID == "":
		return NewRequestError(ErrCodeInvalidRequest, "client_id required")
	case !ValidValue(r.ClientID, MaxClientIDLength) || strings.ContainsFunc(r.ClientID, unicode.IsSpace):
		return NewRequestError(ErrCodeInvalidRequest, "client_id invalid")
	case r.RedirectURI == "":
		return NewRequestError(ErrCodeInvalidRequest, "redirect_uri required")
	case !ValidValue(r.RedirectURI, MaxRedirectURILength) || !ValidRedirectURI(r.RedirectURI):
		return NewRequestError(ErrCodeInvalidRequest, "redirect_uri invalid")
	}

	return nil
}

// Validate checks the remaining parameters once the client and redirect URI have been verified.
func (r *AuthorizeRequest) Validate() *RequestError {
	if err := r.ValidateClient(); err != nil {
		return err
	}

	switch {
	case len(r.duplicates) > 0:
		return NewRequestError(ErrCodeInvalidRequest, r.duplicates[0]+" must not be repeated")
	case r.ResponseType == "":
		return NewRequestError(ErrCodeInvalidRequest, "response_type required")
	case r.ResponseType != "code":
		return NewRequestError(ErrCodeUnsupportedResponse, "response_type must be code")
	case r.CodeChallenge == "":
		return NewRequestError(ErrCodeInvalidRequest, "code_challenge required")
	case r.CodeChallengeMethod != PKCEMethodS256:
		return NewRequestError(ErrCodeInvalidRequest, "code_challenge_method must be S256")
	case !ValidCodeChallenge(r.CodeChallenge):
		return NewRequestError(ErrCodeInvalidRequest, "code_challenge must be 43 base64url characters")
	case !ValidValue(r.Scope, MaxScopeLength):
		return NewRequestError(ErrCodeInvalidScope, "scope invalid")
	case !ValidValue(r.State, MaxStateLength):
		return NewRequestError(ErrCodeInvalidRequest, "state invalid")
	case r.Resource != "" && (!ValidValue(r.Resource, MaxResourceLength) || !ValidAbsoluteURL(r.Resource)):
		return NewRequestError(ErrCodeInvalidTarget, "resource invalid")
	}

	return nil
}

// repeated reports whether the named parameter was included more than once.
func (r *AuthorizeRequest) repeated(name string) bool {
	for _, d := range r.duplicates {
		if d == name {
			return true
		}
	}

	return false
}

// Scopes returns the requested scopes.
func (r *AuthorizeRequest) Scopes() []string {
	if r == nil {
		return nil
	}

	return strings.Fields(r.Scope)
}

// ValidValue reports whether s is valid UTF-8 without control, format or line separator characters
// and at most maxLen bytes long.
func ValidValue(s string, maxLen int) bool {
	if len(s) > maxLen || !utf8.ValidString(s) {
		return false
	}

	return !strings.ContainsFunc(s, invalidRune)
}

// invalidRune reports whether r is a control, format, line or paragraph separator character.
func invalidRune(r rune) bool {
	return unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zl, unicode.Zp)
}

// ValidAbsoluteURL reports whether s is an absolute URL with a host and without a fragment or user info.
func ValidAbsoluteURL(s string) bool {
	u, err := url.Parse(s)

	if err != nil || u.Scheme == "" || u.Host == "" || u.Fragment != "" || u.User != nil {
		return false
	}

	return !strings.HasSuffix(s, "#")
}
