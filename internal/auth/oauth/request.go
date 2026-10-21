package oauth

import (
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// RequestTTL is how long a pending authorization request waits for the user to sign in and decide.
const RequestTTL = 15 * time.Minute

// ErrRequestNotFound is returned when a pending authorization request does not exist or has expired.
var ErrRequestNotFound = errors.New("authorization request not found")

// RequestSpec contains the validated values of an authorization request.
type RequestSpec struct {
	ClientID            string
	ClientName          string
	RedirectURI         string
	Scopes              []string
	Resource            string
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
}

// NewRequest stores a pending authorization request and returns the raw request ID, which only the
// user's browser receives.
func NewRequest(spec RequestSpec) (string, error) {
	switch {
	case spec.ClientID == "" || spec.RedirectURI == "" || spec.Resource == "" || len(spec.Scopes) == 0:
		return "", errors.New("oauth: incomplete authorization request")
	case spec.CodeChallengeMethod != PKCEMethodS256 || !ValidCodeChallenge(spec.CodeChallenge):
		return "", errors.New("oauth: S256 code challenge required")
	}

	raw := rnd.AuthToken()

	m := &entity.AuthCode{
		CodeHash:            HashSecret(raw),
		CodeType:            entity.AuthCodeRequest,
		ClientID:            spec.ClientID,
		ClientName:          txt.Clip(spec.ClientName, 200),
		RedirectURI:         spec.RedirectURI,
		Scope:               JoinScopes(spec.Scopes),
		Resource:            spec.Resource,
		State:               spec.State,
		CodeChallenge:       spec.CodeChallenge,
		CodeChallengeMethod: spec.CodeChallengeMethod,
		ExpiresAt:           time.Now().UTC().Add(RequestTTL),
	}

	if err := m.Create(); err != nil {
		return "", err
	}

	return raw, nil
}

// FindRequest returns the pending authorization request with the specified raw ID, or ErrRequestNotFound.
func FindRequest(raw string) (*entity.AuthCode, error) {
	if raw == "" || len(raw) > 128 {
		return nil, ErrRequestNotFound
	}

	m, err := entity.FindAuthCode(HashSecret(raw), entity.AuthCodeRequest)

	if err != nil {
		return nil, err
	} else if m == nil || m.Expired() {
		return nil, ErrRequestNotFound
	}

	return m, nil
}

// RequestScopes returns the scopes of a pending authorization request.
func RequestScopes(m *entity.AuthCode) []string {
	if m == nil {
		return nil
	}

	return strings.Fields(m.Scope)
}

// CompleteRequest turns an approved pending request into a single-use authorization code for the user and the
// approved scopes, which must be a non-empty subset of the requested ones, and returns the raw code.
func CompleteRequest(m *entity.AuthCode, userUID string, scopes []string) (string, error) {
	if m == nil || m.CodeType != entity.AuthCodeRequest || m.Expired() {
		return "", ErrRequestNotFound
	} else if len(scopes) == 0 {
		return "", errors.New("oauth: no scope approved")
	}

	requested := RequestScopes(m)

	for _, s := range scopes {
		if !slices.Contains(requested, s) {
			return "", errors.New("oauth: scope was not requested")
		}
	}

	spec := CodeSpec{
		ClientID:            m.ClientID,
		ClientName:          m.ClientName,
		UserUID:             userUID,
		RedirectURI:         m.RedirectURI,
		Scope:               JoinScopes(scopes),
		Resource:            m.Resource,
		CodeChallenge:       m.CodeChallenge,
		CodeChallengeMethod: m.CodeChallengeMethod,
	}

	if err := spec.validate(); err != nil {
		return "", err
	}

	raw, err := newCodeValue()

	if err != nil {
		return "", err
	}

	changed, err := m.Replace(HashSecret(raw), entity.AuthCodeCode, entity.Values{
		"user_uid":   spec.UserUID,
		"scope":      spec.Scope,
		"expires_at": time.Now().UTC().Add(CodeTTL),
	})

	if err != nil {
		return "", err
	} else if !changed {
		return "", ErrRequestNotFound
	}

	return raw, nil
}

// DenyRequest deletes a pending authorization request, so it can no longer be approved.
func DenyRequest(m *entity.AuthCode) error {
	if m == nil || m.CodeType != entity.AuthCodeRequest {
		return ErrRequestNotFound
	}

	if deleted, err := m.Consume(); err != nil {
		return err
	} else if !deleted {
		return ErrRequestNotFound
	}

	return nil
}
