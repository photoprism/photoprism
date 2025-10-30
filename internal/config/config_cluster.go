package config

import (
	"errors"
	"fmt"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/internal/service/cluster/theme"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/dns"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// DefaultPortalUrl specifies the default portal URL with variable cluster domain.
var DefaultPortalUrl = "https://portal.${PHOTOPRISM_CLUSTER_DOMAIN}"
var DefaultNodeRole = cluster.RoleInstance
var DefaultJWTAllowedScopes = "config cluster vision metrics"

// ClusterDomain returns the cluster DOMAIN (lowercase DNS name; 1–63 chars).
func (c *Config) ClusterDomain() string {
	if c.options.ClusterDomain != "" {
		return strings.ToLower(c.options.ClusterDomain)
	}

	if _, d, found := c.deriveNodeNameAndDomainFromHttpHost(); found && d != "" {
		return d
	}

	// Attempt to derive from system configuration when not explicitly set.
	if d := dns.GetSystemDomain(); d != "" {
		return d
	}

	return ""
}

// ClusterCIDR returns the configured cluster CIDR used for IP-based allowances.
func (c *Config) ClusterCIDR() string {
	return strings.TrimSpace(c.options.ClusterCIDR)
}

// ClusterUUID returns a stable UUIDv4 that uniquely identifies the Portal.
// Precedence: env PHOTOPRISM_CLUSTER_UUID -> options.yml (ClusterUUID) -> auto-generate and persist.
func (c *Config) ClusterUUID() string {
	// Return if the configured cluster UUID is not in the expected format.
	if !rnd.IsUUID(c.options.ClusterUUID) {
		return ""
	}

	// Respect explicit CLI value if provided.
	if c.cliCtx != nil && c.cliCtx.IsSet("cluster-uuid") {
		return c.options.ClusterUUID
	}

	return c.options.ClusterUUID
}

// PortalUrl returns the URL of the cluster management portal server, if configured.
func (c *Config) PortalUrl() string {
	if c.options.PortalUrl == "" {
		return ""
	}

	d := c.ClusterDomain()

	// Return empty string if default and there's no cluster domain configured.
	if d == "" && c.options.PortalUrl == DefaultPortalUrl {
		return ""
	}

	// Replace variables with the configured cluster domain.
	c.options.PortalUrl = ExpandVars(c.options.PortalUrl, map[string]string{
		"cluster-domain":            d,
		"CLUSTER_DOMAIN":            d,
		"PHOTOPRISM_CLUSTER_DOMAIN": d,
	})

	return c.options.PortalUrl
}

// Portal returns true if the configured node type is "portal".
func (c *Config) Portal() bool {
	return c.NodeRole() == cluster.RolePortal
}

// PortalConfigPath returns the path to the default configuration for cluster portals.
func (c *Config) PortalConfigPath() string {
	return filepath.Join(c.ConfigPath(), fs.PortalDir)
}

// PortalThemePath returns the path to the theme files for cluster portals to use.
func (c *Config) PortalThemePath() string {
	themeDir := filepath.Join(c.PortalConfigPath(), fs.ThemeDir)

	if fs.PathExists(themeDir) && fs.FileExists(filepath.Join(themeDir, fs.AppJsFile)) {
		return themeDir
	}

	// Fallback to the default theme directory in the main config path.
	return c.ThemePath()
}

// NodeConfigPath returns the path to the default configuration for cluster nodes.
func (c *Config) NodeConfigPath() string {
	return filepath.Join(c.ConfigPath(), fs.NodeDir)
}

// NodeThemePath returns the path to the theme files for cluster nodes to use.
func (c *Config) NodeThemePath() string {
	return filepath.Join(c.NodeConfigPath(), fs.ThemeDir)
}

// NodeThemeVersion returns the version to the theme files of the cluster node.
func (c *Config) NodeThemeVersion() string {
	if version, err := theme.DetectVersion(c.NodeThemePath()); err == nil {
		return version
	}

	return ""
}

// JoinToken returns the portal join token used when registering nodes. It
// lazily loads the token from disk (or generates a new one) and caches it in
// memory. Example format: k9sEFe6-A7gt6zqm-gY9gFh0.
func (c *Config) JoinToken() string {
	// Read token from config options (memory).
	if rnd.IsJoinToken(c.options.JoinToken, false) {
		return c.options.JoinToken
	}

	// Read token from file if possible. Uses a cache to reduce I/O.
	if fileName := c.JoinTokenFile(); fileName != "" {
		if c.cache == nil {
			// Skip cache lookup.
		} else if s, hit := c.cache.Get(fileName); hit && s != nil {
			return s.(string)
		}

		if fs.FileExistsNotEmpty(fileName) {
			if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 {
				log.Warnf("config: could not read cluster join token from %s (%s)", fileName, err)
			} else if s := strings.TrimSpace(string(b)); rnd.IsJoinToken(s, false) {
				if c.cache != nil {
					c.cache.SetDefault(fileName, s)
				}
				return s
			} else {
				log.Warnf("config: cluster join token from %s is shorter than %d characters", fileName, rnd.JoinTokenLength)
			}
		}
	}

	// Do not proceed with generating a token on nodes.
	if !c.Portal() {
		return ""
	} else if token, _, err := c.SaveJoinToken(""); err != nil {
		log.Errorf("config: %v", err)
		return ""
	} else {
		return token
	}
}

// SaveJoinToken writes a fresh portal join token to disk and updates the
// in-memory value. When customToken is provided it must already be valid.
func (c *Config) SaveJoinToken(customToken string) (token string, fileName string, err error) {
	fileName = c.JoinTokenFile()

	if fileName == "" {
		return "", "", fmt.Errorf("invalid cluster join token path")
	}

	dir := filepath.Dir(fileName)
	if dir == "" {
		return "", "", fmt.Errorf("invalid cluster secrets directory")
	}

	if customToken != "" {
		if !rnd.IsJoinToken(customToken, false) {
			return "", "", fmt.Errorf("insecure custom cluster join token specified")
		}
		token = customToken
	} else {
		token = rnd.JoinToken()
		if !rnd.IsJoinToken(token, true) {
			return "", "", fmt.Errorf("invalid cluster join token generated")
		}
	}

	// Create secret directory.
	if err = fs.MkdirAll(dir); err != nil {
		// Use memory to store join token if directory is not writable.
		c.options.JoinToken = token
		return "", "", fmt.Errorf("could not create cluster secrets path (%w)", err)
	}

	// Write secret to file.
	if err = fs.WriteFile(fileName, []byte(token), fs.ModeSecretFile); err != nil {
		// Use memory to store join token if file is not writable.
		c.options.JoinToken = token
		return "", "", fmt.Errorf("could not write cluster join token (%w)", err)
	}

	// Use an in-memory cache with a
	// short TTL to cache the token.
	if c.cache != nil {
		c.cache.SetDefault(fileName, token)
		c.options.JoinToken = ""
	} else {
		// Store token in Options
		// if cache is unavailable.
		c.options.JoinToken = token
	}

	return token, fileName, nil
}

// clearJoinTokenFileCache invalidates the cached join token file cache.
func (c *Config) clearJoinTokenFileCache() {
	if c.cache != nil {
		c.cache.Delete(c.JoinTokenFile())
	}
}

// JoinTokenFile returns the path where the portal join token is stored for the
// active configuration (portal nodes use config/portal/secrets/join_token,
// regular nodes use config/node/secrets/join_token).
func (c *Config) JoinTokenFile() string {
	if c.Portal() {
		return c.PortalJoinTokenFile()
	}

	return c.NodeJoinTokenFile()
}

// PortalJoinTokenFile returns the filepath where the portal cluster join token is stored.
func (c *Config) PortalJoinTokenFile() string {
	if filePath := FlagFilePath("JOIN_TOKEN"); filePath != "" {
		return filePath
	}

	return filepath.Join(c.PortalConfigPath(), fs.SecretsDir, fs.JoinTokenFile)

}

// NodeJoinTokenFile returns the filepath where the node cluster join token is stored.
func (c *Config) NodeJoinTokenFile() string {
	if filePath := FlagFilePath("JOIN_TOKEN"); filePath != "" {
		return filePath
	}

	return filepath.Join(c.NodeConfigPath(), fs.SecretsDir, fs.JoinTokenFile)
}

// deriveNodeNameAndDomainFromHttpHost attempts to derive cluster host and domain name from the site URL.
func (c *Config) deriveNodeNameAndDomainFromHttpHost() (hostName, domainName string, found bool) {
	if fqdn := c.SiteDomain(); fqdn != "" && !header.IsIP(fqdn) {
		hostName, domainName, found = strings.Cut(fqdn, ".")
		if hostName = clean.DNSLabel(hostName); found && dns.IsLabel(hostName) && dns.IsDomain(domainName) {
			c.options.NodeName = hostName
			if c.options.ClusterDomain == "" {
				c.options.ClusterDomain = strings.ToLower(domainName)
			}
			return c.options.NodeName, c.options.ClusterDomain, found
		}
	}

	return "", "", false
}

// NodeName returns the cluster node NAME (unique in cluster domain; [a-z0-9-]{1,32}).
func (c *Config) NodeName() string {
	if n := clean.DNSLabel(c.options.NodeName); n != "" {
		return n
	}

	if h, _, found := c.deriveNodeNameAndDomainFromHttpHost(); found && h != "" {
		return h
	}

	// Default: portal nodes → "portal".
	if c.Portal() {
		return "portal"
	}

	// Instances/services: derive from hostname via DNSLabel normalization.
	if hn, _ := dns.GetHostname(); hn != "" {
		if cand := clean.DNSLabel(hn); cand != "" {
			return cand
		}
	}

	// Fallback to a stable short identifier
	s := c.SerialChecksum()
	return "node-" + s
}

// NodeRole returns the cluster node ROLE (portal, instance, or service).
func (c *Config) NodeRole() string {
	if c.Edition() == Portal {
		c.options.NodeRole = cluster.RolePortal
		return c.options.NodeRole
	}

	switch c.options.NodeRole {
	case cluster.RolePortal, cluster.RoleInstance, cluster.RoleService:
		return c.options.NodeRole
	default:
		return DefaultNodeRole
	}
}

// NodeUUID returns the UUID (v7) that identifies this node.
func (c *Config) NodeUUID() string {
	if c.options.NodeUUID != "" {
		return c.options.NodeUUID
	}

	// Generate, persist, and cache a UUIDv7 if still empty.
	uuid := rnd.UUIDv7()
	c.options.NodeUUID = uuid

	if err := c.SaveNodeUUID(uuid); err != nil {
		log.Warnf("config: could not save node UUID to %s (%s)", c.OptionsYaml(), err)
	}

	return uuid
}

// NodeClientID returns the OAuth client ID registered with the portal (auto-assigned via join token).
func (c *Config) NodeClientID() string {
	return clean.ID(c.options.NodeClientID)
}

// NodeClientSecret returns the node OAuth client secret, reading it from disk
// when necessary. Portal registration writes this secret so nodes can obtain
// access tokens in future runs.
func (c *Config) NodeClientSecret() string {
	if c.options.NodeClientSecret != "" {
		return c.options.NodeClientSecret
	}

	fileName := c.NodeClientSecretFile()

	if fileName == "" {
		return ""
	}

	if b, err := os.ReadFile(fileName); err == nil && len(b) > 0 {
		// Do not cache the value. Always read from the disk to ensure
		// that updates from other processes are observed.
		return string(b)
	}

	if err := os.Chmod(filepath.Dir(fileName), fs.ModeDir); err != nil {
		log.Debugf("config: failed to set node secrets dir permissions (%s)", err)
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		log.Debugf("config: node client secret file %s not found", clean.Log(fileName))
	} else if err != nil {
		log.Warnf("config: failed to read node client secret from %s (%s)", clean.Log(fileName), err)
	}

	return c.options.NodeClientSecret
}

// SaveNodeClientSecret stores a new node client secret on disk and updates the
// in-memory value. The secret must already pass rnd.IsClientSecret.
func (c *Config) SaveNodeClientSecret(clientSecret string) (fileName string, err error) {
	fileName = c.NodeClientSecretFile()

	if !rnd.IsClientSecret(clientSecret) {
		return fileName, errors.New("invalid node client secret")
	}

	dir := filepath.Dir(fileName)
	if fileName == "" || dir == "" {
		return fileName, fmt.Errorf("invalid node client secret filename %s", clean.Log(fileName))
	}

	// Create secret directory.
	if err = fs.MkdirAll(dir); err != nil {
		// Use memory to store client secret if directory is not writable.
		c.options.NodeClientSecret = clientSecret
		return fileName, fmt.Errorf("could not create node secrets path (%s)", err)
	}

	// Write secret to file.
	if err = fs.WriteFile(fileName, []byte(clientSecret), fs.ModeSecretFile); err != nil {
		// Use memory to store client secret if file is not writable.
		c.options.NodeClientSecret = clientSecret
		return "", fmt.Errorf("could not write node client secret (%s)", err)
	}

	c.options.NodeClientSecret = ""

	return fileName, nil
}

// NodeClientSecretFile returns the path holding the node client secret (defaults
// to config/node/secrets/client_secret unless overridden via *_FILE).
func (c *Config) NodeClientSecretFile() string {
	if filePath := FlagFilePath("NODE_CLIENT_SECRET"); filePath != "" {
		return filePath
	}

	return filepath.Join(c.NodeConfigPath(), fs.SecretsDir, fs.ClientSecretFile)
}

// JWKSUrl returns the configured JWKS endpoint for portal-issued JWTs. Nodes normally
// persist this URL from the portal's register response, which derives it from SiteUrl;
// manual overrides are only required for custom deployments.
func (c *Config) JWKSUrl() string {
	return strings.TrimSpace(c.options.JWKSUrl)
}

// SetJWKSUrl updates the configured JWKS endpoint for portal-issued JWTs.
func (c *Config) SetJWKSUrl(url string) {
	if c == nil || c.options == nil {
		return
	}

	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		c.options.JWKSUrl = ""
		return
	}

	parsed, err := urlpkg.Parse(trimmed)
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Host == "" {
		log.Warnf("config: ignoring JWKS URL %q (%v)", trimmed, err)
		return
	}

	scheme := strings.ToLower(parsed.Scheme)
	host := parsed.Hostname()

	switch scheme {
	case "https":
		// Always allowed.
	case "http":
		if !dns.IsLoopbackHost(host) {
			log.Warnf("config: rejecting JWKS URL %q (http only allowed for localhost/loopback)", trimmed)
			return
		}
	default:
		log.Warnf("config: rejecting JWKS URL %q (unsupported scheme)", trimmed)
		return
	}

	c.options.JWKSUrl = trimmed
}

