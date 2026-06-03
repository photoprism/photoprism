package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// OAuth2 authorization-code lifetimes. Codes are short-lived per RFC 6749
// §4.1.2 (the client redeems them immediately at the token endpoint); the
// default leaves room for a native-app redirect handoff without weakening the
// single-use guarantee.
const (
	DefaultOAuthCodeTTL = 300 * time.Second
	MaxOAuthCodeTTL     = 600 * time.Second
)

// OAuthCode is a single-use OAuth2 authorization code minted by the authorize
// endpoint and redeemed at the token endpoint. The raw value is returned to the
// client once and never stored; only its SHA-256 hash is persisted so a
// database leak cannot replay outstanding grants.
type OAuthCode struct {
	ID                  uint      `gorm:"primary_key;" json:"ID" yaml:"ID"`
	CodeHash            string    `gorm:"type:VARBINARY(64);unique_index;default:'';" json:"-" yaml:"-"`
	ClientUID           string    `gorm:"type:VARBINARY(42);index;default:'';" json:"ClientUID" yaml:"ClientUID"`
	UserUID             string    `gorm:"type:VARBINARY(42);index;default:'';" json:"UserUID" yaml:"UserUID"`
	RedirectURI         string    `gorm:"type:VARBINARY(2048);default:'';" json:"RedirectURI" yaml:"RedirectURI"`
	Scope               string    `gorm:"size:1024;default:'';" json:"Scope" yaml:"Scope"`
	CodeChallenge       string    `gorm:"size:128;default:'';" json:"CodeChallenge,omitempty" yaml:"CodeChallenge,omitempty"`
	CodeChallengeMethod string    `gorm:"size:16;default:'';" json:"CodeChallengeMethod,omitempty" yaml:"CodeChallengeMethod,omitempty"`
	ExpiresAt           time.Time `gorm:"index" json:"ExpiresAt" yaml:"ExpiresAt"`
	CreatedAt           time.Time `json:"CreatedAt" yaml:"-"`
}

// TableName returns the entity table name.
func (OAuthCode) TableName() string {
	return "oauth_codes"
}

// OAuthCodeSpec carries the fields pinned at authorize time so the matching
// token call can redeem the code. TTL falls back to DefaultOAuthCodeTTL and
// clamps at MaxOAuthCodeTTL.
type OAuthCodeSpec struct {
	ClientUID           string
	UserUID             string
	RedirectURI         string
	Scope               string
	CodeChallenge       string
	CodeChallengeMethod string
	TTL                 time.Duration
}

// validate performs sanity checks on the spec before inserting the row.
func (s OAuthCodeSpec) validate() error {
	if strings.TrimSpace(s.ClientUID) == "" {
		return errors.New("oauth: client uid required")
	}
	if strings.TrimSpace(s.UserUID) == "" {
		return errors.New("oauth: user uid required")
	}
	if strings.TrimSpace(s.RedirectURI) == "" {
		return errors.New("oauth: redirect uri required")
	}
	return nil
}

// HashOAuthCode returns the lowercase-hex SHA-256 of the raw authorization
// code as it is stored in oauth_codes.code_hash.
func HashOAuthCode(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// NewOAuthCode inserts a single-use authorization code matching spec and
// returns both the raw code (which the caller redirects back to the client)
// and the persisted row. The raw code is never stored.
func NewOAuthCode(spec OAuthCodeSpec) (raw string, m *OAuthCode, err error) {
	if err = spec.validate(); err != nil {
		return "", nil, err
	}

	ttl := spec.TTL
	if ttl <= 0 {
		ttl = DefaultOAuthCodeTTL
	}
	if ttl > MaxOAuthCodeTTL {
		ttl = MaxOAuthCodeTTL
	}

	now := time.Now().UTC()
	raw = rnd.AuthToken()

	m = &OAuthCode{
		CodeHash:            HashOAuthCode(raw),
		ClientUID:           spec.ClientUID,
		UserUID:             spec.UserUID,
		RedirectURI:         spec.RedirectURI,
		Scope:               spec.Scope,
		CodeChallenge:       spec.CodeChallenge,
		CodeChallengeMethod: spec.CodeChallengeMethod,
		ExpiresAt:           now.Add(ttl),
	}

	if err = Db().Create(m).Error; err != nil {
		return "", nil, err
	}
	return raw, m, nil
}

// FindOAuthCode looks up the authorization-code row matching raw. It returns
// (nil, nil) when no row exists so callers can map the miss to the OAuth
// invalid_grant response without distinguishing replays from typos.
func FindOAuthCode(raw string) (*OAuthCode, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	m := &OAuthCode{}
	if err := Db().Where("code_hash = ?", HashOAuthCode(raw)).First(m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}

// IsExpired reports whether the code's ExpiresAt has passed.
func (m *OAuthCode) IsExpired() bool {
	if m == nil {
		return true
	}
	return time.Now().UTC().After(m.ExpiresAt)
}

// Delete removes the auth-code row from the database. Single-use semantics: the
// token endpoint deletes the code on first successful redemption, and any
// subsequent FindOAuthCode for the same raw value returns (nil, nil).
func (m *OAuthCode) Delete() error {
	if m == nil || m.ID == 0 {
		return fmt.Errorf("oauth: auth code id is zero")
	}
	return Db().Delete(m).Error
}

// PurgeExpiredOAuthCodes hard-deletes every auth-code row whose ExpiresAt is in
// the past. Safe to call repeatedly; returns the number of rows removed.
func PurgeExpiredOAuthCodes() (int64, error) {
	tx := Db().Where("expires_at < ?", time.Now().UTC()).Delete(&OAuthCode{})
	return tx.RowsAffected, tx.Error
}
