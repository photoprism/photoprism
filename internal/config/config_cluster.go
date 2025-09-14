package config

import (
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/service/cluster"
)

// NodeName returns the unique name of this node within the cluster (lowercase letters and numbers only).
func (c *Config) NodeName() string {
	return clean.TypeLowerDash(c.options.NodeName)
}

// NodeType returns the type of this node for cluster operation (portal, instance, service).
func (c *Config) NodeType() string {
	switch c.options.NodeType {
	case cluster.Portal, cluster.Instance, cluster.Service:
		return c.options.NodeType
	default:
		return cluster.Instance
	}
}

// NodeSecret returns the private node key for intra-cluster communication.
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

// PortalUrl returns the URL of the cluster portal server, if configured.
func (c *Config) PortalUrl() string {
	return c.options.PortalUrl
}

// PortalToken returns the token required to access the portal API endpoints.
func (c *Config) PortalToken() string {
	if c.options.PortalToken != "" {
		return c.options.PortalToken
	} else if fileName := FlagFilePath("PORTAL_TOKEN"); fileName == "" {
		return ""
	} else if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 {
		log.Warnf("config: failed to read portal token from %s (%s)", fileName, err)
		return ""
	} else {
		return string(b)
	}
}

// ClusterPortal returns true if this instance should act as a cluster portal.
func (c *Config) ClusterPortal() bool {
	return c.IsPortal()
}

// IsPortal returns true if the configured node type is "portal".
func (c *Config) IsPortal() bool {
	return c.NodeType() == cluster.Portal
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
