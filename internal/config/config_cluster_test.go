package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/dns"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const shortTestJoinToken = "short-token"

func TestConfig_PortalUrl(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalUrl = ""
		c.options.ClusterDomain = "example.dev"
		assert.Equal(t, "", c.PortalUrl())
		c.options.PortalUrl = DefaultPortalUrl
	})
	t.Run("JoinTokenTooShort", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.JoinToken = shortTestJoinToken
		assert.Equal(t, "", c.JoinToken())
	})
	t.Run("PortalAutoGeneratesJoinToken", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)
		c.options.NodeRole = cluster.RolePortal
		c.options.JoinToken = ""

		token := c.JoinToken()
		assert.NotEmpty(t, token)
		assert.GreaterOrEqual(t, len(token), rnd.JoinTokenLength)
		assert.True(t, rnd.IsJoinToken(token, false))
		assert.True(t, rnd.IsJoinToken(token, true))

		secretFile := filepath.Join(c.PortalConfigPath(), fs.SecretsDir, fs.JoinTokenFile)
		assert.FileExists(t, secretFile)
		info, err := os.Stat(secretFile)
		assert.NoError(t, err)
		assert.Equal(t, fs.ModeSecretFile, info.Mode().Perm())
		assert.Equal(t, token, c.JoinToken())
	})
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalUrl = DefaultPortalUrl
		c.options.ClusterDomain = "foo.bar.baz"
		assert.Equal(t, "https://portal.foo.bar.baz", c.PortalUrl())
	})
	t.Run("SubstitutePhotoPrismClusterDomain", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ClusterDomain = "example.dev"
		// Use curly braces style as found in repo fixtures; resolver normalizes to ${...}.
		c.options.PortalUrl = "https://portal.${PHOTOPRISM_CLUSTER_DOMAIN}"
		assert.Equal(t, "https://portal.example.dev", c.PortalUrl())
		c.options.PortalUrl = DefaultPortalUrl
	})
	t.Run("SubstituteClusterDomain", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ClusterDomain = "example.dev"
		c.options.PortalUrl = "https://portal.${CLUSTER_DOMAIN}"
		assert.Equal(t, "https://portal.example.dev", c.PortalUrl())
		c.options.PortalUrl = DefaultPortalUrl
	})
	t.Run("SubstituteClusterDashDomainCurly", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ClusterDomain = "example.dev"
		// Curly brace variant {cluster-domain} is normalized by ExpandVars.
		c.options.PortalUrl = "https://portal.${cluster-domain}"
		assert.Equal(t, "https://portal.example.dev", c.PortalUrl())
		c.options.PortalUrl = DefaultPortalUrl
	})
	t.Run("LiteralPreserved", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalUrl = "https://portal.example.test"
		c.options.ClusterDomain = "ignored.dev"
		assert.Equal(t, "https://portal.example.test", c.PortalUrl())
		c.options.PortalUrl = DefaultPortalUrl
	})
}

