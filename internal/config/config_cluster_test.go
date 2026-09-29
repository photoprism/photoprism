package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/dns"
	"github.com/photoprism/photoprism/pkg/http/proxy"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const shortTestJoinToken = "short-token"

func TestConfig_ClusterAllowGroups(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Nil(t, c.ClusterAllowGroups())
	})
	t.Run("NormalizesAndSplits", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ClusterAllowGroups = []string{"Media-Acme-Admin, Media-Acme-Viewer", "media-acme-admin"}
		assert.Equal(t, []string{"media-acme-admin", "media-acme-viewer"}, c.ClusterAllowGroups())
	})
	t.Run("IncludesRoleMapKeys", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ClusterAllowGroupRoles = []string{"Media-Acme-Admin=admin"}
		assert.Equal(t, []string{"media-acme-admin"}, c.ClusterAllowGroups(),
			"a role mapping alone must admit its groups")
	})
}

func TestConfig_ClusterAllowGroupRoles(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Nil(t, c.ClusterAllowGroupRoles())
	})
	t.Run("AcceptsAllInstanceRoles", func(t *testing.T) {
		// Roles must validate independently of the registered edition role
		// table, since this is read during early bootstrap before activation.
		c := NewConfig(CliTestContext())
		c.options.ClusterAllowGroupRoles = []string{
			"g-admin=admin", "g-manager=manager", "g-user=user",
			"g-contributor=contributor", "g-viewer:viewer", "g-guest=guest",
		}
		assert.Equal(t, map[string]string{
			"g-admin": "admin", "g-manager": "manager", "g-user": "user",
			"g-contributor": "contributor", "g-viewer": "viewer", "g-guest": "guest",
		}, c.ClusterAllowGroupRoles())
	})
	t.Run("ToleratesWhitespaceSeparatedPairs", func(t *testing.T) {
		// A StringSlice env splits on commas, so a space-separated value arrives
		// as a single element; it must still resolve to all its pairs.
		c := NewConfig(CliTestContext())
		c.options.ClusterAllowGroupRoles = []string{"g-admin=admin g-viewer=viewer"}
		assert.Equal(t, map[string]string{"g-admin": "admin", "g-viewer": "viewer"}, c.ClusterAllowGroupRoles())
	})
	t.Run("DropsInvalidEntries", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ClusterAllowGroupRoles = []string{"a=cluster_admin", "b=visitor", "c=bogus", "=admin", "noseparator", ""}
		assert.Nil(t, c.ClusterAllowGroupRoles(), "non-instance roles and malformed pairs must be dropped")
	})
}

func TestConfig_ClusterGroupsFullView(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.False(t, c.ClusterGroupsFullView())
	c.options.ClusterGroupsFullView = true
	assert.True(t, c.ClusterGroupsFullView())
}

func TestSplitGroupList(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, splitGroupList("a,b c"))
	assert.Empty(t, splitGroupList("  ,  "))
}

func TestReportGroupRoles(t *testing.T) {
	assert.Equal(t, "", reportGroupRoles(nil))
	assert.Equal(t, "a=admin, b=guest", reportGroupRoles(map[string]string{"b": "guest", "a": "admin"}))
}

func TestConfig_PortalLoginUrl(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Equal(t, "", c.PortalLoginUrl())
	})
	t.Run("SetterValidates", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.SetPortalLoginUrl("  https://portal.example.com/portal/login  ")
		assert.Equal(t, "https://portal.example.com/portal/login", c.PortalLoginUrl())
		c.SetPortalLoginUrl("http://127.0.0.1:2342/portal/login")
		assert.Equal(t, "http://127.0.0.1:2342/portal/login", c.PortalLoginUrl(), "http loopback must be allowed")
		c.SetPortalLoginUrl("http://portal.example.com/portal/login")
		assert.Equal(t, "http://127.0.0.1:2342/portal/login", c.PortalLoginUrl(), "http non-loopback must be rejected")
		c.SetPortalLoginUrl("ftp://portal.example.com/login")
		assert.Equal(t, "http://127.0.0.1:2342/portal/login", c.PortalLoginUrl(), "unsupported schemes must be rejected")
		c.SetPortalLoginUrl("")
		assert.Equal(t, "", c.PortalLoginUrl(), "an empty value must clear the URL")
	})
	t.Run("GetterRejectsStoredInvalidValues", func(t *testing.T) {
		// Values can bypass the setter (env, flag, hand-edited options.yml, or a
		// stale persisted entry) — the getter must never hand them to the browser.
		c := NewConfig(CliTestContext())
		for _, v := range []string{"javascript:alert(1)", "http://portal.example.com/login", "://nope", "ftp://x/login"} {
			c.options.PortalLoginUrl = v
			assert.Equal(t, "", c.PortalLoginUrl(), "stored value %q must be rejected on read", v)
		}
		c.options.PortalLoginUrl = "https://portal.example.com/portal/login"
		assert.Equal(t, "https://portal.example.com/portal/login", c.PortalLoginUrl())
	})
}

