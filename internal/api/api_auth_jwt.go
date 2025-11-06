package api

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/auth/acl"
	clusterjwt "github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
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

	if conf == nil {
		if log.IsLevelEnabled(logrus.DebugLevel) {
			log.Debug("auth: skipping portal jwt (config unavailable)")
		}
		return nil
	}

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
		if log.IsLevelEnabled(logrus.DebugLevel) {
			log.Debugf(
				"auth: portal jwt rejected for resource %s (client=%s required_scope=%q perms=%s)",
				resource,
				clean.IP(clientIP, "?"),
				strings.Join(expected.Scope, " "),
				perms.String(),
			)
		}
		return nil
	}

	// Check if config allows resource access to be authorized with JWT.
	allowedScopes := conf.JWTAllowedScopes()
	if !acl.ScopeAttrPermits(allowedScopes, resource, perms) {
		if log.IsLevelEnabled(logrus.DebugLevel) {
			log.Debugf("auth: portal jwt scope blocked by node allow-list (allowed=%q resource=%s perms=%s)", allowedScopes.String(), resource, perms.String())
		}
		return nil
	}

	// Check if token allows access to specified resource.
	tokenScopes := acl.ScopeAttr(claims.Scope)
	if !acl.ScopeAttrPermits(tokenScopes, resource, perms) {
		if log.IsLevelEnabled(logrus.DebugLevel) {
			log.Debugf("auth: portal jwt missing required scope (token=%q resource=%s perms=%s)", clean.Scope(claims.Scope), resource, perms.String())
		}
		return nil
	}

	claims.Scope = tokenScopes.String()

	var issuedAt, notBefore, expiresAt *time.Time
	if claims.IssuedAt != nil {
		t := claims.IssuedAt.Time
		issuedAt = &t
	}
	if claims.NotBefore != nil {
		t := claims.NotBefore.Time
		notBefore = &t
	}
	if claims.ExpiresAt != nil {
		t := claims.ExpiresAt.Time
		expiresAt = &t
	}

	return entity.NewSessionFromJWT(c, &entity.JWT{
		Token:     authToken,
		ID:        claims.ID,
		Issuer:    claims.Issuer,
		Subject:   claims.Subject,
		Scope:     claims.Scope,
		Audience:  claims.Audience,
		IssuedAt:  issuedAt,
		NotBefore: notBefore,
		ExpiresAt: expiresAt,
	})
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
	if conf == nil || conf.Portal() {
		if log.IsLevelEnabled(logrus.DebugLevel) {
			log.Debug("auth: skipping portal jwt (not a node)")
		}
		return false
	}

	if conf.JWKSUrl() == "" {
		if log.IsLevelEnabled(logrus.DebugLevel) {
			log.Debug("auth: skipping portal jwt (jwks url not configured)")
		}
		return false
	}

	cidr := strings.TrimSpace(conf.ClusterCIDR())
	if cidr == "" {
		return true
	}

	ip := net.ParseIP(clientIP)
	_, block, err := net.ParseCIDR(cidr)
	if err != nil || ip == nil {
		if log.IsLevelEnabled(logrus.DebugLevel) {
			log.Debugf("auth: skipping portal jwt (invalid cidr %q or client ip %q)", clean.Log(cidr), clean.Log(clientIP))
		}
		return false
	}

	if !block.Contains(ip) {
		if log.IsLevelEnabled(logrus.DebugLevel) {
			log.Debugf("auth: skipping portal jwt (client ip %q outside allowed cidr %q)", clean.Log(clientIP), clean.Log(cidr))
		}
		return false
	}

	return true
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
		if log.IsLevelEnabled(logrus.DebugLevel) {
			log.Debug("auth: portal jwt verification skipped (no issuer candidates)")
		}
		return nil
	}

	var lastErr error

	for _, issuer := range issuers {
		expected.Issuer = issuer
		claims, err := get.VerifyJWT(ctx, token, expected)
		if err == nil {
			return claims
		}
		lastErr = err
		if log.IsLevelEnabled(logrus.DebugLevel) {
			log.Debugf("auth: portal jwt issuer candidate %s rejected (%s)", clean.Log(issuer), clean.Error(err))
		}
	}

	if lastErr != nil && log.IsLevelEnabled(logrus.DebugLevel) {
		log.Debugf("auth: portal jwt verification failed after %d issuer attempts (%s)", len(issuers), clean.Error(lastErr))
	}

	return nil
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