func TestConfig_Cluster(t *testing.T) {
	t.Run("Flags", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Defaults
		assert.False(t, c.Portal())

		// Toggle values
		c.Options().NodeRole = string(cluster.RolePortal)
		assert.True(t, c.Portal())
		c.Options().NodeRole = ""
	})
	t.Run("JWKSUrlSetter", func(t *testing.T) {
		const existing = "https://existing.example/.well-known/jwks.json"
		tests := []struct {
			name   string
			prev   string
			input  string
			expect string
		}{
			{
				name:   "TrimHTTPS",
				prev:   "",
				input:  "  https://portal.example/.well-known/jwks.json  ",
				expect: "https://portal.example/.well-known/jwks.json",
			},
			{
				name:   "CaseInsensitiveScheme",
				prev:   "",
				input:  "HTTPS://portal.example/.well-known/jwks.json",
				expect: "HTTPS://portal.example/.well-known/jwks.json",
			},
			{
				name:   "AllowHTTPOnLocalhost",
				prev:   "",
				input:  "http://localhost:2342/.well-known/jwks.json",
				expect: "http://localhost:2342/.well-known/jwks.json",
			},
			{
				name:   "AllowHTTPOnLoopbackIPv4",
				prev:   "",
				input:  "http://127.0.0.1/.well-known/jwks.json",
				expect: "http://127.0.0.1/.well-known/jwks.json",
			},
			{
				name:   "AllowHTTPOnLoopbackIPv6",
				prev:   "",
				input:  "http://[::1]/.well-known/jwks.json",
				expect: "http://[::1]/.well-known/jwks.json",
			},
			{
				name:   "RejectHTTPNonLoopback",
				prev:   existing,
				input:  "http://portal.example/.well-known/jwks.json",
				expect: existing,
			},
			{
				name:   "RejectUnsupportedScheme",
				prev:   existing,
				input:  "ftp://portal.example/.well-known/jwks.json",
				expect: existing,
			},
			{
				name:   "RejectMalformedURL",
				prev:   existing,
				input:  "://not-a-url",
				expect: existing,
			},
			{
				name:   "ClearValue",
				prev:   existing,
				input:  "",
				expect: "",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				c := NewConfig(CliTestContext())
				c.options.JWKSUrl = tc.prev
				c.SetJWKSUrl(tc.input)
				assert.Equal(t, tc.expect, c.JWKSUrl())
			})
		}
	})
	t.Run("JWTAllowedScopes", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.JWTScope = "cluster vision"
		assert.Equal(t, list.ParseAttr("cluster vision"), c.JWTAllowedScopes())
		c.options.JWTScope = ""
		assert.Equal(t, list.ParseAttr("config cluster vision metrics"), c.JWTAllowedScopes())
	})
	t.Run("Paths", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Use an isolated config path so we don't affect repo storage fixtures.
		tempCfg := t.TempDir()
		c.options.ConfigPath = tempCfg
		c.options.NodeClientSecret = ""
		c.options.PortalUrl = ""
		c.options.JoinToken = ""
		c.options.OptionsYaml = filepath.Join(tempCfg, "options.yml")
		// Clear values potentially loaded at NewConfig creation.
		c.options.NodeClientSecret = ""
		c.options.PortalUrl = ""
		c.options.JoinToken = ""
		c.options.OptionsYaml = filepath.Join(tempCfg, "options.yml")
		// Clear values that may have been loaded from repo fixtures before we
		// isolated the config path.
		c.options.NodeClientSecret = ""
		c.options.PortalUrl = ""
		c.options.JoinToken = ""
		c.options.OptionsYaml = filepath.Join(tempCfg, "options.yml")

		// PortalConfigPath always points to a "cluster" subfolder under ConfigPath.
		expectedCluster := filepath.Join(c.ConfigPath(), fs.PortalDir)
		assert.Equal(t, expectedCluster, c.PortalConfigPath())

		// PortalThemePath falls back to ThemePath if cluster dir does not exist.
		expectedTheme := filepath.Join(c.ConfigPath(), fs.ThemeDir)
		assert.Equal(t, expectedTheme, c.PortalThemePath())

		// When only the cluster directory exists (without a theme subfolder), it still falls back to ThemePath.
		assert.NoError(t, os.MkdirAll(expectedCluster, fs.ModeDir))
		assert.Equal(t, expectedTheme, c.PortalThemePath())

		// When the cluster theme directory exists, PortalThemePath returns it only when app.js is present.
		expectedClusterTheme := filepath.Join(expectedCluster, fs.ThemeDir)
		assert.NoError(t, os.MkdirAll(expectedClusterTheme, fs.ModeDir))
		// Still falls back without app.js.
		assert.Equal(t, expectedTheme, c.PortalThemePath())
		// Create app.js to activate portal-specific theme.
		assert.NoError(t, os.WriteFile(filepath.Join(expectedClusterTheme, fs.AppJsFile), []byte("console.log('theme');\n"), fs.ModeFile))
		assert.Equal(t, expectedClusterTheme, c.PortalThemePath())
	})
	t.Run("PortalAndSecrets", func(t *testing.T) {
		// Isolate config so defaults aren't overridden by repo fixtures: set config-path
		// before creating the Config so NewConfig does not load repository options.yml.
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)

		// Defaults (no options.yml present). Clear the flag default for portal-url
		// so we can assert the derived (unset) behavior.
		c.options.PortalUrl = ""
		assert.Equal(t, "", c.PortalUrl())
		assert.Equal(t, "", c.JoinToken())
		assert.Equal(t, "", c.NodeClientSecret())

		// Set and read back values
		c.options.PortalUrl = "https://portal.example.test"
		c.options.JoinToken = cluster.ExampleJoinToken
		c.options.NodeClientSecret = "node-secret"

		assert.Equal(t, "https://portal.example.test", c.PortalUrl())
		assert.Equal(t, cluster.ExampleJoinToken, c.JoinToken())
		assert.Equal(t, "node-secret", c.NodeClientSecret())
	})
	t.Run("NodePathsAndVersion", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)

		expectedNode := filepath.Join(c.ConfigPath(), fs.NodeDir)
		assert.Equal(t, expectedNode, c.NodeConfigPath())

		expectedTheme := filepath.Join(expectedNode, fs.ThemeDir)
		assert.Equal(t, expectedTheme, c.NodeThemePath())

		// No files yet → empty version.
		assert.Equal(t, "", c.NodeThemeVersion())

		assert.NoError(t, os.MkdirAll(expectedTheme, fs.ModeDir))

		// Version file takes precedence and is sanitized.
		appJsFile := filepath.Join(expectedTheme, fs.AppJsFile)
		assert.NoError(t, os.WriteFile(appJsFile, []byte(`{foo:"bar"}`), fs.ModeFile))
		versionFile := filepath.Join(expectedTheme, fs.VersionTxtFile)
		assert.NoError(t, os.WriteFile(versionFile, []byte(" demo-theme \n"), fs.ModeFile))
		assert.Equal(t, "demo-theme", c.NodeThemeVersion())

		// Removing version file should fall back to app.js modification time.
		assert.NoError(t, os.Remove(versionFile))
		appJS := filepath.Join(expectedTheme, fs.AppJsFile)
		assert.NoError(t, os.WriteFile(appJS, []byte("console.log('theme');\n"), fs.ModeFile))
		modTime := time.Date(2025, 10, 18, 12, 0, 0, 0, time.UTC)
		assert.NoError(t, os.Chtimes(appJS, modTime, modTime))
		assert.Equal(t, modTime.Format(time.RFC3339), c.NodeThemeVersion())
	})
	t.Run("SaveJoinToken", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)
		c.options.NodeRole = cluster.RolePortal

		c.options.JoinToken = "onwnOVt-MZCCkA0z-YJXHnzJ"
		token, tokenFile, err := c.SaveJoinToken("")
		assert.NoError(t, err)
		assert.Empty(t, c.options.JoinToken)
		assert.Equal(t, token, c.JoinToken())
		assert.True(t, rnd.IsJoinToken(token, false))
		assert.FileExists(t, tokenFile)

		data, readErr := os.ReadFile(tokenFile) //nolint:gosec // test reads file from temp directory
		assert.NoError(t, readErr)
		assert.Equal(t, token, strings.TrimSpace(string(data)))
	})
	t.Run("SaveNodeClientSecret", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)

		fileName, err := c.SaveNodeClientSecret(cluster.ExampleClientSecret)
		assert.NoError(t, err)
		assert.FileExists(t, fileName)

		data, readErr := os.ReadFile(fileName) //nolint:gosec // test reads file from temp directory
		assert.NoError(t, readErr)
		assert.Equal(t, cluster.ExampleClientSecret, strings.TrimSpace(string(data)))
	})
	t.Run("NodeClientSecretFromFile", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)

		// Persist secret to node config path.
		_, err := c.SaveNodeClientSecret(cluster.ExampleClientSecret)
		assert.NoError(t, err)

		// Simulate a fresh process reading from disk.
		ctx2 := CliTestContext()
		assert.NoError(t, ctx2.Set("config-path", tempCfg))
		c2 := NewConfig(ctx2)
		c2.options.NodeClientSecret = "" // ensure it must read the file
		assert.Equal(t, cluster.ExampleClientSecret, c2.NodeClientSecret())
	})
	t.Run("NodeClientSecretEnvOverride", func(t *testing.T) {
		secretFile := filepath.Join(t.TempDir(), "client_secret")
		assert.NoError(t, os.WriteFile(secretFile, []byte(cluster.ExampleClientSecret), fs.ModeSecretFile))
		t.Setenv(FlagFileVar("NODE_CLIENT_SECRET"), secretFile)

		c := NewConfig(CliTestContext())
		c.options.NodeClientSecret = ""
		assert.Equal(t, cluster.ExampleClientSecret, c.NodeClientSecret())
	})
	t.Run("NodeClientSecretFallbackOnWrite", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)

		secretDir := filepath.Join(c.NodeConfigPath(), fs.SecretsDir)
		assert.NoError(t, os.MkdirAll(secretDir, fs.ModeDir))
		assert.NoError(t, os.Chmod(secretDir, 0o500)) //nolint:gosec // making directory intentionally non-writable for fallback test

		_, err := c.SaveNodeClientSecret(cluster.ExampleClientSecret)
		assert.Error(t, err)
		assert.Equal(t, cluster.ExampleClientSecret, c.NodeClientSecret())
	})
	t.Run("JoinTokenFilePortal", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)
		c.options.NodeRole = cluster.RolePortal

		expected := filepath.Join(c.PortalConfigPath(), fs.SecretsDir, fs.JoinTokenFile)
		assert.Equal(t, expected, c.JoinTokenFile())
		assert.Equal(t, expected, c.PortalJoinTokenFile())
	})
	t.Run("JoinTokenFileInstance", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)
		c.options.NodeRole = cluster.RoleApp

		expected := filepath.Join(c.NodeConfigPath(), fs.SecretsDir, fs.JoinTokenFile)
		assert.Equal(t, expected, c.JoinTokenFile())
		assert.Equal(t, expected, c.NodeJoinTokenFile())
	})
	t.Run("SaveJoinTokenFallbackOnWrite", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)

		secretDir := filepath.Join(c.NodeConfigPath(), fs.SecretsDir)
		assert.NoError(t, os.MkdirAll(secretDir, fs.ModeDir))
		assert.NoError(t, os.Chmod(secretDir, 0o500)) //nolint:gosec // making directory intentionally non-writable for fallback test

		_, _, err := c.SaveJoinToken("")
		assert.Error(t, err)
		token := c.JoinToken()
		assert.True(t, rnd.IsJoinToken(token, false))
	})
	t.Run("NodeClientSecretFile", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)

		expected := filepath.Join(c.NodeConfigPath(), fs.SecretsDir, fs.ClientSecretFile)
		assert.Equal(t, expected, c.NodeClientSecretFile())
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
		assert.NoError(t, os.MkdirAll(clusterTheme, fs.ModeDir))
		assert.True(t, filepath.IsAbs(c.PortalThemePath()))
	})
	t.Run("NodeName", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.SiteUrl = "https://app.localssl.dev"
		h, d, found := c.deriveNodeNameAndDomainFromHttpHost()
		assert.Equal(t, "app", h)
		assert.Equal(t, "localssl.dev", d)
		assert.True(t, found)
		c.options.NodeName = " Client Credentials幸"
		assert.Equal(t, "client-credentials", c.NodeName())
		c.options.NodeName = ""
		// With defaults, NodeName derives from hostname or falls back to a stable identifier.
		got := c.NodeName()
		assert.NotEmpty(t, got)
		assert.Equal(t, "app", h)
		assert.Equal(t, "localssl.dev", d)
		// Must be DNS label compatible (lowercase [a-z0-9-], 1–32, start/end alnum).
		assert.Regexp(t, `^[a-z0-9](?:[a-z0-9-]{0,30}[a-z0-9])?$`, got)
	})
	t.Run("NodeNameNormalization", func(t *testing.T) {
		orig := dns.GetHostname
		dns.GetHostname = func() (string, error) { return "", nil }
		t.Cleanup(func() { dns.GetHostname = orig })

		c := NewConfig(CliTestContext())
		c.options.NodeName = " My.Host/Name:Prod "
		assert.Equal(t, "my-host-name-prod", c.NodeName())

		c.options.NodeName = "-._a--"
		assert.Equal(t, "a", c.NodeName())

		c.options.NodeName = strings.Repeat("a", 40)
		assert.Equal(t, strings.Repeat("a", 32), c.NodeName())
	})
	t.Run("NodeNameFromHostname", func(t *testing.T) {
		orig := dns.GetHostname
		dns.GetHostname = func() (string, error) { return "My.Host/Name:Prod", nil }
		t.Cleanup(func() { dns.GetHostname = orig })

		c := NewConfig(CliTestContext())
		c.options.NodeName = ""
		assert.Equal(t, "my-host-name-prod", c.NodeName())
	})
	t.Run("NodeRoleValues", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Default / unknown → node
		c.options.NodeRole = ""
		assert.Equal(t, string(cluster.RoleApp), c.NodeRole())
		c.options.NodeRole = "unknown"
		assert.Equal(t, string(cluster.RoleApp), c.NodeRole())

		// Explicit values
		c.options.NodeRole = string(cluster.RoleApp)
		assert.Equal(t, string(cluster.RoleApp), c.NodeRole())
		c.options.NodeRole = string(cluster.RolePortal)
		assert.Equal(t, string(cluster.RolePortal), c.NodeRole())
		c.options.NodeRole = string(cluster.RoleService)
		assert.Equal(t, string(cluster.RoleService), c.NodeRole())
	})
	t.Run("SecretsFromFiles", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Create temp secret/token files.
		dir := t.TempDir()
		nsFile := filepath.Join(dir, "node_client_secret")
		tkFile := filepath.Join(dir, "portal_token")
		assert.NoError(t, os.WriteFile(nsFile, []byte(cluster.ExampleClientSecret), fs.ModeSecretFile))
		assert.NoError(t, os.WriteFile(tkFile, []byte(cluster.ExampleJoinTokenAlt), fs.ModeSecretFile))

		// Clear inline values so file-based lookup is used.
		c.options.NodeClientSecret = ""
		c.options.JoinToken = ""

		// Point env vars at the files and verify.
		t.Setenv("PHOTOPRISM_NODE_CLIENT_SECRET_FILE", nsFile)
		t.Setenv("PHOTOPRISM_JOIN_TOKEN_FILE", tkFile)
		assert.Equal(t, cluster.ExampleClientSecret, c.NodeClientSecret())
		assert.Equal(t, cluster.ExampleJoinTokenAlt, c.JoinToken())

		// Refreshing the token file should invalidate the cache.
		time.Sleep(5 * time.Millisecond)
		newToken := cluster.ExampleJoinToken
		assert.NoError(t, os.WriteFile(tkFile, []byte(newToken), fs.ModeSecretFile))
		c.clearJoinTokenFileCache()
		assert.Equal(t, newToken, c.JoinToken())

		// Empty / missing should yield empty strings.
		t.Setenv("PHOTOPRISM_NODE_CLIENT_SECRET_FILE", filepath.Join(dir, "missing"))
		t.Setenv("PHOTOPRISM_JOIN_TOKEN_FILE", filepath.Join(dir, "missing"))
		c.options.NodeClientSecret = ""
		c.options.JoinToken = ""
		c.clearJoinTokenFileCache()
		assert.Equal(t, "", c.NodeClientSecret())
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
	assert.NoError(t, os.WriteFile(filepath.Join(tempCfg, "options.yml"), b, fs.ModeFile))

	// Set env; file value must win for consistency with other options.
	t.Setenv("PHOTOPRISM_CLUSTER_UUID", "22222222-2222-4222-8222-222222222222")
	// Load options.yml into options struct (we updated ConfigPath after creation).
	assert.NoError(t, c.options.Load(c.OptionsYaml()))
	got := c.ClusterUUID()
	assert.Equal(t, "11111111-1111-4111-8111-111111111111", got)
}

