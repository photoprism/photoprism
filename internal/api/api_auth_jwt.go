package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	clusterjwt "github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
)

// authAnyJWT attempts to authenticate a Portal-issued JWT when a cluster node
// receives a request without an existing session. It verifies the token against
// the node's cached JWKS, ensures the issuer/audience/scope match the expected
// portal values, and, if valid, returns a client session mirroring the JWT
// claims. It returns nil on any validation failure so the caller can fall back
// to existing auth flows. By default, only cluster and vision resources are
// eligible, but nodes may opt in to additional scopes via PHOTOPRISM_JWT_SCOPE.
func authAnyJWT(c *gin.Context, clientIP, authToken string, resource acl.Resource, perms acl.Permissions) *entity.Session {
	// Check if token may be a JWT.
	if !shouldAttemptJWT(c, authToken) {
		return nil
	}

	conf := get.Config()

	// Determine whether JWT authentication is possible
	// based on the local config and client IP address.
	if !shouldAllowJWT(conf, clientIP) {
		return nil
	}

	requiredScope := resource.String()
	expected := expectedClaimsFor(conf, requiredScope)

	// verifyTokenFromPortal handles cryptographic validation (signature, issuer,
	// audience, temporal claims) and enforces that the token includes any scopes
	// listed in expected.Scope. Local authorization still happens below so nodes
	// can apply their own allow-list semantics.
	claims := verifyTokenFromPortal(c.Request.Context(), authToken, expected, jwtIssuerCandidates(conf))

	if claims == nil {
		return nil
	}

	// Check if config allows resource access to be authorized with JWT.
	allowedScopes := conf.JWTAllowedScopes()
	if !acl.ScopeAttrPermits(allowedScopes, resource, perms) {
		return nil
	}

	// Check if token allows access to specified resource.
	tokenScopes := acl.ScopeAttr(claims.Scope)
	if !acl.ScopeAttrPermits(tokenScopes, resource, perms) {
		return nil
	}

	claims.Scope = tokenScopes.String()

	return sessionFromJWTClaims(claims, clientIP)
}

// shouldAttemptJWT reports whether JWT verification should run for the supplied
// request context and token.
func shouldAttemptJWT(c *gin.Context, token string) bool {
	if c == nil {
		return false
	}

	if token == "" || strings.Count(token, ".") != 2 {
		return false
	}

	return true
}

// shouldAllowJWT reports whether the current node configuration permits JWT
// authentication for the request originating from clientIP.
func shouldAllowJWT(conf *config.Config, clientIP string) bool {
	if conf == nil || conf.IsPortal() {
		return false
	}

	if conf.JWKSUrl() == "" {
		return false
	}

	cidr := strings.TrimSpace(conf.ClusterCIDR())
	if cidr == "" {
		return true
	}

	ip := net.ParseIP(clientIP)
	_, block, err := net.ParseCIDR(cidr)
	if err != nil || ip == nil {
		return false
	}

	return block.Contains(ip)
}

// expectedClaimsFor builds the ExpectedClaims used to validate JWTs for the
// current node and required scope.
func expectedClaimsFor(conf *config.Config, requiredScope string) clusterjwt.ExpectedClaims {
	expected := clusterjwt.ExpectedClaims{
		Audience: fmt.Sprintf("node:%s", conf.NodeUUID()),
		JWKSURL:  conf.JWKSUrl(),
	}

	if requiredScope != "" {
		expected.Scope = []string{requiredScope}
	}

	return expected
}

// verifyTokenFromPortal checks the token against each candidate issuer and
// returns the verified claims on success.
func verifyTokenFromPortal(ctx context.Context, token string, expected clusterjwt.ExpectedClaims, issuers []string) *clusterjwt.Claims {
	if len(issuers) == 0 {
		return nil
	}

	for _, issuer := range issuers {
		expected.Issuer = issuer
		claims, err := get.VerifyJWT(ctx, token, expected)
		if err == nil {
			return claims
		}
	}

	return nil
}

// sessionFromJWTClaims constructs a Session populated with fields derived from
// the verified JWT claims.
func sessionFromJWTClaims(claims *clusterjwt.Claims, clientIP string) *entity.Session {
	sess := &entity.Session{
		Status:       http.StatusOK,
		ClientUID:    claims.Subject,
		AuthScope:    clean.Scope(claims.Scope),
		AuthIssuer:   claims.Issuer,
		AuthID:       claims.ID,
		GrantType:    authn.GrantJwtBearer.String(),
		AuthProvider: authn.ProviderClient.String(),
	}

	sess.SetMethod(authn.MethodJWT)
	sess.SetClientName(claims.Subject)
	sess.SetClientIP(clientIP)

	return sess
}

// jwtIssuerCandidates returns the possible issuer values the node should accept
// for Portal JWTs. It prefers the explicit portal cluster identifier and then
// falls back to configured URLs so legacy installations migrate seamlessly.
func jwtIssuerCandidates(conf *config.Config) []string {
	var out []string
	if uuid := conf.ClusterUUID(); uuid != "" {
		out = append(out, fmt.Sprintf("portal:%s", uuid))
	}
	if portal := strings.TrimSpace(conf.PortalUrl()); portal != "" {
		out = append(out, strings.TrimRight(portal, "/"))
	}
	if site := strings.TrimSpace(conf.SiteUrl()); site != "" {
		out = append(out, strings.TrimRight(site, "/"))
	}
	return out
}
