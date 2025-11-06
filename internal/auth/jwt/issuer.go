package jwt

import (
	"errors"
	"strings"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"

	"github.com/photoprism/photoprism/pkg/rnd"
)

const PrefixTokenID = "jwt"

var (
	// DefaultTokenTTL is the default lifetime for issued tokens.
	DefaultTokenTTL = 300 * time.Second
	// MaxTokenTTL clamps configurable lifetimes to a safe upper bound.
	MaxTokenTTL = 900 * time.Second
)

// TokenTTL controls the default lifetime used when a ClaimsSpec does not override TTL.
var TokenTTL = DefaultTokenTTL

// ClaimsSpec describes the claims to embed in a signed token.
type ClaimsSpec struct {
	Issuer   string
	Subject  string
	Audience string
	Scope    []string
	TTL      time.Duration
}

// validate performs sanity checks on the claim specification before issuing a token.
func (s ClaimsSpec) validate() error {
	if strings.TrimSpace(s.Issuer) == "" {
		return errors.New("jwt: issuer required")
	}
	if strings.TrimSpace(s.Subject) == "" {
		return errors.New("jwt: subject required")
	}
	if strings.TrimSpace(s.Audience) == "" {
		return errors.New("jwt: audience required")
	}
	if len(s.Scope) == 0 {
		return errors.New("jwt: scope required")
	}
	return nil
}

// Issuer signs JWTs on behalf of the Portal using the manager's active key.
type Issuer struct {
	manager *Manager
	now     func() time.Time
}

// NewIssuer returns an Issuer bound to the provided Manager.
func NewIssuer(m *Manager) *Issuer {
	return &Issuer{manager: m, now: time.Now}
}

// Issue signs a JWT using the manager's active key according to spec.
func (i *Issuer) Issue(spec ClaimsSpec) (string, error) {
	if i == nil || i.manager == nil {
		return "", errors.New("jwt: issuer not initialized")
	}
	if err := spec.validate(); err != nil {
		return "", err
	}

	ttl := spec.TTL
	if ttl <= 0 {
		ttl = TokenTTL
	}
	if ttl > MaxTokenTTL {
		ttl = MaxTokenTTL
	}

	key, err := i.manager.EnsureActiveKey()
	if err != nil {
		return "", err
	}

	issuedAt := i.now().UTC()
	expiresAt := issuedAt.Add(ttl)

	claims := &Claims{
		Scope: strings.Join(spec.Scope, " "),
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    spec.Issuer,
			Subject:   spec.Subject,
			Audience:  gojwt.ClaimStrings{spec.Audience},
			IssuedAt:  gojwt.NewNumericDate(issuedAt),
			NotBefore: gojwt.NewNumericDate(issuedAt),
			ExpiresAt: gojwt.NewNumericDate(expiresAt),
			ID:        rnd.AuthTokenID(PrefixTokenID),
		},
	}

	token := gojwt.NewWithClaims(gojwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = key.Kid
	token.Header["typ"] = "JWT"

	signed, err := token.SignedString(key.PrivateKey)
	if err != nil {
		return "", err
	}
	return signed, nil
}
