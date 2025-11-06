package entity

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// JWT captures the subset of JWT fields needed to construct
// an in-memory session for portal-to-node authentication flows.
type JWT struct {
	Token     string
	ID        string
	Issuer    string
	Subject   string
	Scope     string
	Audience  []string
	IssuedAt  *time.Time
	NotBefore *time.Time
	ExpiresAt *time.Time
}

// NewSessionFromJWT constructs an in-memory session based on verified
// portal JWT claims. The session is not persisted but mirrors the same
// bookkeeping as a regular session so downstream logging and ACL checks
// behave consistently.
func NewSessionFromJWT(c *gin.Context, jwt *JWT) *Session {
	if jwt == nil {
		return nil
	}

	// Create new session
	sess := &Session{
		Status: http.StatusOK,
		RefID:  jwt.ID,
	}

	// Determine token string.
	token := jwt.Token
	if token == "" {
		token = header.AuthToken(c)
	}
	sess.SetAuthToken(token)

	// Set scope/claims metadata.
	sess.SetScope(jwt.Scope)
	sess.SetAuthID(jwt.ID, jwt.Issuer)
	sess.SetGrantType(authn.GrantJwtBearer)
	sess.SetMethod(authn.MethodJWT)
	sess.SetProvider(authn.ProviderAccessToken)
	sess.SetClientName(jwt.Subject)
	sess.SetClientIP(header.ClientIP(c))
	sess.SetUserAgent(header.ClientUserAgent(c))

	// Derive timestamps from JWT claims when available.
	now := time.Now().UTC()
	issuedAt := now
	if jwt.IssuedAt != nil {
		issuedAt = jwt.IssuedAt.UTC()
	}
	notBefore := issuedAt
	if jwt.NotBefore != nil {
		nbf := jwt.NotBefore.UTC()
		if nbf.After(notBefore) {
			notBefore = nbf
		}
	}
	sess.CreatedAt = issuedAt
	sess.UpdatedAt = notBefore
	sess.LastActive = notBefore.Unix()

	// Apply expiration if provided; otherwise disable expiration.
	if jwt.ExpiresAt != nil {
		sess.SessExpires = jwt.ExpiresAt.UTC().Unix()
	} else {
		sess.SetExpiresIn(-1)
	}
	sess.SetTimeout(-1)

	return sess
}