func TestValidClusterURL(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for _, v := range []string{"", "https://portal.example.com/login", "http://127.0.0.1:2342/login", "http://localhost/login"} {
			_, ok := validClusterURL(v)
			assert.True(t, ok, "%q must be valid", v)
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, v := range []string{"http://portal.example.com/login", "javascript:alert(1)", "://nope", "/relative/login"} {
			_, ok := validClusterURL(v)
			assert.False(t, ok, "%q must be invalid", v)
		}
	})
	t.Run("Trims", func(t *testing.T) {
		v, ok := validClusterURL("  https://a.example.com/  ")
		assert.True(t, ok)
		assert.Equal(t, "https://a.example.com/", v)
	})
}

func TestConfig_PortalOIDCIssuer(t *testing.T) {
	t.Run("DefaultsToSiteUrl", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.Equal(t, c.SiteUrl(), c.PortalOIDCIssuer())
	})
	t.Run("FollowsSiteUrlOverride", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.SiteUrl = "https://portal.example.com/"
		assert.Equal(t, "https://portal.example.com/", c.PortalOIDCIssuer())
	})
	t.Run("ExplicitOverrideWins", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.SiteUrl = "https://site.example.com/"
		c.options.PortalOIDCIssuer = "https://portal-issuer.example.com/"
		assert.Equal(t, "https://portal-issuer.example.com/", c.PortalOIDCIssuer())
	})
}

func TestConfig_PortalOIDCTTL(t *testing.T) {
	t.Run("DefaultsToFiveMinutes", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalOIDCTTL = 0
		assert.Equal(t, 300*time.Second, c.PortalOIDCTTL())
	})
	t.Run("ClampsToMin", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalOIDCTTL = 10
		assert.Equal(t, 60*time.Second, c.PortalOIDCTTL())
	})
	t.Run("ClampsToMax", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalOIDCTTL = 99999
		assert.Equal(t, 900*time.Second, c.PortalOIDCTTL())
	})
	t.Run("HonorsExplicit", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalOIDCTTL = 600
		assert.Equal(t, 600*time.Second, c.PortalOIDCTTL())
	})
}

func TestConfig_PortalOIDCCodeTTL(t *testing.T) {
	t.Run("DefaultsToSixtySeconds", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalOIDCCodeTTL = 0
		assert.Equal(t, 60*time.Second, c.PortalOIDCCodeTTL())
	})
	t.Run("ClampsToMin", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalOIDCCodeTTL = 5
		assert.Equal(t, 30*time.Second, c.PortalOIDCCodeTTL())
	})
	t.Run("ClampsToMax", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalOIDCCodeTTL = 1000
		assert.Equal(t, 300*time.Second, c.PortalOIDCCodeTTL())
	})
}

