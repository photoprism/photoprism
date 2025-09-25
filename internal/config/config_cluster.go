package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/service/http/header"
)

// DefaultPortalUrl specifies the default portal URL with variable cluster domain.
var DefaultPortalUrl = "https://portal.${PHOTOPRISM_CLUSTER_DOMAIN}"
var DefaultNodeRole = cluster.RoleInstance

// ClusterDomain returns the cluster DOMAIN (lowercase DNS name; 1–63 chars).
func (c *Config) ClusterDomain() string {
	if c.options.ClusterDomain != "" {
		return strings.ToLower(c.options.ClusterDomain)
	}

	if _, d, found := c.deriveNodeNameAndDomainFromHttpHost(); found && d != "" {
		return d
	}

	// Attempt to derive from system configuration when not explicitly set.
	if d := deriveSystemDomain(); d != "" {
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

// IsPortal returns true if the configured node type is "portal".
func (c *Config) IsPortal() bool {
	return c.NodeRole() == cluster.RolePortal
}

// PortalConfigPath returns the path to the default configuration for cluster nodes.
func (c *Config) PortalConfigPath() string {
	return filepath.Join(c.ConfigPath(), fs.PortalDir)
}

// PortalThemePath returns the path to the theme files for cluster nodes to use.
func (c *Config) PortalThemePath() string {
	// Prefer the cluster-specific theme directory if it exists.
	if dir := filepath.Join(c.PortalConfigPath(), fs.ThemeDir); fs.PathExists(dir) {
		return dir
	}

	// Fallback to the default theme directory in the main config path.
	return c.ThemePath()
}

// JoinToken returns the token required to access the portal API endpoints.
func (c *Config) JoinToken() string {
	if c.options.JoinToken != "" {
		return c.options.JoinToken
	} else if fileName := FlagFilePath("JOIN_TOKEN"); fileName == "" {
		return ""
	} else if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 {
		log.Warnf("config: failed to read portal token from %s (%s)", fileName, err)
		return ""
	} else {
		return string(b)
	}
}

// deriveNodeNameAndDomainFromHttpHost attempts to derive cluster host and domain name from the site URL.
func (c *Config) deriveNodeNameAndDomainFromHttpHost() (hostName, domainName string, found bool) {
	if fqdn := c.SiteDomain(); fqdn != "" && !header.IsIP(fqdn) {
		hostName, domainName, found = strings.Cut(fqdn, ".")
		if hostName = clean.DNSLabel(hostName); found && isDNSLabel(hostName) && isDNSDomain(domainName) {
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
	if c.IsPortal() {
		return "portal"
	}

	// Instances/services: derive from hostname via DNSLabel normalization.
	if hn, _ := getHostname(); hn != "" {
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

// NodeClientSecret returns the OAuth client SECRET registered with the portal (auto-assigned via join token).
func (c *Config) NodeClientSecret() string {
	if c.options.NodeClientSecret != "" {
		return c.options.NodeClientSecret
	} else if fileName := FlagFilePath("NODE_CLIENT_SECRET"); fileName == "" {
		return ""
	} else if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 {
		log.Warnf("config: failed to read node client secret from %s (%s)", fileName, err)
		return ""
	} else {
		return string(b)
	}
}

// AdvertiseUrl returns the advertised node URL for intra-cluster calls (scheme://host[:port]).
func (c *Config) AdvertiseUrl() string {
	if c.options.AdvertiseUrl != "" {
		return strings.TrimRight(c.options.AdvertiseUrl, "/") + "/"
	}
	// Derive from cluster domain and node name if available; otherwise fall back to SiteUrl().
	if d := c.ClusterDomain(); d != "" {
		if n := c.NodeName(); n != "" && isDNSLabel(n) {
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

	var m map[string]interface{}

	if fs.FileExists(fileName) {
		if b, err := os.ReadFile(fileName); err == nil && len(b) > 0 {
			_ = yaml.Unmarshal(b, &m)
		}
	}

	if m == nil {
		m = map[string]interface{}{}
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

	var m map[string]interface{}
	if fs.FileExists(fileName) {
		if b, err := os.ReadFile(fileName); err == nil && len(b) > 0 {
			_ = yaml.Unmarshal(b, &m)
		}
	}
	if m == nil {
		m = map[string]interface{}{}
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
