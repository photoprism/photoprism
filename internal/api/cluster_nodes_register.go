package api

import (
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/internal/service/cluster/provisioner"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/internal/service/cluster/theme"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// ClusterNodesRegister registers the Portal-only node registration endpoint.
//
//	@Summary	registers a node, provisions DB credentials, and issues ClientSecret
//	@Id			ClusterNodesRegister
//	@Tags		Cluster
//	@Accept		json
//	@Produce	json
//	@Param		request				body		object	true	"registration payload (NodeName required; optional: NodeRole, Labels, AdvertiseUrl, SiteUrl, AppName, AppVersion, Theme, NodeUUID, RotateDatabase, RotateSecret). New-node joins require the Bearer join token. Existing-node mutations require a Bearer OAuth access token that belongs to the same node client."
//	@Success	200,201				{object}	cluster.RegisterResponse
//	@Failure	400,401,403,409,429	{object}	i18n.Response
//	@Router		/api/v1/cluster/nodes/register [post]
func ClusterNodesRegister(router *gin.RouterGroup) {
	router.POST("/cluster/nodes/register", func(c *gin.Context) {
		// Prevent CDNs from caching this endpoint.
		if header.IsCdn(c.Request) {
			AbortNotFound(c)
			return
		}

		conf := get.Config()

		// Must be a portal.
		if !conf.Portal() {
			AbortFeatureDisabled(c)
			return
		}

		// Don't cache requests to this endpoint.
		c.Header(header.CacheControl, header.CacheControlNoStore)

		// Rate limit by IP (reuse existing limiter).
		clientIp := ClientIP(c)
		r := limiter.Auth.Request(clientIp)

		if r.Reject() || limiter.Auth.Reject(clientIp) {
			event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", status.RateLimited})
			limiter.AbortJSON(c)
			return
		}

		// Optional IP-based allowance via ClusterCIDR.
		if cidr := strings.TrimSpace(conf.ClusterCIDR()); cidr != "" {
			if !clusterCIDRAllowsClientIP(cidr, clientIp) {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", "client ip outside cluster-cidr", status.Denied})
				r.Success() // Return reserved tokens before aborting.
				AbortUnauthorized(c)
				return
			}
		}

		// Parse request.
		var req cluster.RegisterRequest

		LimitRequestBodyBytes(c, MaxClusterRegisterBytes)

		if err := c.ShouldBindJSON(&req); err != nil {
			if IsRequestBodyTooLarge(err) {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", "request too large", status.Error(err)})
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}

			event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", "invalid form", status.Error(err)})
			AbortBadRequest(c)
			return
		}

		appName := clean.TypeUnicode(req.AppName)
		appVersion := clean.TypeUnicode(req.AppVersion)
		nodeTheme := clean.TypeUnicode(req.Theme)

		name := clean.DNSLabel(req.NodeName)

		// Enforce DNS label semantics for node names: lowercase [a-z0-9-], 1–32, start/end alnum.
		if name == "" || len(name) > 32 || name[0] == '-' || name[len(name)-1] == '-' {
			event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", "invalid name", status.Failed})
			AbortBadRequest(c)
			return
		}

		for i := 0; i < len(name); i++ {
			b := name[i]
			if b != '-' && (b < 'a' || b > 'z') && (b < '0' || b > '9') {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", "invalid name chars", status.Failed})
				AbortBadRequest(c)
				return
			}
		}

		// Validate advertise URL if provided (http/https allowed for intra-cluster routing).
		if u := strings.TrimSpace(req.AdvertiseUrl); u != "" {
			if !validateAdvertiseURL(u) {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", "invalid advertise url", status.Failed})
				AbortBadRequest(c)
				return
			}
		}

		// Validate site URL if provided (https required for non-local domains).
		if su := strings.TrimSpace(req.SiteUrl); su != "" {
			if !validateSiteURL(su) {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", "invalid site url", status.Failed})
				AbortBadRequest(c)
				return
			}
		}

		// Sanitize requested NodeUUID; generation happens later depending on path (existing vs new).
		requestedUUID := rnd.SanitizeUUID(req.NodeUUID)

		// Registry (client-backed).
		regy, err := reg.NewClientRegistryWithConfig(conf)

		if err != nil {
			event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "register", status.Error(err)})
			AbortUnexpectedError(c)
			return
		}

		portalTheme := ""
		if t, err := theme.DetectVersion(conf.PortalThemePath()); err == nil {
			portalTheme = t
		}

		// Join token is only valid in Authorization Bearer format for first-time joins.
		joinTokenValid := false
		if expected := strings.TrimSpace(conf.JoinToken()); expected != "" {
			if bearer := header.BearerToken(c); bearer != "" &&
				subtle.ConstantTimeCompare([]byte(expected), []byte(bearer)) == 1 {
				joinTokenValid = true
			}
		}

		// Resolve existing node by requested name before selecting auth mode.
		existingNode, _ := regy.FindByName(name)
		node := existingNode

		// Existing-name requests are mutation paths and require node OAuth ownership.
		if existingNode != nil {
			// A valid join token must never mutate an existing node registration.
			if joinTokenValid {
				r.Success()
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "join token cannot mutate existing node", status.Denied}, clean.Log(name))
				c.JSON(http.StatusConflict, gin.H{"error": registerNameConflictError(name)})
				return
			}

			if s := Auth(c, acl.ResourceCluster, acl.ActionUpdateOwn); s.Abort(c) {
				return
			} else if !s.IsClient() || s.ClientUID == "" {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "register mutation requires client token", status.Denied}, clean.Log(name))
				AbortUnauthorized(c)
				return
			} else if s.ClientUID != existingNode.ClientID {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "name claimed by different client", status.Denied}, clean.Log(name))
				c.JSON(http.StatusConflict, gin.H{"error": registerNameConflictError(name)})
				return
			}
		} else if !joinTokenValid {
			// Without a valid join token, only an authenticated node client may
			// mutate its own registration under a new (currently unused) name.
			s := Auth(c, acl.ResourceCluster, acl.ActionUpdateOwn)

			if s.Abort(c) {
				return
			}

			if !s.IsClient() || s.ClientUID == "" {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", "invalid join token", status.Denied})
				r.Success()
				AbortUnauthorized(c)
				return
			}

			owner, findErr := regy.FindByClientID(s.ClientUID)

			if findErr != nil || owner == nil {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", "client token is not a registered node", status.Denied})
				AbortUnauthorized(c)
				return
			}

			role := cluster.NormalizeNodeRole(owner.Role)

			if role != cluster.RoleInstance && role != cluster.RoleService {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", "client token role not allowed", status.Denied})
				AbortUnauthorized(c)
				return
			}

			node = owner
		}

		// Existing-node mutation path (including node-owned rename to an unused name).
		if node != nil {
			if oldName := node.Name; oldName != "" && oldName != name {
				node.Name = name
				event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "change name old %s new %s", status.Updated}, clean.Log(name), clean.Log(oldName), clean.Log(name))
			}

			if req.AdvertiseUrl != "" {
				node.AdvertiseUrl = req.AdvertiseUrl
			}
			if req.Labels != nil {
				node.Labels = req.Labels
			}
			if s := normalizeSiteURL(req.SiteUrl); s != "" {
				node.SiteUrl = s
			}
			if appName != "" {
				node.AppName = appName
			}
			if appVersion != "" {
				node.AppVersion = appVersion
			}
			if nodeTheme != "" {
				node.Theme = nodeTheme
			}

			if requestedUUID != "" {
				oldUUID := node.UUID
				if oldUUID != requestedUUID {
					node.UUID = requestedUUID
					event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "change uuid old %s new %s", status.Updated}, clean.Log(name), clean.Log(oldUUID), clean.Log(requestedUUID))
				}
			} else if node.UUID == "" {
				node.UUID = rnd.UUIDv7()
				event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "assign uuid %s", status.Created}, clean.Log(name), clean.Log(node.UUID))
			}

			if putErr := regy.Put(node); putErr != nil {
				event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "persist", status.Error(putErr)}, clean.Log(name))
				AbortUnexpectedError(c)
				return
			}

			var respSecret *cluster.RegisterSecrets

			if req.RotateSecret {
				if node, err = regy.RotateSecret(node.UUID); err != nil {
					event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "rotate secret", status.Error(err)}, clean.Log(name))
					AbortUnexpectedError(c)
					return
				}

				respSecret = &cluster.RegisterSecrets{ClientSecret: node.ClientSecret, RotatedAt: node.RotatedAt}
				event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "rotate secret", status.Succeeded}, clean.Log(name))

				if putErr := regy.Put(node); putErr != nil {
					event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "persist rotated secret", status.Error(putErr)}, clean.Log(name))
					AbortUnexpectedError(c)
					return
				}
			}

			var creds provisioner.Credentials
			haveCreds := false
			shouldProvisionDB := req.RotateDatabase

			if shouldProvisionDB {
				var credsErr error

				creds, _, credsErr = provisioner.EnsureCredentials(c, conf, node.UUID, node.Name, req.RotateDatabase)

				if credsErr != nil {
					event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "ensure database", status.Error(credsErr)}, clean.Log(name))
					c.JSON(http.StatusConflict, gin.H{"error": credsErr.Error()})
					return
				}

				haveCreds = true

				if node.Database == nil {
					node.Database = &cluster.NodeDatabase{}
				}

				node.Database.Name = creds.Name
				node.Database.User = creds.User
				node.Database.Driver = creds.Driver

				if creds.RotatedAt != "" {
					node.Database.RotatedAt = creds.RotatedAt
				}

				if putErr := regy.Put(node); putErr != nil {
					event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "persist", status.Error(putErr)}, clean.Log(name))
					AbortUnexpectedError(c)
					return
				}

				if req.RotateDatabase {
					event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "node", "%s", "rotate database", status.Succeeded}, clean.Log(name))
				}
			}

			resp := cluster.RegisterResponse{
				UUID:               conf.ClusterUUID(),
				ClusterCIDR:        conf.ClusterCIDR(),
				Node:               reg.BuildClusterNode(*node, reg.NodeOptsForSession(nil)),
				Secrets:            respSecret,
				JWKSUrl:            buildJWKSURL(conf),
				AlreadyRegistered:  true,
				AlreadyProvisioned: node.Database != nil && node.Database.Name != "",
			}

			if portalTheme != "" {
				resp.Theme = portalTheme
				log.Debugf("cluster: reporting portal theme hint %s for instance %s", clean.Log(portalTheme), clean.Log(name))
			}

			if node.Database != nil {
				driver := node.Database.Driver
				if driver == "" {
					driver = provisioner.DatabaseDriver
				}
				resp.Database = cluster.RegisterDatabase{Host: conf.DatabaseHost(), Port: conf.DatabasePort(), Name: node.Database.Name, User: node.Database.User, Driver: driver, RotatedAt: node.Database.RotatedAt}
			}

			if req.RotateDatabase && haveCreds {
				resp.Database.Password = creds.Password
				resp.Database.DSN = creds.DSN
				resp.Database.RotatedAt = creds.RotatedAt
			}

			event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "node", "%s", status.Confirmed}, clean.Log(name))
			c.JSON(http.StatusOK, resp)
			return
		}

		if !joinTokenValid {
			event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", "invalid join token", status.Denied})
			r.Success()
			AbortUnauthorized(c)
			return
		}

		// New node (client UID will be generated in registry.Put).
		n := &reg.Node{
			Node: cluster.Node{
				Name:       name,
				Role:       clean.TypeLowerDash(req.NodeRole),
				UUID:       requestedUUID,
				Labels:     req.Labels,
				AppName:    appName,
				AppVersion: appVersion,
				Theme:      nodeTheme,
			},
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
		shouldProvisionDB := req.RotateDatabase
		var creds provisioner.Credentials

		if shouldProvisionDB {
			if creds, _, err = provisioner.EnsureCredentials(c, conf, n.UUID, name, true); err != nil {
				event.AuditWarn([]string{clientIp, string(acl.ResourceCluster), "register", "ensure database", status.Error(err)})
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}

			if n.Database == nil {
				n.Database = &cluster.NodeDatabase{}
			}

			n.Database.Name, n.Database.User, n.Database.RotatedAt = creds.Name, creds.User, creds.RotatedAt
			n.Database.Driver = creds.Driver
		}

		if err = regy.Put(n); err != nil {
			event.AuditErr([]string{clientIp, string(acl.ResourceCluster), "register", "persist", status.Error(err)})
			AbortUnexpectedError(c)
			return
		}

		resp := cluster.RegisterResponse{
			UUID:               conf.ClusterUUID(),
			ClusterCIDR:        conf.ClusterCIDR(),
			Node:               reg.BuildClusterNode(*n, reg.NodeOptsForSession(nil)),
			Secrets:            &cluster.RegisterSecrets{ClientSecret: n.ClientSecret, RotatedAt: n.RotatedAt},
			JWKSUrl:            buildJWKSURL(conf),
			AlreadyRegistered:  false,
			AlreadyProvisioned: shouldProvisionDB,
		}

		if portalTheme != "" {
			resp.Theme = portalTheme
			log.Debugf("cluster: portal theme hint %s for instance %s", clean.Log(portalTheme), clean.Log(name))
		}

		// If DB provisioning is skipped, leave Database fields zero-value.
		if shouldProvisionDB {
			resp.Database = cluster.RegisterDatabase{Host: conf.DatabaseHost(), Port: conf.DatabasePort(), Name: creds.Name, User: creds.User, Driver: creds.Driver, Password: creds.Password, DSN: creds.DSN, RotatedAt: creds.RotatedAt}
		}

		event.AuditInfo([]string{clientIp, string(acl.ResourceCluster), "node", "%s", status.Joined}, clean.Log(name))
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

// validateAdvertiseURL checks that the URL is absolute with a host and scheme.
// HTTP and HTTPS are both allowed to support internal cluster traffic.
func validateAdvertiseURL(u string) bool {
	parsed, err := url.Parse(strings.TrimSpace(u))

	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	switch parsed.Scheme {
	case "http", "https":
		return true
	default:
		return false
	}
}

func buildJWKSURL(conf *config.Config) string {
	if conf == nil {
		return "/.well-known/jwks.json"
	}

	path := conf.BaseUri("/.well-known/jwks.json")

	if path == "" {
		path = "/.well-known/jwks.json"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	site := strings.TrimRight(conf.SiteUrl(), "/")

	if site == "" {
		return path
	}

	return site + path
}

// validateSiteURL checks that the URL is absolute with a host and scheme.
// HTTPS is required for non-local hosts; HTTP is only allowed for loopback or
// cluster-internal service domains.
func validateSiteURL(u string) bool {
	parsed, err := url.Parse(strings.TrimSpace(u))

	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	host := strings.ToLower(parsed.Hostname())

	if parsed.Scheme == "https" {
		return true
	}

	if parsed.Scheme == "http" {
		if host == "localhost" || host == "127.0.0.1" || host == "::1" || isClusterServiceHost(host) {
			return true
		}
		return false
	}

	return false
}

// isClusterServiceHost reports whether the host refers to a cluster-internal
// service DNS name so that HTTP can be permitted for intra-cluster traffic.
func isClusterServiceHost(host string) bool {
	host = strings.TrimSuffix(host, ".")

	if host == "" {
		return false
	}

	// Allow cluster internal service hosts.
	if strings.HasSuffix(host, ".svc") || strings.Contains(host, ".svc.") {
		return true
	}

	// Allow hosts with .local or .internal domain.
	if strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return true
	}

	return false
}

// clusterCIDRAllowsClientIP reports whether clientIP is within the configured cidr.
func clusterCIDRAllowsClientIP(cidr, clientIP string) bool {
	ip := net.ParseIP(clientIP)
	_, block, err := net.ParseCIDR(cidr)

	if err != nil || ip == nil || block == nil {
		return false
	}

	return block.Contains(ip)
}

// registerNameConflictError returns a clear operator-facing conflict message.
func registerNameConflictError(name string) string {
	return fmt.Sprintf(
		"node name %q is already registered; delete the stale registration first and retry join",
		clean.DNSLabel(name),
	)
}
