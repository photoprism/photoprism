package oauth

import (
	"errors"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// CodeTTL is the lifetime of an authorization code.
const CodeTTL = time.Minute

// CodeSpec contains the values an authorization code is bound to.
type CodeSpec struct {
	ClientID            string
	ClientName          string
	UserUID             string
	RedirectURI         string
	Scope               string
	Resource            string
	CodeChallenge       string
	CodeChallengeMethod string
}

// validate checks that the values the code must be bound to are present.
func (s CodeSpec) validate() error {
	switch {
	case strings.TrimSpace(s.ClientID) == "":
		return errors.New("oauth: client id required")
	case strings.TrimSpace(s.UserUID) == "":
		return errors.New("oauth: user uid required")
	case strings.TrimSpace(s.RedirectURI) == "":
		return errors.New("oauth: redirect uri required")
	case strings.TrimSpace(s.Resource) == "":
		return errors.New("oauth: resource required")
	case s.CodeChallengeMethod != PKCEMethodS256 || !ValidCodeChallenge(s.CodeChallenge):
		return errors.New("oauth: S256 code challenge required")
	}

	return nil
}

// IssueCode stores a new single-use authorization code bound to spec and returns its raw value,
// which is only passed to the client.
func IssueCode(spec CodeSpec) (string, error) {
	if err := spec.validate(); err != nil {
		return "", err
	}

	expires := time.Now().UTC().Add(CodeTTL).Truncate(time.Second)
	raw, err := newCodeValue(expires)

	if err != nil {
		return "", err
	}

	m := &entity.AuthCode{
		CodeHash:            HashSecret(raw),
		CodeType:            entity.AuthCodeCode,
		ClientID:            spec.ClientID,
		ClientName:          txt.Clip(spec.ClientName, 200),
		UserUID:             spec.UserUID,
		RedirectURI:         spec.RedirectURI,
		Scope:               spec.Scope,
		Resource:            spec.Resource,
		CodeChallenge:       spec.CodeChallenge,
		CodeChallengeMethod: spec.CodeChallengeMethod,
		ExpiresAt:           expires,
	}

	if err = m.Create(); err != nil {
		return "", err
	}

	return raw, nil
}

// RedeemCode consumes the authorization code and returns it if the client ID, redirect URI and code
// verifier match the values it is bound to. A code is consumed by the first attempt, even if it fails,
// and a code that fails VerifyCode is refused before it is looked up.
func RedeemCode(raw, clientID, redirectURI, verifier string) (*entity.AuthCode, error) {
	if err := VerifyCode(raw); err != nil {
		return nil, err
	}

	m, err := entity.FindAuthCode(HashSecret(raw), entity.AuthCodeCode)

	if err != nil {
		return nil, err
	} else if m == nil {
		return nil, ErrInvalidGrant
	}

	if deleted, consumeErr := m.Consume(); consumeErr != nil {
		return nil, consumeErr
	} else if !deleted {
		return nil, ErrInvalidGrant
	}

	switch {
	case m.Expired():
		return nil, ErrInvalidGrant
	case m.UserUID == "" || m.Resource == "":
		return nil, ErrInvalidGrant
	case clientID == "" || m.ClientID != clientID:
		return nil, ErrInvalidGrant
	case redirectURI == "" || m.RedirectURI != redirectURI:
		return nil, ErrInvalidGrant
	case !VerifyPKCE(verifier, m.CodeChallenge, m.CodeChallengeMethod):
		return nil, ErrInvalidGrant
	}

	return m, nil
}

// VerifyCode returns ErrInvalidToken if the authorization code has an invalid shape or signature, and
// ErrTokenExpired if its signed expiry has passed.
func VerifyCode(raw string) error {
	if raw == "" || len(raw) > 128 {
		return ErrInvalidToken
	} else if err := tokens.VerifyCode(raw); errors.Is(err, tokens.ErrExpiredValue) {
		return ErrTokenExpired
	} else if err != nil {
		return ErrInvalidToken
	}

	return nil
}

// newCodeValue returns a new random authorization code value, signed with the expiry.
func newCodeValue(expires time.Time) (string, error) {
	if raw := tokens.SignCode(rnd.AuthToken(), expires); raw != "" {
		return raw, nil
	}

	return "", ErrSigningKey
}

// PurgeExpired deletes expired authorization requests and codes.
func PurgeExpired() {
	if n, err := entity.DeleteExpiredAuthCodes(); err != nil {
		log.Warnf("oauth: failed to delete expired authorization codes (%s)", clean.Error(err))
	} else if n > 0 {
		log.Debugf("oauth: deleted %d expired authorization codes", n)
	}
}