func TestConfig_PortalOIDCDefaultPolicyChooser(t *testing.T) {
	t.Run("DefaultIsChooser", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalOIDCDefaultPolicy = ""
		assert.True(t, c.PortalOIDCDefaultPolicyChooser())
	})
	t.Run("Direct", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalOIDCDefaultPolicy = "direct"
		assert.False(t, c.PortalOIDCDefaultPolicyChooser())
	})
	t.Run("UnknownValueDefaultsToChooser", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalOIDCDefaultPolicy = "garbage"
		assert.True(t, c.PortalOIDCDefaultPolicyChooser())
	})
}

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
		c.options.Edition = Portal
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
		if err == nil {
			assert.Equal(t, fs.ModeSecretFile, info.Mode().Perm())
		}
		assert.Equal(t, token, c.JoinToken())
	})
	t.Run("RegularInstallCannotEnablePortalRole", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.Edition = Community
		c.options.NodeRole = cluster.RolePortal
		c.options.JoinToken = ""

		assert.Equal(t, string(cluster.RoleInstance), c.NodeRole())
		assert.False(t, c.Portal())
		assert.Equal(t, "", c.JoinToken())
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
	t.Run("OptionUnchanged", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.PortalUrl = "https://portal.${PHOTOPRISM_CLUSTER_DOMAIN}"
		c.options.ClusterDomain = "example.dev"
		assert.Equal(t, "https://portal.example.dev", c.PortalUrl())
		assert.Equal(t, "https://portal.${PHOTOPRISM_CLUSTER_DOMAIN}", c.options.PortalUrl, "the configured value must keep its variables")
		c.options.ClusterDomain = "example.org"
		assert.Equal(t, "https://portal.example.org", c.PortalUrl())
		c.options.PortalUrl = DefaultPortalUrl
	})
	t.Run("Concurrent", func(t *testing.T) {
		// Concurrent calls must not write the options, which only a run with -race detects.
		c := NewConfig(CliTestContext())
		c.options.PortalUrl = DefaultPortalUrl
		c.options.ClusterDomain = "example.dev"
		var wg sync.WaitGroup
		urls := make([]string, 4)
		for i := range urls {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				urls[i] = c.PortalUrl()
			}(i)
		}
		wg.Wait()
		for _, url := range urls {
			assert.Equal(t, "https://portal.example.dev", url)
		}
		assert.Equal(t, DefaultPortalUrl, c.options.PortalUrl)
	})
}

