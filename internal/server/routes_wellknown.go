package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/wellknown"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// registerWellknownRoutes adds "/.well-known/" service discovery routes.
func registerWellknownRoutes(router *gin.Engine, conf *config.Config) {
	// Registers the "/.well-known/oauth-authorization-server" service discovery endpoint for OAuth2 clients.
	router.Any(conf.BaseUri("/.well-known/oauth-authorization-server"), func(c *gin.Context) {
		c.JSON(http.StatusOK, wellknown.NewOAuthAuthorizationServer(conf))
	})

	// Registers the "/.well-known/openid-configuration" service discovery endpoint for OpenID Connect clients.
	router.Any(conf.BaseUri("/.well-known/openid-configuration"), func(c *gin.Context) {
		c.JSON(http.StatusOK, wellknown.NewOpenIDConfiguration(conf))
	})

	// Registers the "/.well-known/jwks.json" endpoint for cluster JWT verification.
	router.GET(conf.BaseUri("/.well-known/jwks.json"), func(c *gin.Context) {
		if !conf.Portal() {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		manager := get.JWTManager()

		if manager == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "jwks unavailable"})
			return
		}

		jwks := manager.JWKS()
		payload, err := json.Marshal(jwks)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "jwks marshal failed"})
			return
		}

		sum := sha256.Sum256(payload)
		etag := fmt.Sprintf("\"%x\"", sum[:8])
		ttl := conf.JWKSCacheTTL()

		if ttl <= 0 {
			ttl = 300
		}

		c.Header(header.CacheControl, fmt.Sprintf("max-age=%d, public", ttl))
		c.Header(header.ETag, etag)
		c.Data(http.StatusOK, header.ContentTypeJson, payload)
	})
}
