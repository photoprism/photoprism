package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/service/cluster"
)

func TestConfig_Cluster(t *testing.T) {
	t.Run("Flags", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Defaults
		assert.False(t, c.ClusterPortal())
		assert.False(t, c.IsPortal())

		// Toggle values
		c.Options().NodeType = string(cluster.Portal)
		assert.True(t, c.ClusterPortal())
		assert.True(t, c.IsPortal())
		c.Options().NodeType = ""
	})

	t.Run("Paths", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Use an isolated config path so we don't affect repo storage fixtures.
		tempCfg := t.TempDir()
		c.options.ConfigPath = tempCfg

		// PortalConfigPath always points to a "cluster" subfolder under ConfigPath.
		expectedCluster := filepath.Join(c.ConfigPath(), fs.ClusterDir)
		assert.Equal(t, expectedCluster, c.PortalConfigPath())

		// PortalThemePath falls back to ThemePath if cluster dir does not exist.
		expectedTheme := filepath.Join(c.ConfigPath(), fs.ThemeDir)
		assert.Equal(t, expectedTheme, c.PortalThemePath())

		// When only the cluster directory exists (without a theme subfolder), it still falls back to ThemePath.
		assert.NoError(t, os.MkdirAll(expectedCluster, 0o755))
		assert.Equal(t, expectedTheme, c.PortalThemePath())

		// When the cluster theme directory exists, PortalThemePath returns it.
		expectedClusterTheme := filepath.Join(expectedCluster, fs.ThemeDir)
		assert.NoError(t, os.MkdirAll(expectedClusterTheme, 0o755))
		assert.Equal(t, expectedClusterTheme, c.PortalThemePath())
	})

	t.Run("PortalAndSecrets", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Defaults
		assert.Equal(t, "", c.PortalUrl())
		assert.Equal(t, "", c.PortalToken())
		assert.Equal(t, "", c.NodeSecret())

		// Set and read back values
		c.options.PortalUrl = "https://portal.example.test"
		c.options.PortalToken = "portal-token"
		c.options.NodeSecret = "node-secret"

		assert.Equal(t, "https://portal.example.test", c.PortalUrl())
		assert.Equal(t, "portal-token", c.PortalToken())
		assert.Equal(t, "node-secret", c.NodeSecret())
	})

	t.Run("AbsolutePaths", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		tempCfg := t.TempDir()
		c.options.ConfigPath = tempCfg

		// ThemePath should be absolute.
		assert.True(t, filepath.IsAbs(c.ThemePath()))

		// PortalThemePath should be absolute (fallback case).
		assert.True(t, filepath.IsAbs(c.PortalThemePath()))

		// Create cluster theme directory and verify again.
		clusterTheme := filepath.Join(c.PortalConfigPath(), fs.ThemeDir)
		assert.NoError(t, os.MkdirAll(clusterTheme, 0o755))
		assert.True(t, filepath.IsAbs(c.PortalThemePath()))
	})

	t.Run("NodeName", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.NodeName = " Client Credentials幸"
		assert.Equal(t, "client-credentials", c.NodeName())
		c.options.NodeName = ""
		assert.Equal(t, "", c.NodeName())
	})

	t.Run("NodeTypeValues", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Default / unknown → node
		c.options.NodeType = ""
		assert.Equal(t, string(cluster.Instance), c.NodeType())
		c.options.NodeType = "unknown"
		assert.Equal(t, string(cluster.Instance), c.NodeType())

		// Explicit values
		c.options.NodeType = string(cluster.Instance)
		assert.Equal(t, string(cluster.Instance), c.NodeType())
		c.options.NodeType = string(cluster.Portal)
		assert.Equal(t, string(cluster.Portal), c.NodeType())
		c.options.NodeType = string(cluster.Service)
		assert.Equal(t, string(cluster.Service), c.NodeType())
	})

	t.Run("SecretsFromFiles", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Create temp secret/token files.
		dir := t.TempDir()
		nsFile := filepath.Join(dir, "node_secret")
		tkFile := filepath.Join(dir, "portal_token")
		assert.NoError(t, os.WriteFile(nsFile, []byte("s3cr3t"), 0o600))
		assert.NoError(t, os.WriteFile(tkFile, []byte("t0k3n"), 0o600))

		// Clear inline values so file-based lookup is used.
		c.options.NodeSecret = ""
		c.options.PortalToken = ""

		// Point env vars at the files and verify.
		t.Setenv("PHOTOPRISM_NODE_SECRET_FILE", nsFile)
		t.Setenv("PHOTOPRISM_PORTAL_TOKEN_FILE", tkFile)
		assert.Equal(t, "s3cr3t", c.NodeSecret())
		assert.Equal(t, "t0k3n", c.PortalToken())

		// Empty / missing should yield empty strings.
		t.Setenv("PHOTOPRISM_NODE_SECRET_FILE", filepath.Join(dir, "missing"))
		t.Setenv("PHOTOPRISM_PORTAL_TOKEN_FILE", filepath.Join(dir, "missing"))
		assert.Equal(t, "", c.NodeSecret())
		assert.Equal(t, "", c.PortalToken())
	})
}
