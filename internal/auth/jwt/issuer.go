package jwt

import (
	"errors"
	"strings"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// PrefixTokenID prefixes JWT token IDs to make them distinguishable.
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

// UserClaimsSpec describes the claims to embed in a Portal OIDC ID/access
// token. Callers populate only the OIDC profile fields permitted by the
// requested OAuth scopes; the issuer does not enforce scope-claim coupling.
type UserClaimsSpec struct {
	Issuer   string
	Subject  string
	Audience string
	TTL      time.Duration

	NodeUUID string
	AuthTime time.Time
	Nonce    string
	Role     string

	Email             string
	EmailVerified     *bool
	Name              string
	GivenName         string
	FamilyName        string
	PreferredUsername string
	Locale            string
	Zoneinfo          string
	Picture           string

	Groups []string
}

// validate performs sanity checks on the user-claim specification before issuing a token.
func (s UserClaimsSpec) validate() error {
	if strings.TrimSpace(s.Issuer) == "" {
		return errors.New("jwt: issuer required")
	}
	if strings.TrimSpace(s.Subject) == "" {
		return errors.New("jwt: subject required")
	}
	if strings.TrimSpace(s.Audience) == "" {
		return errors.New("jwt: audience required")
	}
	if strings.TrimSpace(s.NodeUUID) == "" {
		return errors.New("jwt: node uuid required")
	}
	return nil
}

// UserClaims is the claim set emitted in Portal OIDC ID and access tokens.
// Field tags use the OIDC-standard claim names so instances can consume the
// token through their existing OIDC RP code without translation.
type UserClaims struct {
	Nonce             string   `json:"nonce,omitempty"`
	AuthTime          int64    `json:"auth_time,omitempty"`
	Email             string   `json:"email,omitempty"`
	EmailVerified     *bool    `json:"email_verified,omitempty"`
	Name              string   `json:"name,omitempty"`
	GivenName         string   `json:"given_name,omitempty"`
	FamilyName        string   `json:"family_name,omitempty"`
	PreferredUsername string   `json:"preferred_username,omitempty"`
	Locale            string   `json:"locale,omitempty"`
	Zoneinfo          string   `json:"zoneinfo,omitempty"`
	Picture           string   `json:"picture,omitempty"`
	Groups            []string `json:"groups,omitempty"`
	PortalRole        string   `json:"pp_role,omitempty"`
	PortalNodeUUID    string   `json:"pp_node_uuid,omitempty"`
	PortalIssuerKind  string   `json:"pp_issuer_kind,omitempty"`
	gojwt.RegisteredClaims
}

// IssueUser signs a Portal OIDC ID/access token using the manager's active
// key. The token shape mirrors what an instance OIDC RP would receive from
// any upstream IdP, with three Portal-specific claims (pp_role, pp_node_uuid,
// pp_issuer_kind) added so instances can correlate the token with the node
// they were targeted for.
func (i *Issuer) IssueUser(spec UserClaimsSpec) (string, error) {
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

	claims := &UserClaims{
		Nonce:             spec.Nonce,
		Email:             spec.Email,
		EmailVerified:     spec.EmailVerified,
		Name:              spec.Name,
		GivenName:         spec.GivenName,
		FamilyName:        spec.FamilyName,
		PreferredUsername: spec.PreferredUsername,
		Locale:            spec.Locale,
		Zoneinfo:          spec.Zoneinfo,
		Picture:           spec.Picture,
		Groups:            spec.Groups,
		PortalRole:        spec.Role,
		PortalNodeUUID:    spec.NodeUUID,
		PortalIssuerKind:  acl.RolePortal.String(),
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
	if !spec.AuthTime.IsZero() {
		claims.AuthTime = spec.AuthTime.UTC().Unix()
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
