package api

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
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

// RegisterRequireClientSecret controls whether registrations that reference an
// existing ClientID must also present the matching client secret. Enabled by default.
var RegisterRequireClientSecret = true

// ClusterNodesRegister registers the Portal-only node registration endpoint.
//
//	@Summary	registers a node, provisions DB credentials, and issues clientSecret
//	@Id			ClusterNodesRegister
//	@Tags		Cluster
//	@Accept		json
//	@Produce	json
//	@Param		request				body		object	true	"registration payload (nodeName required; optional: nodeRole, labels, advertiseUrl, siteUrl; to authorize UUID/name changes include clientId+clientSecret; rotation: rotateDatabase, rotateSecret)"
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
			NodeUUID       string            `json:"nodeUUID"`
			NodeRole       string            `json:"nodeRole"`
			Labels         map[string]string `json:"labels"`
			AdvertiseUrl   string            `json:"advertiseUrl"`
			SiteUrl        string            `json:"siteUrl"`
			ClientID       string            `json:"clientId"`
			ClientSecret   string            `json:"clientSecret"`
			RotateDatabase bool              `json:"rotateDatabase"`
			RotateSecret   bool              `json:"rotateSecret"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "form invalid", "%s"}, clean.Error(err))
			AbortBadRequest(c)
			return
		}

		// If an existing ClientID is provided, require the corresponding client secret for verification.
		if RegisterRequireClientSecret && req.ClientID != "" {
			if !rnd.IsUID(req.ClientID, entity.ClientUID) {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "invalid client id"})
				AbortBadRequest(c)
				return
			}
			pw := entity.FindPassword(req.ClientID)
			if pw == nil || req.ClientSecret == "" || !pw.Valid(req.ClientSecret) {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "invalid client secret"})
				AbortUnauthorized(c)
				return
			}
		}

		name := clean.DNSLabel(req.NodeName)

		// Enforce DNS label semantics for node names: lowercase [a-z0-9-], 1–32, start/end alnum.
		if name == "" || len(name) > 32 || name[0] == '-' || name[len(name)-1] == '-' {
			event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "invalid name"})
			AbortBadRequest(c)
			return
		}
		for i := 0; i < len(name); i++ {
			b := name[i]
			if !(b == '-' || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')) {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "invalid name chars"})
				AbortBadRequest(c)
				return
			}
		}

		// Validate advertise URL if provided (https required for non-local domains).
		if u := strings.TrimSpace(req.AdvertiseUrl); u != "" {
			if !validateAdvertiseURL(u) {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "invalid advertise url"})
				AbortBadRequest(c)
				return
			}
		}

		// Validate site URL if provided (https required for non-local domains).
		if su := strings.TrimSpace(req.SiteUrl); su != "" {
			if !validateSiteURL(su) {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "invalid site url"})
				AbortBadRequest(c)
				return
			}
		}

		// Sanitize requested NodeUUID; generation happens later depending on path (existing vs new).
		requestedUUID := rnd.SanitizeUUID(req.NodeUUID)

		// Registry (client-backed).
		regy, err := reg.NewClientRegistryWithConfig(conf)

		if err != nil {
			event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "registry", event.Failed, "%s"}, clean.Error(err))
			AbortUnexpectedError(c)
			return
		}

		// Try to find existing node.
		if n, _ := regy.FindByName(name); n != nil {
			// If caller attempts to change UUID by name without proving client secret, block with 409.
			if RegisterRequireClientSecret {
				if requestedUUID != "" && n.UUID != "" && requestedUUID != n.UUID && req.ClientID == "" {
					event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "uuid change requires client secret", event.Denied, "name %s", clean.LogQuote(name)})
					c.JSON(http.StatusConflict, gin.H{"error": "client secret required to change node uuid"})
					return
				}
			}
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
			// Apply UUID changes for existing node: if a UUID was requested and differs, or if none exists yet.
			if requestedUUID != "" {
				oldUUID := n.UUID
				if oldUUID != requestedUUID {
					n.UUID = requestedUUID
					// Emit audit event for UUID change.
					event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "uuid changed", event.Succeeded, "name %s", "old %s", "new %s"}, clean.LogQuote(name), clean.Log(oldUUID), clean.Log(requestedUUID))
				}
			} else if n.UUID == "" {
				// Assign a fresh UUID if missing and none requested.
				n.UUID = rnd.UUIDv7()
				event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "uuid changed", event.Succeeded, "name %s", "old %s", "new %s"}, clean.LogQuote(name), clean.Log(""), clean.Log(n.UUID))
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
				if n, err = regy.RotateSecret(n.UUID); err != nil {
					event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "rotate secret", event.Failed, "%s"}, clean.Error(err))
					AbortUnexpectedError(c)
					return
				}
				respSecret = &cluster.RegisterSecrets{ClientSecret: n.ClientSecret, RotatedAt: n.RotatedAt}
				event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "rotate secret", event.Succeeded, "node %s"}, clean.LogQuote(name))

				// Extra safety: ensure the updated secret is persisted even if subsequent steps fail.
				if putErr := regy.Put(n); putErr != nil {
					event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "persist rotated secret", event.Failed, "%s"}, clean.Error(putErr))
					AbortUnexpectedError(c)
					return
				}
			}

			// Ensure that a database for this node exists (rotation optional).
			creds, _, credsErr := provisioner.GetCredentials(c, conf, n.UUID, name, req.RotateDatabase)

			if credsErr != nil {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "ensure database", event.Failed, "%s"}, clean.Error(credsErr))
				c.JSON(http.StatusConflict, gin.H{"error": credsErr.Error()})
				return
			}

			if req.RotateDatabase {
				n.Database.RotatedAt = creds.RotatedAt
				n.Database.Driver = provisioner.DatabaseDriver
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
				UUID:               conf.ClusterUUID(),
				Node:               reg.BuildClusterNode(*n, opts),
				Database:           cluster.RegisterDatabase{Host: conf.DatabaseHost(), Port: conf.DatabasePort(), Name: n.Database.Name, User: n.Database.User, Driver: provisioner.DatabaseDriver},
				Secrets:            respSecret,
				AlreadyRegistered:  true,
				AlreadyProvisioned: true,
			}

			// Include password/dsn only if rotated now.
			if req.RotateDatabase {
				resp.Database.Password = creds.Password
				resp.Database.DSN = creds.DSN
				resp.Database.RotatedAt = creds.RotatedAt
			}

			c.Header(header.CacheControl, header.CacheControlNoStore)
			c.JSON(http.StatusOK, resp)
			return
		}

		// New node (client UID will be generated in registry.Put).
		n := &reg.Node{
			Name:   name,
			Role:   clean.TypeLowerDash(req.NodeRole),
			UUID:   requestedUUID,
			Labels: req.Labels,
		}
		if n.UUID == "" {
			n.UUID = rnd.UUIDv7()
		}
		// Derive a sensible default advertise URL when not provided by the client.
		if req.AdvertiseUrl != "" {
			n.AdvertiseUrl = req.AdvertiseUrl
		} else if d := conf.ClusterDomain(); d != "" {
			n.AdvertiseUrl = "https://" + name + "." + d
		}
		if s := normalizeSiteURL(req.SiteUrl); s != "" {
			n.SiteUrl = s
		}

		// Generate node secret (must satisfy client secret format for entity.Client).
		n.ClientSecret = rnd.ClientSecret()
		n.RotatedAt = nowRFC3339()

		// Ensure DB (force rotation at create path to return password).
		creds, _, err := provisioner.GetCredentials(c, conf, n.UUID, name, true)
		if err != nil {
			event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "ensure database", event.Failed, "%s"}, clean.Error(err))
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		n.Database.Name, n.Database.User, n.Database.RotatedAt = creds.Name, creds.User, creds.RotatedAt
		n.Database.Driver = provisioner.DatabaseDriver

		if err = regy.Put(n); err != nil {
			event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "nodes", "register", "persist node", event.Failed, "%s"}, clean.Error(err))
			AbortUnexpectedError(c)
			return
		}

		resp := cluster.RegisterResponse{
			Node:               reg.BuildClusterNode(*n, reg.NodeOptsForSession(nil)),
			Secrets:            &cluster.RegisterSecrets{ClientSecret: n.ClientSecret, RotatedAt: n.RotatedAt},
			Database:           cluster.RegisterDatabase{Host: conf.DatabaseHost(), Port: conf.DatabasePort(), Name: creds.Name, User: creds.User, Driver: provisioner.DatabaseDriver, Password: creds.Password, DSN: creds.DSN, RotatedAt: creds.RotatedAt},
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

// validateAdvertiseURL checks that the URL is absolute with a host and scheme,
// and requires https for non-local hosts. http is allowed only for localhost/127.0.0.1/::1.
func validateAdvertiseURL(u string) bool {
	parsed, err := url.Parse(strings.TrimSpace(u))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if parsed.Scheme == "https" {
		return true
	}
	if parsed.Scheme == "http" {
		if host == "localhost" || host == "127.0.0.1" || host == "::1" {
			return true
		}
		return false
	}
	return false
}

// validateSiteURL applies the same rules as validateAdvertiseURL.
func validateSiteURL(u string) bool { return validateAdvertiseURL(u) }
