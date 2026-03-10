package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/service/cluster"
)

func TestConfig_Report(t *testing.T) {
	m := NewConfig(CliTestContext())
	r, _ := m.Report()
	assert.GreaterOrEqual(t, len(r), 1)

	values := make(map[string]string, len(r))

	for _, row := range r {
		if len(row) < 2 {
			continue
		}

		values[row[0]] = row[1]
	}

	assert.Equal(t, m.FrontendUri(""), values["frontend-uri"])
}

func TestConfig_ReportServicesCIDROrder(t *testing.T) {
	conf := NewConfig(CliTestContext())
	rows, _ := conf.Report()

	indexOf := func(name string) int {
		for i := range rows {
			if len(rows[i]) > 0 && rows[i][0] == name {
				return i
			}
		}

		return -1
	}

	proxyProtoHTTPS := indexOf("proxy-proto-https")
	servicesCIDR := indexOf("services-cidr")
	disableTLS := indexOf("disable-tls")

	assert.Greater(t, proxyProtoHTTPS, -1)
	assert.Greater(t, servicesCIDR, -1)
	assert.Greater(t, disableTLS, -1)
	assert.Greater(t, servicesCIDR, proxyProtoHTTPS)
	assert.Less(t, servicesCIDR, disableTLS)
}

func TestConfig_ReportHttpHeaderSettingsOrder(t *testing.T) {
	conf := NewConfig(CliTestContext())
	rows, _ := conf.Report()

	indexOf := func(name string) int {
		for i := range rows {
			if len(rows[i]) > 0 && rows[i][0] == name {
				return i
			}
		}

		return -1
	}

	httpCompression := indexOf("http-compression")
	httpHeaderTimeout := indexOf("http-header-timeout")
	httpHeaderBytes := indexOf("http-header-bytes")
	httpIdleTimeout := indexOf("http-idle-timeout")
	httpCachePublic := indexOf("http-cache-public")

	assert.Greater(t, httpCompression, -1)
	assert.Greater(t, httpHeaderTimeout, -1)
	assert.Greater(t, httpHeaderBytes, -1)
	assert.Greater(t, httpIdleTimeout, -1)
	assert.Greater(t, httpCachePublic, -1)
	assert.Greater(t, httpHeaderTimeout, httpCompression)
	assert.Greater(t, httpHeaderBytes, httpHeaderTimeout)
	assert.Greater(t, httpIdleTimeout, httpHeaderBytes)
	assert.Less(t, httpIdleTimeout, httpCachePublic)
}

func TestConfig_ReportDatabaseSection(t *testing.T) {
	collect := func(rows [][]string) map[string]string {
		result := make(map[string]string, len(rows))

		for _, row := range rows {
			if len(row) < 2 {
				continue
			}

			result[row[0]] = row[1]
		}

		return result
	}
	t.Run("SQLiteReportsDSN", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Equal(t, SQLite3, values["database-driver"])
		assert.Equal(t, conf.DatabaseDSN(), values["database-dsn"])
		_, hasName := values["database-name"]
		assert.False(t, hasName)
	})
	t.Run("MariaDBReportsIndividualFields", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		conf.options.DatabaseDriver = MySQL
		conf.options.DatabaseServer = "db.internal:3306"
		conf.options.DatabaseName = "photoprism"
		conf.options.DatabaseUser = "app"
		conf.options.DatabasePassword = "secret"

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Equal(t, MySQL, values["database-driver"])
		assert.Equal(t, "photoprism", values["database-name"])
		assert.Equal(t, "db.internal:3306", values["database-server"])
		assert.Equal(t, "db.internal", values["database-host"])
		assert.Equal(t, "3306", values["database-port"])
		assert.Equal(t, "app", values["database-user"])
		assert.Equal(t, strings.Repeat("*", len("secret")), values["database-password"])
		_, hasDSN := values["database-dsn"]
		assert.False(t, hasDSN)
	})
	t.Run("MariaDBReportsDSNWhenConfigured", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		conf.options.DatabaseDriver = MySQL
		conf.options.DatabaseDSN = "user:pass@tcp(db.internal:3306)/photoprism"

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Equal(t, MySQL, values["database-driver"])
		assert.Equal(t, "user:***@tcp(db.internal:3306)/photoprism?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true&timeout=15s", values["database-dsn"])
		_, hasName := values["database-name"]
		assert.False(t, hasName)
		_, hasPassword := values["database-password"]
		assert.False(t, hasPassword)
	})
}