func TestConfig_Cluster(t *testing.T) {
	t.Run("Flags", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		// Defaults
		assert.False(t, c.Portal())

		// Regular installations cannot enable portal mode through the role flag.
		c.Options().NodeRole = string(cluster.RolePortal)
		assert.False(t, c.Portal())

		// Portal edition is always treated as a portal node.
		c.Options().Edition = Portal
		assert.True(t, c.Portal())
		c.Options().NodeRole = ""
	})
	t.Run("PortalProxy", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		c.options.PortalProxy = true
		assert.False(t, c.PortalProxy())

		c.options.NodeRole = string(cluster.RolePortal)
		assert.False(t, c.PortalProxy())

		c.options.Edition = Portal
		assert.True(t, c.PortalProxy())

		c.options.PortalProxy = false
		assert.False(t, c.PortalProxy())
	})
	t.Run("PortalProxyUri", func(t *testing.T) {
		c := NewConfig(CliTestContext())

		assert.Equal(t, proxy.DefaultPathPrefix, c.PortalProxyUri())

		c.options.PortalProxyUri = "/instance"
		assert.Equal(t, "/instance", c.PortalProxyUri())

		c.options.PortalProxyUri = "https://proxy.example.com/instance/"
		assert.Equal(t, "https://proxy.example.com/instance/", c.PortalProxyUri())
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
		assert.Equal(t, list.ParseAttr("config cluster vision metrics mcp users"), c.JWTAllowedScopes())
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
		c.options.Edition = Portal
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
	t.Run("NodeClientSecretPrefersFileOverInline", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)

		_, err := c.SaveNodeClientSecret(cluster.ExampleClientSecret)
		assert.NoError(t, err)

		// Inline value should not override the persisted secret file.
		c.options.NodeClientSecret = "stale-inline-secret"
		assert.Equal(t, cluster.ExampleClientSecret, c.NodeClientSecret())
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
	t.Run("SaveClusterOptionsUpdate", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)
		c.options.ConfigPath = tempCfg
		c.options.OptionsYaml = filepath.Join(tempCfg, "options.yml")

		seed := map[string]any{
			"Existing":         "value",
			"NodeClientSecret": "legacy-inline-secret",
		}
		b, err := yaml.Marshal(seed)
		assert.NoError(t, err)
		assert.NoError(t, os.WriteFile(c.OptionsYaml(), b, fs.ModeFile))

		update := cluster.OptionsUpdate{}
		update.SetClusterUUID("4a47c940-d5de-41b3-88a2-eb816cc659ca")
		update.SetNodeClientID(cluster.ExampleClientID)
		update.SetDatabaseName("cluster_database")
		update.SetDatabaseUser("cluster_user")

		wrote, err := c.SaveClusterOptionsUpdate(update)
		assert.NoError(t, err)
		assert.True(t, wrote)

		content, readErr := os.ReadFile(c.OptionsYaml())
		assert.NoError(t, readErr)

		var merged map[string]any
		assert.NoError(t, yaml.Unmarshal(content, &merged))
		assert.Equal(t, "value", merged["Existing"])
		assert.Equal(t, "legacy-inline-secret", merged["NodeClientSecret"])
		assert.Equal(t, "4a47c940-d5de-41b3-88a2-eb816cc659ca", merged["ClusterUUID"])
		assert.Equal(t, cluster.ExampleClientID, merged["NodeClientID"])
		assert.Equal(t, "cluster_database", merged["DatabaseName"])
		assert.Equal(t, "cluster_user", merged["DatabaseUser"])

		// Applying the same values again should not rewrite.
		wrote, err = c.SaveClusterOptionsUpdate(update)
		assert.NoError(t, err)
		assert.False(t, wrote)
	})
	t.Run("SaveClusterOptionsUpdateInvalidUUID", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ConfigPath = t.TempDir()
		c.options.OptionsYaml = filepath.Join(c.options.ConfigPath, "options.yml")

		update := cluster.OptionsUpdate{}
		update.SetClusterUUID("invalid-uuid")
		wrote, err := c.SaveClusterOptionsUpdate(update)
		assert.Error(t, err)
		assert.False(t, wrote)
	})
	t.Run("SaveClusterOptionsUpdateInvalidDatabase", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ConfigPath = t.TempDir()
		c.options.OptionsYaml = filepath.Join(c.options.ConfigPath, "options.yml")

		update := cluster.OptionsUpdate{}
		update.SetDatabaseName("--option=value")
		wrote, err := c.SaveClusterOptionsUpdate(update)
		assert.EqualError(t, err, "invalid database name")
		assert.False(t, wrote)
		assert.NoFileExists(t, c.OptionsYaml())
	})
	t.Run("SaveClusterOptionsUpdateDSNParams", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ConfigPath = t.TempDir()
		c.options.OptionsYaml = filepath.Join(c.options.ConfigPath, "options.yml")

		update := cluster.OptionsUpdate{}
		update.SetDatabaseDriver("mysql")
		update.SetDatabaseDSN("cluster_u0123456789a:secret@tcp(mariadb:4001)/cluster_d0123456789a?option=value")
		wrote, err := c.SaveClusterOptionsUpdate(update)
		require.NoError(t, err)
		assert.True(t, wrote)

		content, err := os.ReadFile(c.OptionsYaml())
		require.NoError(t, err)

		var merged map[string]any
		require.NoError(t, yaml.Unmarshal(content, &merged))
		assert.Equal(t, fmt.Sprintf("cluster_u0123456789a:secret@tcp(mariadb:4001)/cluster_d0123456789a?%s&timeout=%ds",
			dsn.Params[dsn.DriverMySQL], c.DatabaseTimeout()), merged["DatabaseDSN"])
	})
	t.Run("SaveOptionsPatch", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)
		c.options.ConfigPath = tempCfg
		c.options.OptionsYaml = filepath.Join(tempCfg, "options.yml")

		seed := map[string]any{
			"Existing": "value",
		}
		b, err := yaml.Marshal(seed)
		assert.NoError(t, err)
		assert.NoError(t, os.WriteFile(c.OptionsYaml(), b, fs.ModeFile))

		patch := Values{
			"SiteUrl": "https://photos.example.com/",
			"Public":  true,
		}

		wrote, err := c.SaveOptionsPatch(patch)
		assert.NoError(t, err)
		assert.True(t, wrote)

		content, readErr := os.ReadFile(c.OptionsYaml())
		assert.NoError(t, readErr)

		var merged map[string]any
		assert.NoError(t, yaml.Unmarshal(content, &merged))
		assert.Equal(t, "value", merged["Existing"])
		assert.Equal(t, "https://photos.example.com/", merged["SiteUrl"])
		assert.Equal(t, true, merged["Public"])

		wrote, err = c.SaveOptionsPatch(patch)
		assert.NoError(t, err)
		assert.False(t, wrote)
	})
	t.Run("SaveOptionsPatchEmpty", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ConfigPath = t.TempDir()
		c.options.OptionsYaml = filepath.Join(c.options.ConfigPath, "options.yml")

		wrote, err := c.SaveOptionsPatch(nil)
		assert.NoError(t, err)
		assert.False(t, wrote)
	})
	t.Run("JoinTokenFilePortal", func(t *testing.T) {
		tempCfg := t.TempDir()
		ctx := CliTestContext()
		assert.NoError(t, ctx.Set("config-path", tempCfg))
		c := NewConfig(ctx)
		c.options.Edition = Portal
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
		c.options.NodeRole = cluster.RoleInstance

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
	t.Run("NodeNameNotOverriddenByClusterDomainDerivation", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.NodeName = "smiling-ocelot"
		c.options.ClusterDomain = ""
		c.options.SiteUrl = "https://media.glowworm.com/i/smiling-ocelot/"

		assert.Equal(t, "glowworm.com", c.ClusterDomain())
		assert.Equal(t, "smiling-ocelot", c.NodeName())
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
		c.options.NodeRole = "app"
		assert.Equal(t, string(cluster.RoleInstance), c.NodeRole())
		c.options.NodeRole = string(cluster.RolePortal)
		assert.Equal(t, string(cluster.RoleInstance), c.NodeRole())
		c.options.NodeRole = string(cluster.RoleService)
		assert.Equal(t, string(cluster.RoleService), c.NodeRole())

		// Portal edition always resolves to portal.
		c.options.Edition = Portal
		c.options.NodeRole = string(cluster.RoleInstance)
		assert.Equal(t, string(cluster.RolePortal), c.NodeRole())
		c.options.NodeRole = string(cluster.RoleService)
		assert.Equal(t, string(cluster.RolePortal), c.NodeRole())
		c.options.NodeRole = string(cluster.RolePortal)
		assert.Equal(t, string(cluster.RolePortal), c.NodeRole())
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

func TestConfig_JWTRotateDays(t *testing.T) {
	c := NewConfig(CliTestContext())
	original := c.options.JWTRotateDays

	t.Cleanup(func() {
		c.options.JWTRotateDays = original
	})

	t.Run("Default", func(t *testing.T) {
		c.options.JWTRotateDays = 90
		assert.Equal(t, 90, c.JWTRotateDays())
	})
	t.Run("Custom", func(t *testing.T) {
		c.options.JWTRotateDays = 30
		assert.Equal(t, 30, c.JWTRotateDays())
	})
	t.Run("Disabled", func(t *testing.T) {
		// 0 means the operator asked for manual rotation, so it must not read as unset.
		c.options.JWTRotateDays = 0
		assert.Equal(t, 0, c.JWTRotateDays())
	})
	t.Run("Negative", func(t *testing.T) {
		c.options.JWTRotateDays = -7
		assert.Equal(t, 0, c.JWTRotateDays())
	})
}

// TestValidateClusterOptionsUpdate checks the values a cluster options update may carry.
func TestValidateClusterOptionsUpdate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		update := cluster.OptionsUpdate{}
		update.SetClusterUUID("4a47c940-d5de-41b3-88a2-eb816cc659ca")
		update.SetDatabaseDriver("mysql")
		update.SetDatabaseName("cluster_d0123456789a")
		update.SetDatabaseUser("cluster_u0123456789a")
		update.SetDatabaseServer("mariadb:4001")
		update.SetDatabaseDSN("cluster_u0123456789a:secret@tcp(mariadb:4001)/cluster_d0123456789a?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true")
		assert.NoError(t, validateClusterOptionsUpdate(update))
		update.SetDatabaseDriver("MariaDB")
		update.SetDatabaseName("photo-prism_2")
		update.SetDatabaseServer("")
		assert.NoError(t, validateClusterOptionsUpdate(update))
		assert.NoError(t, validateClusterOptionsUpdate(cluster.OptionsUpdate{}))
	})
	t.Run("InvalidDriver", func(t *testing.T) {
		for _, driver := range []string{"", "sqlite", "postgres", "x"} {
			update := cluster.OptionsUpdate{}
			update.SetDatabaseDriver(driver)
			assert.EqualError(t, validateClusterOptionsUpdate(update), "invalid database driver", "driver %q", driver)
		}
	})
	t.Run("InvalidDatabase", func(t *testing.T) {
		for _, tc := range []struct {
			set func(*cluster.OptionsUpdate, string)
			err string
		}{
			{(*cluster.OptionsUpdate).SetDatabaseName, "invalid database name"},
			{(*cluster.OptionsUpdate).SetDatabaseUser, "invalid database user"},
			{(*cluster.OptionsUpdate).SetDatabaseServer, "invalid database server"},
		} {
			for _, value := range []string{"-x", "--option=value", " -x", "db?option=value", "db/x", "db@x"} {
				update := cluster.OptionsUpdate{}
				tc.set(&update, value)
				assert.EqualError(t, validateClusterOptionsUpdate(update), tc.err, "value %q", value)
			}
		}
	})
	t.Run("InvalidDSN", func(t *testing.T) {
		for dsnValue, err := range map[string]string{
			"user:secret@tcp(mariadb:4001)/--option=value?parseTime=true": "invalid database name",
			"-x:secret@tcp(mariadb:4001)/photoprism":                      "invalid database user",
			"user:secret@tcp(--option=value)/photoprism":                  "invalid database server",
			"user:secret@unix(/run/mysqld/mysqld.sock)/photoprism":        "invalid database dsn",
			"/srv/photoprism/index.db":                                    "invalid database dsn",
			"-x":                                                          "invalid database dsn",
		} {
			update := cluster.OptionsUpdate{}
			update.SetDatabaseDSN(dsnValue)
			assert.EqualError(t, validateClusterOptionsUpdate(update), err, "dsn %q", dsnValue)
		}
	})
	t.Run("InvalidUUID", func(t *testing.T) {
		update := cluster.OptionsUpdate{}
		update.SetNodeUUID("invalid-uuid")
		assert.EqualError(t, validateClusterOptionsUpdate(update), "invalid node UUID")
	})
}

