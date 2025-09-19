package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// ClusterDomain returns the cluster DOMAIN (lowercase DNS name; 1–63 chars).
func (c *Config) ClusterDomain() string {
	return c.options.ClusterDomain
}

// ClusterUUID returns a stable UUIDv4 that uniquely identifies the Portal.
// Precedence: env PHOTOPRISM_CLUSTER_UUID -> options.yml (ClusterUUID) -> auto-generate and persist.
func (c *Config) ClusterUUID() string {
	// Use value loaded into options only if it is persisted in the current options.yml.
	// This avoids tests (or defaults) loading a UUID from an unrelated file path.
	if c.options.ClusterUUID != "" {
		// Respect explicit CLI value if provided.
		if c.cliCtx != nil && c.cliCtx.IsSet("cluster-uuid") {
			return c.options.ClusterUUID
		}
		// Otherwise, only trust a persisted value from the current options.yml.
		if fs.FileExists(c.OptionsYaml()) {
			return c.options.ClusterUUID
		}
	}

	// Generate, persist, and cache in memory if still empty.
	id := rnd.UUID()
	c.options.ClusterUUID = id

	if err := c.saveClusterUUID(id); err != nil {
		log.Warnf("config: failed to persist ClusterUUID to %s (%s)", c.OptionsYaml(), err)
	}

	return id
}

// PortalUrl returns the URL of the cluster portal server, if configured.
func (c *Config) PortalUrl() string {
	return c.options.PortalUrl
}

// IsPortal returns true if the configured node type is "portal".
func (c *Config) IsPortal() bool {
	return c.NodeRole() == cluster.RolePortal
}

// PortalConfigPath returns the path to the default configuration for cluster nodes.
func (c *Config) PortalConfigPath() string {
	return filepath.Join(c.ConfigPath(), fs.ClusterDir)
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

// NodeName returns the cluster node NAME (unique in cluster domain; [a-z0-9-]{1,32}).
func (c *Config) NodeName() string {
	return clean.TypeLowerDash(c.options.NodeName)
}

// NodeRole returns the cluster node ROLE (portal, instance, or service).
func (c *Config) NodeRole() string {
	switch c.options.NodeRole {
	case cluster.RolePortal, cluster.RoleInstance, cluster.RoleService:
		return c.options.NodeRole
	default:
		return cluster.RoleInstance
	}
}

// NodeID returns the client ID registered with the portal (auto-assigned via join token).
func (c *Config) NodeID() string {
	return clean.ID(c.options.NodeID)
}

// NodeSecret returns client SECRET registered with the portal (auto-assigned via join token).
func (c *Config) NodeSecret() string {
	if c.options.NodeSecret != "" {
		return c.options.NodeSecret
	} else if fileName := FlagFilePath("NODE_SECRET"); fileName == "" {
		return ""
	} else if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 {
		log.Warnf("config: failed to read node secret from %s (%s)", fileName, err)
		return ""
	} else {
		return string(b)
	}
}

// AdvertiseUrl returns the advertised node URL for intra-cluster calls (scheme://host[:port]).
func (c *Config) AdvertiseUrl() string {
	if c.options.AdvertiseUrl == "" {
		return c.SiteUrl()
	}

	return strings.TrimRight(c.options.AdvertiseUrl, "/") + "/"
}

// saveClusterUUID writes or updates the ClusterUUID key in options.yml without
// touching unrelated keys. Creates the file and directories if needed.
func (c *Config) saveClusterUUID(id string) error {
	// Always resolve against the current ConfigPath and remember it explicitly
	// so subsequent calls don't accidentally point to a previous default.
	cfgDir := c.ConfigPath()
	if err := fs.MkdirAll(cfgDir); err != nil {
		return err
	}
	fileName := filepath.Join(cfgDir, "options.yml")

	var m map[string]interface{}

	if fs.FileExists(fileName) {
		if b, err := os.ReadFile(fileName); err == nil && len(b) > 0 {
			_ = yaml.Unmarshal(b, &m)
		}
	}

	if m == nil {
		m = map[string]interface{}{}
	}

	m["ClusterUUID"] = id

	if b, err := yaml.Marshal(m); err != nil {
		return err
	} else if err = os.WriteFile(fileName, b, 0o644); err != nil {
		return err
	}

	// Remember options.yml path for subsequent loads and ensure in-memory options see the value.
	if c.options != nil {
		c.options.OptionsYaml = fileName
		_ = c.options.Load(fileName)
	}

	return nil
}
