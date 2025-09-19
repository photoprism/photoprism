package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestConfig_Cluster(t *testing.T) {
	t.Run("Flags", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Defaults
		assert.False(t, c.IsPortal())

		// Toggle values
		c.Options().NodeRole = string(cluster.RolePortal)
		assert.True(t, c.IsPortal())
		c.Options().NodeRole = ""
	})

	t.Run("Paths", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Use an isolated config path so we don't affect repo storage fixtures.
		tempCfg := t.TempDir()
		c.options.ConfigPath = tempCfg
		c.options.NodeSecret = ""
		c.options.PortalUrl = ""
		c.options.JoinToken = ""
		c.options.OptionsYaml = filepath.Join(tempCfg, "options.yml")
		// Clear values potentially loaded at NewConfig creation.
		c.options.NodeSecret = ""
		c.options.PortalUrl = ""
		c.options.JoinToken = ""
		c.options.OptionsYaml = filepath.Join(tempCfg, "options.yml")
		// Clear values that may have been loaded from repo fixtures before we
		// isolated the config path.
		c.options.NodeSecret = ""
		c.options.PortalUrl = ""
		c.options.JoinToken = ""
		c.options.OptionsYaml = filepath.Join(tempCfg, "options.yml")

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
		// Isolate config so defaults aren't overridden by repo fixtures: set config-path
		// before creating the Config so NewConfig does not load repository options.yml.
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)

		// Defaults (no options.yml present)
		assert.Equal(t, "", c.PortalUrl())
		assert.Equal(t, "", c.JoinToken())
		assert.Equal(t, "", c.NodeSecret())

		// Set and read back values
		c.options.PortalUrl = "https://portal.example.test"
		c.options.JoinToken = "join-token"
		c.options.NodeSecret = "node-secret"

		assert.Equal(t, "https://portal.example.test", c.PortalUrl())
		assert.Equal(t, "join-token", c.JoinToken())
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

	t.Run("NodeRoleValues", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Default / unknown → node
		c.options.NodeRole = ""
		assert.Equal(t, string(cluster.RoleInstance), c.NodeRole())
		c.options.NodeRole = "unknown"
		assert.Equal(t, string(cluster.RoleInstance), c.NodeRole())

		// Explicit values
		c.options.NodeRole = string(cluster.RoleInstance)
		assert.Equal(t, string(cluster.RoleInstance), c.NodeRole())
		c.options.NodeRole = string(cluster.RolePortal)
		assert.Equal(t, string(cluster.RolePortal), c.NodeRole())
		c.options.NodeRole = string(cluster.RoleService)
		assert.Equal(t, string(cluster.RoleService), c.NodeRole())
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
		c.options.JoinToken = ""

		// Point env vars at the files and verify.
		t.Setenv("PHOTOPRISM_NODE_SECRET_FILE", nsFile)
		t.Setenv("PHOTOPRISM_JOIN_TOKEN_FILE", tkFile)
		assert.Equal(t, "s3cr3t", c.NodeSecret())
		assert.Equal(t, "t0k3n", c.JoinToken())

		// Empty / missing should yield empty strings.
		t.Setenv("PHOTOPRISM_NODE_SECRET_FILE", filepath.Join(dir, "missing"))
		t.Setenv("PHOTOPRISM_JOIN_TOKEN_FILE", filepath.Join(dir, "missing"))
		assert.Equal(t, "", c.NodeSecret())
		assert.Equal(t, "", c.JoinToken())
	})
}

func TestConfig_ClusterUUID_FileOverridesEnv(t *testing.T) {
	c := NewConfig(CliTestContext())

	// Isolate config path.
	tempCfg := t.TempDir()
	c.options.ConfigPath = tempCfg

	// Prepare options.yml with a UUID; file should override env/CLI.
	opts := map[string]any{"ClusterUUID": "11111111-1111-4111-8111-111111111111"}
	b, _ := yaml.Marshal(opts)
	assert.NoError(t, os.WriteFile(filepath.Join(tempCfg, "options.yml"), b, 0o644))

	// Set env; file value must win for consistency with other options.
	t.Setenv("PHOTOPRISM_CLUSTER_UUID", "22222222-2222-4222-8222-222222222222")
	// Load options.yml into options struct (we updated ConfigPath after creation).
	assert.NoError(t, c.options.Load(c.OptionsYaml()))
	got := c.ClusterUUID()
	assert.Equal(t, "11111111-1111-4111-8111-111111111111", got)
}

func TestConfig_ClusterUUID_FromOptions(t *testing.T) {
	c := NewConfig(CliTestContext())
	tempCfg := t.TempDir()
	c.options.ConfigPath = tempCfg

	opts := map[string]any{"ClusterUUID": "33333333-3333-4333-8333-333333333333"}
	b, _ := yaml.Marshal(opts)
	assert.NoError(t, os.WriteFile(filepath.Join(tempCfg, "options.yml"), b, 0o644))

	// Ensure env is not set.
	t.Setenv("PHOTOPRISM_CLUSTER_UUID", "")

	// Load options.yml into options struct (we updated ConfigPath after creation).
	assert.NoError(t, c.options.Load(c.OptionsYaml()))
	// Access the value via getter.
	got := c.ClusterUUID()
	assert.Equal(t, "33333333-3333-4333-8333-333333333333", got)
}

func TestConfig_ClusterUUID_FromCLIFlag(t *testing.T) {
	// Create a config path so NewConfig reads/writes here and options.yml does not exist.
	tempCfg := t.TempDir()

	// Start from the default CLI test context and override flags we care about.
	ctx := CliTestContext()
	assert.NoError(t, ctx.Set("config-path", tempCfg))
	assert.NoError(t, ctx.Set("cluster-uuid", "44444444-4444-4444-8444-444444444444"))

	c := NewConfig(ctx)

	// No env and no options.yml: should take the CLI flag value directly from options.
	t.Setenv("PHOTOPRISM_CLUSTER_UUID", "")
	got := c.ClusterUUID()
	assert.Equal(t, "44444444-4444-4444-8444-444444444444", got)
}

func TestConfig_ClusterUUID_GenerateAndPersist(t *testing.T) {
	c := NewConfig(CliTestContext())
	tempCfg := t.TempDir()
	c.options.ConfigPath = tempCfg

	// No env, no options.yml → should generate and persist.
	t.Setenv("PHOTOPRISM_CLUSTER_UUID", "")

	got := c.ClusterUUID()
	if !rnd.IsUUID(got) {
		t.Fatalf("expected a UUIDv4, got %q", got)
	}

	// Verify content persisted to options.yml.
	b, err := os.ReadFile(filepath.Join(tempCfg, "options.yml"))
	assert.NoError(t, err)
	var m map[string]any
	assert.NoError(t, yaml.Unmarshal(b, &m))
	assert.Equal(t, got, m["ClusterUUID"])

	// Second call returns the same value (from options in-memory / file).
	got2 := c.ClusterUUID()
	assert.Equal(t, got, got2)
}
