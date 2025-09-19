package api

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/internal/service/cluster/provisioner"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/service/http/header"
)

// ClusterNodesRegister registers the Portal-only node registration endpoint.
//
//	@Summary	registers a node, provisions DB credentials, and issues nodeSecret
//	@Id			ClusterNodesRegister
//	@Tags		Cluster
//	@Accept		json
//	@Produce	json
//	@Param		request				body		object	true	"registration payload (nodeName required; optional: nodeRole, labels, advertiseUrl, siteUrl, rotateDatabase, rotateSecret)"
//	@Success	200,201				{object}	cluster.RegisterResponse
//	@Failure	400,401,403,409,429	{object}	i18n.Response
//	@Router		/api/v1/cluster/nodes/register [post]
func ClusterNodesRegister(router *gin.RouterGroup) {
	router.POST("/cluster/nodes/register", func(c *gin.Context) {
		conf := get.Config()

		// Must be a portal.
		if !conf.IsPortal() {
			AbortFeatureDisabled(c)
			return
		}

		// Rate limit by IP (reuse existing limiter).
		clientIp := ClientIP(c)
		r := limiter.Auth.Request(clientIp)

		if r.Reject() || limiter.Auth.Reject(clientIp) {
			event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "rate limit", event.Denied})
			limiter.AbortJSON(c)
			return
		}

		// Token check (Bearer).
		expected := conf.JoinToken()
		token := header.BearerToken(c)

		if expected == "" || token == "" || subtle.ConstantTimeCompare([]byte(expected), []byte(token)) != 1 {
			event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "auth", event.Denied})
			r.Success() // return reserved tokens; still unauthorized
			AbortUnauthorized(c)
			return
		}

		// Parse request.
		var req struct {
			NodeName       string            `json:"nodeName"`
			NodeRole       string            `json:"nodeRole"`
			Labels         map[string]string `json:"labels"`
			AdvertiseUrl   string            `json:"advertiseUrl"`
			SiteUrl        string            `json:"siteUrl"`
			RotateDatabase bool              `json:"rotateDatabase"`
			RotateSecret   bool              `json:"rotateSecret"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "form invalid", "%s"}, clean.Error(err))
			AbortBadRequest(c)
			return
		}

		name := clean.TypeLowerDash(req.NodeName)

		if name == "" || len(name) < 1 || len(name) > 63 {
			event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "invalid name"})
			AbortBadRequest(c)
			return
		}

		// Registry (client-backed).
		regy, err := reg.NewClientRegistryWithConfig(conf)

		if err != nil {
			event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "registry", event.Failed, "%s"}, clean.Error(err))
			AbortUnexpectedError(c)
			return
		}

		// Try to find existing node.
		if n, _ := regy.FindByName(name); n != nil {
			// Update mutable metadata when provided.
			if req.AdvertiseUrl != "" {
				n.AdvertiseUrl = req.AdvertiseUrl
			}
			if req.Labels != nil {
				n.Labels = req.Labels
			}
			if s := normalizeSiteURL(req.SiteUrl); s != "" {
				n.SiteUrl = s
			}
			// Persist metadata changes so UpdatedAt advances.
			if putErr := regy.Put(n); putErr != nil {
				event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "persist node", event.Failed, "%s"}, clean.Error(putErr))
				AbortUnexpectedError(c)
				return
			}
			// Optional rotations.
			var respSecret *cluster.RegisterSecrets
			if req.RotateSecret {
				if n, err = regy.RotateSecret(n.ID); err != nil {
					event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "rotate secret", event.Failed, "%s"}, clean.Error(err))
					AbortUnexpectedError(c)
					return
				}
				respSecret = &cluster.RegisterSecrets{NodeSecret: n.Secret, SecretRotatedAt: n.SecretRot}
				event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "rotate secret", event.Succeeded, "node %s"}, clean.LogQuote(name))

				// Extra safety: ensure the updated secret is persisted even if subsequent steps fail.
				if putErr := regy.Put(n); putErr != nil {
					event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "persist rotated secret", event.Failed, "%s"}, clean.Error(putErr))
					AbortUnexpectedError(c)
					return
				}
			}

			// Ensure that a database for this node exists (rotation optional).
			creds, _, credsErr := provisioner.EnsureNodeDatabase(c, conf, name, req.RotateDatabase)

			if credsErr != nil {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "ensure database", event.Failed, "%s"}, clean.Error(credsErr))
				c.JSON(http.StatusConflict, gin.H{"error": credsErr.Error()})
				return
			}

			if req.RotateDatabase {
				n.DB.RotAt = creds.LastRotatedAt
				if putErr := regy.Put(n); putErr != nil {
					event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "persist node", event.Failed, "%s"}, clean.Error(putErr))
					AbortUnexpectedError(c)
					return
				}
				event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "rotate db", event.Succeeded, "node %s"}, clean.LogQuote(name))
			}

			// Build response with struct types.
			opts := reg.NodeOptsForSession(nil) // registration is token-based, not session; default redaction is fine
			resp := cluster.RegisterResponse{
				Node:               reg.BuildClusterNode(*n, opts),
				Database:           cluster.RegisterDatabase{Host: conf.DatabaseHost(), Port: conf.DatabasePort(), Name: n.DB.Name, User: n.DB.User},
				Secrets:            respSecret,
				AlreadyRegistered:  true,
				AlreadyProvisioned: true,
			}

			// Include password/dsn only if rotated now.
			if req.RotateDatabase {
				resp.Database.Password = creds.Password
				resp.Database.DSN = creds.DSN
				resp.Database.RotatedAt = creds.LastRotatedAt
			}

			c.Header(header.CacheControl, header.CacheControlNoStore)
			c.JSON(http.StatusOK, resp)
			return
		}

		// New node.
		n := &reg.Node{
			ID:           rnd.UUID(),
			Name:         name,
			Role:         clean.TypeLowerDash(req.NodeRole),
			Labels:       req.Labels,
			AdvertiseUrl: req.AdvertiseUrl,
		}
		if s := normalizeSiteURL(req.SiteUrl); s != "" {
			n.SiteUrl = s
		}

		// Generate node secret.
		n.Secret = rnd.Base62(48)
		n.SecretRot = nowRFC3339()

		// Ensure DB (force rotation at create path to return password).
		creds, _, err := provisioner.EnsureNodeDatabase(c, conf, name, true)
		if err != nil {
			event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "ensure database", event.Failed, "%s"}, clean.Error(err))
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		n.DB.Name, n.DB.User, n.DB.RotAt = creds.Name, creds.User, creds.LastRotatedAt

		if err = regy.Put(n); err != nil {
			event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "persist node", event.Failed, "%s"}, clean.Error(err))
			AbortUnexpectedError(c)
			return
		}

		resp := cluster.RegisterResponse{
			Node:               reg.BuildClusterNode(*n, reg.NodeOptsForSession(nil)),
			Secrets:            &cluster.RegisterSecrets{NodeSecret: n.Secret, SecretRotatedAt: n.SecretRot},
			Database:           cluster.RegisterDatabase{Host: conf.DatabaseHost(), Port: conf.DatabasePort(), Name: creds.Name, User: creds.User, Password: creds.Password, DSN: creds.DSN, RotatedAt: creds.LastRotatedAt},
			AlreadyRegistered:  false,
			AlreadyProvisioned: false,
		}

		c.Header(header.CacheControl, header.CacheControlNoStore)
		event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", event.Created, event.Succeeded, "node %s"}, clean.LogQuote(name))
		c.JSON(http.StatusCreated, resp)
	})
}

// normalizeSiteURL validates and normalizes a site URL for storage.
// Rules: require http/https scheme, non-empty host, <=255 chars; lowercase host.
func normalizeSiteURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	if len(u) > 255 {
		return ""
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	if parsed.Host == "" {
		return ""
	}
	parsed.Host = strings.ToLower(parsed.Host)
	return parsed.String()
}