// TestValidateClusterDatabase checks the database name, user, and server rules.
func TestValidateClusterDatabase(t *testing.T) {
	str := func(s string) *string { return &s }
	t.Run("Success", func(t *testing.T) {
		assert.NoError(t, validateClusterDatabase(str("cluster_d0123456789a"), str("cluster_u0123456789a"), str("mariadb:4001")))
		assert.NoError(t, validateClusterDatabase(str(""), str(""), str("")))
		assert.NoError(t, validateClusterDatabase(nil, nil, nil))
	})
	t.Run("InvalidName", func(t *testing.T) {
		assert.EqualError(t, validateClusterDatabase(str(strings.Repeat("a", 65)), nil, nil), "invalid database name")
		assert.EqualError(t, validateClusterDatabase(str("db?x"), nil, nil), "invalid database name")
	})
	t.Run("InvalidUser", func(t *testing.T) {
		assert.EqualError(t, validateClusterDatabase(nil, str("user:x"), nil), "invalid database user")
	})
	t.Run("InvalidServer", func(t *testing.T) {
		assert.EqualError(t, validateClusterDatabase(nil, nil, str("mariadb:4001)/x")), "invalid database server")
	})
}

// TestClusterDatabaseDSN checks that a Portal DSN is validated and rebuilt with accepted parameters.
func TestClusterDatabaseDSN(t *testing.T) {
	defaults := fmt.Sprint(dsn.Params[dsn.DriverMySQL])
	t.Run("Provisioned", func(t *testing.T) {
		provisioned := "cluster_u0123456789a:Sup3rSecret@tcp(mariadb:4001)/cluster_d0123456789a?" + defaults
		result, dropped, err := clusterDatabaseDSN(provisioned, 15)
		require.NoError(t, err)
		assert.Equal(t, provisioned+"&timeout=15s", result)
		assert.Empty(t, dropped)
	})
	t.Run("AcceptedParams", func(t *testing.T) {
		result, dropped, err := clusterDatabaseDSN("user:secret@tcp(mariadb:4001)/photoprism?tls=skip-verify&charset=utf8mb4,utf8&option=value&collation=utf8mb4_unicode_ci&parseTime=true&timeout=60s", 15)
		require.NoError(t, err)
		assert.Equal(t, "user:secret@tcp(mariadb:4001)/photoprism?tls=skip-verify&charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=60s", result)
		assert.Equal(t, []string{"option"}, dropped)
	})
	t.Run("ProxyAndTLS", func(t *testing.T) {
		for _, query := range []string{"tls=true", "tls=skip-verify", "tls=preferred", "tls=false", "tls=1", "interpolateParams=true",
			"rejectReadOnly=1", "maxAllowedPacket=67108864", "maxAllowedPacket=1073741824", "readTimeout=30s&writeTimeout=30s"} {
			result, dropped, err := clusterDatabaseDSN("user:secret@tcp(mariadb:4001)/photoprism?"+query, 15)
			require.NoError(t, err, query)
			assert.Equal(t, "user:secret@tcp(mariadb:4001)/photoprism?"+query+"&"+defaults+"&timeout=15s", result, query)
			assert.Empty(t, dropped, query)
		}
	})
	t.Run("DefaultsAdded", func(t *testing.T) {
		result, dropped, err := clusterDatabaseDSN("user:secret@tcp([::1]:3306)/photoprism?option=value&other=1", 15)
		require.NoError(t, err)
		assert.Equal(t, "user:secret@tcp([::1]:3306)/photoprism?"+defaults+"&timeout=15s", result)
		assert.Equal(t, []string{"option", "other"}, dropped)
		result, _, err = clusterDatabaseDSN("user:secret@tcp(mariadb:4001)/db", 30)
		require.NoError(t, err)
		assert.Equal(t, "user:secret@tcp(mariadb:4001)/db?"+defaults+"&timeout=30s", result)
	})
	t.Run("CharsetPair", func(t *testing.T) {
		for query, want := range map[string]string{
			"charset=utf8":                   "charset=utf8&parseTime=true&timeout=15s",
			"charset=utf8mb3&parseTime=1":    "charset=utf8mb3&parseTime=1&timeout=15s",
			"collation=utf8mb3_general_ci":   "collation=utf8mb3_general_ci&parseTime=true&timeout=15s",
			"charset=utf8mb4&parseTime=true": "charset=utf8mb4&parseTime=true&timeout=15s",
		} {
			result, _, err := clusterDatabaseDSN("user:secret@tcp(mariadb:4001)/photoprism?"+query, 15)
			require.NoError(t, err, query)
			assert.Equal(t, "user:secret@tcp(mariadb:4001)/photoprism?"+want, result, query)
		}
	})
	t.Run("DriverView", func(t *testing.T) {
		// The Go driver reads the rebuilt DSN with the validated parts and without other options.
		for s, accepted := range map[string]bool{
			"user:secret@tcp(mariadb:4001)/photoprism?clientFoundRows=true&columnsWithAlias=true&time_zone=x": true,
			"user:secret@tcp(mariadb:4001)/photoprism?x=/y?option=value":                                      false,
			"user:secret@tcp(mariadb:4001)/photoprism?x=/y&option=value":                                      true,
			"user:p/w@x?option=value@tcp(mariadb:4001)/photoprism?tls=skip-verify&interpolateParams=true":     true,
		} {
			result, _, err := clusterDatabaseDSN(s, 15)
			if !accepted {
				assert.Error(t, err, s)
				continue
			}
			require.NoError(t, err, s)
			d := dsn.Parse(result)
			cfg, err := mysql.ParseDSN(result)
			require.NoError(t, err, s)
			assert.Equal(t, "photoprism", cfg.DBName, s)
			assert.Equal(t, d.User, cfg.User, s)
			assert.Equal(t, d.Password, cfg.Passwd, s)
			assert.Equal(t, "mariadb:4001", cfg.Addr, s)
			assert.False(t, cfg.ClientFoundRows, s)
			assert.False(t, cfg.ColumnsWithAlias, s)
			assert.Empty(t, cfg.Params, s)
		}
	})
	t.Run("Empty", func(t *testing.T) {
		result, dropped, err := clusterDatabaseDSN("", 15)
		require.NoError(t, err)
		assert.Equal(t, "", result)
		assert.Empty(t, dropped)
	})
	t.Run("Invalid", func(t *testing.T) {
		for s, errMsg := range map[string]string{
			"user:secret@tcp(mariadb:4001)/db?x?option=value":                       "invalid database name",
			"postgres://user:secret@db:5432/photoprism":                             "invalid database dsn",
			"user:secret@tcp(mariadb:4001)/":                                        "invalid database dsn",
			":secret@tcp(mariadb:4001)/photoprism":                                  "invalid database dsn",
			"user:secret@tcp(mariadb:4001)/photoprism?charset=latin1":               "invalid database dsn",
			"user:secret@tcp(mariadb:4001)/photoprism?collation=latin1_swedish_ci":  "invalid database dsn",
			"user:secret@tcp(mariadb:4001)/photoprism?tls=custom":                   "invalid database dsn",
			"user:secret@tcp(mariadb:4001)/photoprism?tls=false&tls=true":           "invalid database dsn",
			"user:secret@tcp(mariadb:4001)/photoprism?parseTime=false":              "invalid database dsn",
			"user:secret@tcp(mariadb:4001)/photoprism?parseTime=tRuE":               "invalid database dsn",
			"user:secret@tcp(mariadb:4001)/photoprism?maxAllowedPacket=12345678901": "invalid database dsn",
			"user:secret@tcp(mariadb:4001)/photoprism?maxAllowedPacket=1073741825":  "invalid database dsn",
			"user:secret@tcp(mariadb:4001)/photoprism?maxAllowedPacket=01":          "invalid database dsn",
			"user:secret@tcp(mariadb:4001)/photoprism?maxAllowedPacket=+5":          "invalid database dsn",
		} {
			_, _, err := clusterDatabaseDSN(s, 15)
			assert.EqualError(t, err, errMsg, s)
		}
	})
}

