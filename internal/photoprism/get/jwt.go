package get

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/pkg/clean"
)

var (
	onceJWTManager  sync.Once
	onceJWTIssuer   sync.Once
	onceJWTVerifier sync.Once
)

// initJWTManager lazily initializes the shared portal key manager for JWT issuance.
func initJWTManager() {
	if conf == nil {
		return
	} else if !conf.Portal() {
		return
	}

	manager, err := jwt.NewManager(conf)

	if err != nil {
		log.Warnf("jwt: manager init failed (%s)", clean.Error(err))
		return
	}

	if _, err = manager.EnsureActiveKey(); err != nil {
		log.Warnf("jwt: ensure signing key failed (%s)", clean.Error(err))
	}

	services.JWTManager = manager
}

// JWTManager returns the portal key manager; nil on nodes.
func JWTManager() *jwt.Manager {
	onceJWTManager.Do(initJWTManager)
	return services.JWTManager
}

// initJWTIssuer lazily binds the shared issuer to the active portal key manager.
func initJWTIssuer() {
	manager := JWTManager()
	if manager == nil {
		return
	}
	services.JWTIssuer = jwt.NewIssuer(manager)
}

// JWTIssuer returns the portal JWT issuer helper; nil on nodes.
func JWTIssuer() *jwt.Issuer {
	onceJWTIssuer.Do(initJWTIssuer)
	return services.JWTIssuer
}

// JWTVerifier returns a verifier bound to the current config.
func JWTVerifier() *jwt.Verifier {
	onceJWTVerifier.Do(initJWTVerifier)
	return services.JWTVerifier
}

// VerifyJWT verifies a token using the shared verifier instance.
func VerifyJWT(ctx context.Context, token string, expected jwt.ExpectedClaims) (*jwt.Claims, error) {
	verifier := JWTVerifier()
	if verifier == nil {
		return nil, errors.New("jwt: verifier not available")
	}
	return verifier.VerifyToken(ctx, token, expected)
}

// initJWTVerifier lazily constructs the shared verifier for the current configuration.
func initJWTVerifier() {
	if conf != nil {
		services.JWTVerifier = jwt.NewVerifier(conf)
	}
}

// resetJWTVerifier clears the cached verifier so it can be rebuilt for a new configuration.
func resetJWTVerifier() {
	services.JWTVerifier = nil
	onceJWTVerifier = sync.Once{}
}

// resetJWTIssuer clears the cached issuer so it can be recreated for a new configuration.
func resetJWTIssuer() {
	services.JWTIssuer = nil
	onceJWTIssuer = sync.Once{}
}

// resetJWTManager clears the cached key manager so subsequent calls reload keys for the active configuration.
func resetJWTManager() {
	services.JWTManager = nil
	onceJWTManager = sync.Once{}
}

// resetJWT clears all cached JWT helpers.
func resetJWT() {
	resetJWTVerifier()
	resetJWTIssuer()
	resetJWTManager()
}

// IssuePortalJWT signs a token using the shared portal issuer with the provided claims.
func IssuePortalJWT(spec jwt.ClaimsSpec) (string, error) {
	if issuer := JWTIssuer(); issuer == nil {
		return "", errors.New("jwt: issuer not available")
	} else {
		return issuer.Issue(spec)
	}
}

// IssuePortalJWTForNode issues a portal-signed JWT targeting the specified node UUID.
func IssuePortalJWTForNode(nodeUUID string, scopes []string, ttl time.Duration) (string, error) {
	if conf == nil {
		return "", errors.New("jwt: missing config")
	} else if !conf.Portal() {
		return "", errors.New("jwt: not supported on nodes")
	}

	clusterUUID := strings.TrimSpace(conf.ClusterUUID())
	if clusterUUID == "" {
		return "", errors.New("jwt: cluster uuid not configured")
	}

	nodeUUID = strings.TrimSpace(nodeUUID)
	if nodeUUID == "" {
		return "", errors.New("jwt: node uuid required")
	}
	if len(scopes) == 0 {
		return "", errors.New("jwt: at least one scope is required")
	}

	normalized := make([]string, 0, len(scopes))
	for _, s := range scopes {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	if len(normalized) == 0 {
		return "", errors.New("jwt: at least one scope is required")
	}

	spec := jwt.ClaimsSpec{
		Issuer:   fmt.Sprintf("portal:%s", clusterUUID),
		Subject:  fmt.Sprintf("portal:%s", clusterUUID),
		Audience: fmt.Sprintf("node:%s", nodeUUID),
		Scope:    normalized,
		TTL:      ttl,
	}

	return IssuePortalJWT(spec)
}