// JWKSCacheTTL returns the JWKS cache lifetime in seconds (default 300, max 3600).
func (c *Config) JWKSCacheTTL() int {
	if c.options.JWKSCacheTTL <= 0 {
		return 300
	}
	if c.options.JWKSCacheTTL > 3600 {
		return 3600
	}
	return c.options.JWKSCacheTTL
}

// JWTLeeway returns the permitted clock skew in seconds (default 60, max 300).
func (c *Config) JWTLeeway() int {
	if c.options.JWTLeeway <= 0 {
		return 60
	}
	if c.options.JWTLeeway > 300 {
		return 300
	}
	return c.options.JWTLeeway
}

// JWTAllowedScopes returns an optional allow-list of accepted JWT scopes.
func (c *Config) JWTAllowedScopes() list.Attr {
	if s := strings.TrimSpace(c.options.JWTScope); s != "" {
		parsed := list.ParseAttr(strings.ToLower(s))
		if len(parsed) > 0 {
			return parsed
		}
	}

	return list.ParseAttr(DefaultJWTAllowedScopes)
}

// AdvertiseUrl returns the advertised node URL for intra-cluster calls (scheme://host[:port]).
// Portal validation permits HTTPS for external hosts and HTTP only for loopback
// or cluster-internal service domains (e.g., *.svc, *.cluster.local, *.internal).
func (c *Config) AdvertiseUrl() string {
	if c.options.AdvertiseUrl != "" {
		return strings.TrimRight(c.options.AdvertiseUrl, "/") + "/"
	}
	// Derive from cluster domain and node name if available; otherwise fall back to SiteUrl().
	if d := c.ClusterDomain(); d != "" {
		if n := c.NodeName(); n != "" && dns.IsLabel(n) {
			return "https://" + n + "." + d + "/"
		}
	}
	return c.SiteUrl()
}