// TestConfig_DatabaseServerDSN checks that the database host and port come from the configured server.
func TestConfig_DatabaseServerDSN(t *testing.T) {
	t.Run("DiscretePassword", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.DatabaseDriver = dsn.DriverMySQL
		c.options.DatabaseDSN = ""
		c.options.DatabaseServer = "mariadb:4001"
		c.options.DatabaseName = "photoprism"
		c.options.DatabaseUser = "photoprism"
		c.options.DatabasePassword = "a://b@tcp(example.com:1)/c"
		assert.Equal(t, "mariadb", c.DatabaseHost())
		assert.Equal(t, 4001, c.DatabasePort())
		cfg, err := mysql.ParseDSN(c.DatabaseDSN())
		require.NoError(t, err)
		assert.Equal(t, "mariadb:4001", cfg.Addr)
		assert.Equal(t, "photoprism", cfg.DBName)
		assert.Equal(t, "a://b@tcp(example.com:1)/c", cfg.Passwd)
	})
	t.Run("Socket", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.DatabaseDriver = dsn.DriverMySQL
		c.options.DatabaseDSN = ""
		c.options.DatabaseServer = "/run/mysqld/mysqld.sock"
		assert.Equal(t, "/run/mysqld/mysqld.sock", c.DatabaseHost())
		assert.Equal(t, 3306, c.DatabasePort())
		assert.Contains(t, c.DatabaseDSN(), "@unix(/run/mysqld/mysqld.sock)/")
	})
	t.Run("ConfiguredDSN", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.DatabaseDriver = dsn.DriverMySQL
		c.options.DatabaseDSN = "user:secret@tcp(db.example.com:3307)/photoprism?parseTime=true"
		assert.Equal(t, "db.example.com", c.DatabaseHost())
		assert.Equal(t, 3307, c.DatabasePort())
	})
}

// TestClusterDatabaseParamNames checks which dropped DSN parameter names may be logged.
func TestClusterDatabaseParamNames(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, []string{"option", "time_zone", "a.b-c"}, clusterDatabaseParamNames([]string{"option", "time_zone", "a.b-c"}))
	})
	t.Run("Filtered", func(t *testing.T) {
		assert.Equal(t, []string{"option"}, clusterDatabaseParamNames([]string{"", "x, y", "option", "a b", strings.Repeat("a", 65)}))
		assert.Empty(t, clusterDatabaseParamNames(nil))
	})
}