func TestConfig_ReportPortalSettingsVisibility(t *testing.T) {
	collect := func(rows [][]string) map[string]string {
		result := make(map[string]string, len(rows))

		for _, row := range rows {
			if len(row) < 2 {
				continue
			}

			result[row[0]] = row[1]
		}

		return result
	}

	t.Run("NonPortalOmitsPortalSettings", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		conf.options.NodeRole = cluster.RoleInstance

		rows, _ := conf.Report()
		values := collect(rows)

		_, hasProxy := values["portal-proxy"]
		_, hasURI := values["portal-proxy-uri"]
		_, hasConfigPath := values["portal-config-path"]
		_, hasThemePath := values["portal-theme-path"]

		assert.False(t, hasProxy)
		assert.False(t, hasURI)
		assert.False(t, hasConfigPath)
		assert.False(t, hasThemePath)
	})
	t.Run("PortalIncludesPortalSettings", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		conf.options.Edition = Portal
		conf.options.NodeRole = cluster.RolePortal
		conf.options.PortalProxy = true
		conf.options.PortalProxyUri = "https://proxy.example.com/instance/"

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Equal(t, "true", values["portal-proxy"])
		assert.Equal(t, "https://proxy.example.com/instance/", values["portal-proxy-uri"])
		assert.Equal(t, conf.PortalConfigPath(), values["portal-config-path"])
		assert.Equal(t, conf.PortalThemePath(), values["portal-theme-path"])
	})
}

func TestConfig_ReportThemeURLVisibility(t *testing.T) {
	indexOf := func(rows [][]string, name string) int {
		for i := range rows {
			if len(rows[i]) > 0 && rows[i][0] == name {
				return i
			}
		}

		return -1
	}

	collect := func(rows [][]string) map[string]string {
		result := make(map[string]string, len(rows))

		for _, row := range rows {
			if len(row) < 2 {
				continue
			}

			result[row[0]] = row[1]
		}

		return result
	}

	originalFeatures := Features
	t.Cleanup(func() { Features = originalFeatures })

	t.Run("CommunityOmitsThemeURL", func(t *testing.T) {
		Features = Community
		conf := NewConfig(CliTestContext())
		conf.options.NodeRole = cluster.RoleInstance
		conf.SetThemeUrl("https://cdn.photoprism.app/theme.zip")

		rows, _ := conf.Report()
		values := collect(rows)

		_, hasThemeURL := values["theme-url"]
		assert.False(t, hasThemeURL)
		assert.Equal(t, -1, indexOf(rows, "theme-url"))
	})
	t.Run("PortalIncludesThemeURL", func(t *testing.T) {
		Features = Community
		conf := NewConfig(CliTestContext())
		conf.options.Edition = Portal
		conf.options.NodeRole = cluster.RolePortal
		conf.SetThemeUrl("https://demo:secret@cdn.photoprism.app/theme.zip")

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Contains(t, values["theme-url"], "https://demo:")
		assert.NotContains(t, values["theme-url"], "secret")
		assert.Greater(t, indexOf(rows, "theme-url"), indexOf(rows, "default-theme"))
		assert.Less(t, indexOf(rows, "theme-url"), indexOf(rows, "places-locale"))
	})
	t.Run("ProIncludesThemeURL", func(t *testing.T) {
		Features = Pro
		conf := NewConfig(CliTestContext())
		conf.options.NodeRole = cluster.RoleInstance
		conf.SetThemeUrl("https://cdn.photoprism.app/theme.zip")

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Equal(t, "https://cdn.photoprism.app/theme.zip", values["theme-url"])
		assert.Greater(t, indexOf(rows, "theme-url"), indexOf(rows, "default-theme"))
		assert.Less(t, indexOf(rows, "theme-url"), indexOf(rows, "places-locale"))
	})
}

func TestConfig_ReportURIRedaction(t *testing.T) {
	collect := func(rows [][]string) map[string]string {
		result := make(map[string]string, len(rows))

		for _, row := range rows {
			if len(row) < 2 {
				continue
			}

			result[row[0]] = row[1]
		}

		return result
	}

	originalFeatures := Features
	t.Cleanup(func() { Features = originalFeatures })

	Features = Pro

	conf := NewConfig(CliTestContext())
	conf.options.PortalUrl = "https://portal:secret@example.com"
	conf.options.JWKSUrl = "https://jwks:secret@jwks.example.com/.well-known/jwks.json"
	conf.options.AdvertiseUrl = "https://cluster:secret@node.example.com"
	conf.options.HttpsProxy = "https://proxy:secret@proxy.example.com:8443"
	conf.options.VisionUri = "https://vision:secret@vision.example.com/api/v1/vision"
	conf.SetThemeUrl("https://theme:secret@cdn.photoprism.app/theme.zip")

	rows, _ := conf.Report()
	values := collect(rows)

	assert.Equal(t, "https://portal:xxxxx@example.com", values["portal-url"])
	assert.Equal(t, "https://jwks:xxxxx@jwks.example.com/.well-known/jwks.json", values["jwks-url"])
	assert.Equal(t, "https://cluster:xxxxx@node.example.com/", values["advertise-url"])
	assert.Equal(t, "https://proxy:xxxxx@proxy.example.com:8443", values["https-proxy"])
	assert.Equal(t, "https://vision:xxxxx@vision.example.com/api/v1/vision", values["vision-uri"])
	assert.Equal(t, "https://theme:xxxxx@cdn.photoprism.app/theme.zip", values["theme-url"])
}