// SaveClusterUUID writes or updates the ClusterUUID key in options.yml without
// touching unrelated keys. Creates the file and directories if needed.
func (c *Config) SaveClusterUUID(uuid string) error {
	if !rnd.IsUUID(uuid) {
		return errors.New("invalid cluster UUID")
	}

	// Always resolve against the current ConfigPath and remember it explicitly
	// so subsequent calls don't accidentally point to a previous default.
	cfgDir := c.ConfigPath()
	if err := fs.MkdirAll(cfgDir); err != nil {
		return err
	}

	fileName := c.OptionsYaml()

	var m Values

	if fs.FileExists(fileName) {
		if b, err := os.ReadFile(fileName); err == nil && len(b) > 0 {
			_ = yaml.Unmarshal(b, &m)
		}
	}

	if m == nil {
		m = Values{}
	}

	m["ClusterUUID"] = uuid

	if b, err := yaml.Marshal(m); err != nil {
		return err
	} else if err = os.WriteFile(fileName, b, fs.ModeFile); err != nil {
		return err
	}

	c.options.ClusterUUID = uuid

	// Remember options.yml path for subsequent loads and ensure in-memory options see the value.
	if c.options != nil {
		_ = c.options.Load(fileName)
	}

	return nil
}

// SaveNodeUUID writes or updates the NodeUUID key in options.yml without touching unrelated keys.
func (c *Config) SaveNodeUUID(uuid string) error {
	if !rnd.IsUUID(uuid) {
		return errors.New("invalid node UUID")
	}

	cfgDir := c.ConfigPath()

	if err := fs.MkdirAll(cfgDir); err != nil {
		return err
	}

	fileName := c.OptionsYaml()

	var m Values
	if fs.FileExists(fileName) {
		if b, err := os.ReadFile(fileName); err == nil && len(b) > 0 {
			_ = yaml.Unmarshal(b, &m)
		}
	}
	if m == nil {
		m = Values{}
	}
	m["NodeUUID"] = uuid
	if b, err := yaml.Marshal(m); err != nil {
		return err
	} else if err = os.WriteFile(fileName, b, fs.ModeFile); err != nil {
		return err
	}

	c.options.NodeUUID = uuid

	return nil
}