func TestConfig_ClusterUUID_FromOptions(t *testing.T) {
	c := NewConfig(CliTestContext())
	optionsOriginal := c.OptionsYaml()
	tempCfg := t.TempDir()

	if err := fs.MkdirAll(tempCfg); err != nil {
		t.Fatal(err)
	}

	c.options.ConfigPath = tempCfg
	optionsYaml := filepath.Join(tempCfg, "options.yml")
	c.options.OptionsYaml = optionsYaml

	opts := map[string]any{"ClusterUUID": "33333333-3333-4333-8333-333333333333"}
	b, _ := yaml.Marshal(opts)
	assert.NoError(t, os.WriteFile(optionsYaml, b, fs.ModeFile))

	// Ensure env is not set.
	t.Setenv("PHOTOPRISM_CLUSTER_UUID", "")

	// Load options.yml into options struct (we updated ConfigPath after creation).
	assert.NoError(t, c.options.Load(optionsYaml))
	// Access the value via getter.
	got := c.ClusterUUID()
	assert.Equal(t, "33333333-3333-4333-8333-333333333333", got)
	c.options.OptionsYaml = optionsOriginal
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
	optionsOriginal := c.OptionsYaml()

	tempCfg := t.TempDir()

	if err := fs.MkdirAll(tempCfg); err != nil {
		t.Fatal(err)
	}

	c.options.ConfigPath = tempCfg
	optionsYaml := filepath.Join(tempCfg, "options.yml")
	c.options.OptionsYaml = optionsYaml

	// No env, no options.yml → should generate and persist.
	t.Setenv("PHOTOPRISM_CLUSTER_UUID", "")

	if err := c.SaveClusterUUID(rnd.UUID()); err != nil {
		t.Fatal(err)
	}

	got := c.ClusterUUID()
	if !rnd.IsUUID(got) {
		t.Fatalf("expected a UUIDv4, got %q", got)
	}

	// Verify content persisted to options.yml.
	b, err := os.ReadFile(optionsYaml) //nolint:gosec // test reads generated options file
	assert.NoError(t, err)
	var m map[string]any
	assert.NoError(t, yaml.Unmarshal(b, &m))
	assert.Equal(t, got, m["ClusterUUID"])

	// Second call returns the same value (from options in-memory / file).
	got2 := c.ClusterUUID()
	assert.Equal(t, got, got2)

	c.options.OptionsYaml = optionsOriginal
}
