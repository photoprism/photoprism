package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/dsn"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/mutex"
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

		assert.Equal(t, dsn.DriverSQLite3, values["database-driver"])
		assert.Equal(t, conf.DatabaseDSN(), values["database-dsn"])
		_, hasName := values["database-name"]
		assert.False(t, hasName)
	})
	t.Run("MariaDBReportsIndividualFields", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		conf.options.DatabaseDriver = dsn.DriverMySQL
		conf.options.DatabaseServer = "db.internal:3306"
		conf.options.DatabaseName = "photoprism"
		conf.options.DatabaseUser = "app"
		conf.options.DatabasePassword = "secret"

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Equal(t, dsn.DriverMySQL, values["database-driver"])
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

		conf.options.DatabaseDriver = dsn.DriverMySQL
		conf.options.DatabaseDSN = "user:pass@tcp(db.internal:3306)/photoprism"

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Equal(t, dsn.DriverMySQL, values["database-driver"])
		assert.Equal(t, "user:***@tcp(db.internal:3306)/photoprism?timeout=15s&charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true", values["database-dsn"])
		_, hasName := values["database-name"]
		assert.False(t, hasName)
		_, hasPassword := values["database-password"]
		assert.False(t, hasPassword)
	})
	t.Run("PostgresReportsIndividualFields", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		conf.options.DatabaseDriver = dsn.DriverPostgreSQL
		conf.options.DatabaseServer = "db.internal:4002"
		conf.options.DatabaseName = "photoprism"
		conf.options.DatabaseUser = "app"
		conf.options.DatabasePassword = "secret"

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Equal(t, dsn.DriverPostgreSQL, values["database-driver"])
		assert.Equal(t, "photoprism", values["database-name"])
		assert.Equal(t, "db.internal:4002", values["database-server"])
		assert.Equal(t, "db.internal", values["database-host"])
		assert.Equal(t, "4002", values["database-port"])
		assert.Equal(t, "app", values["database-user"])
		assert.Equal(t, strings.Repeat("*", len("secret")), values["database-password"])
		_, hasDSN := values["database-dsn"]
		assert.False(t, hasDSN)
	})
	t.Run("PostgresReportsDSNWhenConfigured", func(t *testing.T) {
		conf := NewConfig(CliTestContext())
		resetDatabaseOptions(conf)

		conf.options.DatabaseDriver = dsn.DriverPostgreSQL
		conf.options.DatabaseDSN = "postgres://user:pass@db.internal:4002/photoprism"

		rows, _ := conf.Report()
		values := collect(rows)

		assert.Equal(t, dsn.DriverPostgreSQL, values["database-driver"])
		assert.Equal(t, "postgres://user:***@db.internal:4002/photoprism?connect_timeout=15000&sslmode=disable&TimeZone=UTC", values["database-dsn"])
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

func TestConfig_FaceReportSections(t *testing.T) {
	m := NewConfig(CliTestContext())
	sections := m.FaceReportSections()

	titles := make([]string, 0, len(sections))
	values := make(map[string]string)
	notes := make(map[string]string, len(sections))

	for _, section := range sections {
		assert.Equal(t, []string{"Name", "Value"}, section.Cols)
		assert.NotEmpty(t, section.Rows, section.Title)

		titles = append(titles, section.Title)
		notes[section.Title] = section.Note

		for _, row := range section.Rows {
			if len(row) < 2 {
				continue
			}

			assert.NotContains(t, values, row[0], "reported in more than one section")
			values[row[0]] = row[1]
		}
	}

	t.Run("Sections", func(t *testing.T) {
		assert.Equal(t, []string{"Global Options", "Face Detection", "Face Recognition"}, titles)
		assert.Empty(t, notes["Global Options"])
		assert.Contains(t, notes["Face Detection"], "Detector:")
		assert.Contains(t, notes["Face Recognition"], "Model:")
	})
	t.Run("Options", func(t *testing.T) {
		// The rows are named by the option `photoprism show config` reports, so both reports name
		// the same value the same way and can be read against each other.
		expected := []string{
			"xmp-faces",
			"face-run",
			"face-detector",
			"face-detector-path",
			"face-detector-threads",
			"face-size",
			"face-size-retry",
			"face-score",
			"face-migrate-size",
			"face-migrate-score",
			"face-overlap",
			"face-model",
			"face-model-path",
			"face-model-threads",
			"face-cluster-size",
			"face-cluster-score",
			"face-cluster-core",
			"face-cluster-dist",
			"face-cluster-radius",
			"face-match-dist",
			"face-collision-dist",
			"face-epsilon-dist",
		}

		for _, name := range expected {
			assert.Contains(t, values, name)
		}
	})
	t.Run("OnlyFaceOptions", func(t *testing.T) {
		for _, name := range []string{"originals-path", "auth-mode", "ffmpeg-bin", "site-url", "vision-yaml"} {
			assert.NotContains(t, values, name)
		}
	})
	t.Run("EngineHiddenWhenAuto", func(t *testing.T) {
		assert.NotContains(t, values, "face-engine")
	})
}

func TestConfig_FaceStatus(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, []string{"Face detection and recognition are unavailable."}, (*Config)(nil).FaceStatus())
	})
	t.Run("Disabled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.DisableFaces = true

		assert.Equal(t, []string{"Face detection and recognition are disabled."}, c.FaceStatus())
	})
	t.Run("NoDetector", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceDetector = face.DetectorNone

		assert.Contains(t, c.FaceStatus()[0], "Face detection is disabled")
	})
	t.Run("NeverScheduled", func(t *testing.T) {
		// Everything is installed and configured, and nothing will ever run it. No value in the
		// tables says so; that is what the prose is for.
		c := NewConfig(CliTestContext())
		c.options.FaceRun = vision.RunNever

		assert.Contains(t, c.FaceStatus()[0], "never scheduled to run")
	})
	t.Run("MigrationLockIsNamed", func(t *testing.T) {
		// A lock nothing reports is the difference between an instance that explains why it
		// indexes nothing and one that has to be diagnosed by finding a file.
		c := NewConfig(CliTestContext())
		c.options.StoragePath = t.TempDir()

		lock, err := mutex.AcquireFileLock(c.FacesLockFile(), "test migration")
		require.NoError(t, err)
		t.Cleanup(lock.Release)

		status := strings.Join(c.FaceStatus(), " ")

		assert.Contains(t, status, "A face embedding migration is in progress")
		assert.Contains(t, status, "test migration")
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

func TestFaceModelStatus(t *testing.T) {
	restore := face.ConfiguredModel()

	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})

	t.Run("Ok", func(t *testing.T) {
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelFaceNet, Model: face.FindEmbeddingModel(face.ModelFaceNet)}))
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelFaceNet)
		c.options.FaceModel = face.ModelFaceNet
		assert.Empty(t, c.faceModelStatus())
	})
	t.Run("Failed", func(t *testing.T) {
		// A model that fails to load reports ModelNone, so the status is the only signal.
		require.Error(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelSFace, Model: face.FindEmbeddingModel(face.ModelSFace)}))
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelSFace
		assert.Contains(t, c.faceModelStatus(), "failed to load")
	})
	t.Run("Paused", func(t *testing.T) {
		t.Cleanup(face.UnblockEmbeddings)
		require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelFaceNet, Model: face.FindEmbeddingModel(face.ModelFaceNet)}))
		face.BlockEmbeddings("12 marker(s) use facenet")

		c := NewConfig(CliTestContext())
		c.options.ModelsPath = installTestModels(t, face.ModelFaceNet)
		c.options.FaceModel = face.ModelFaceNet

		assert.Contains(t, c.faceModelStatus(), "paused")
	})
}

func TestConfig_faceModelReport(t *testing.T) {
	// The embedder is process-wide, and a test that leaves it in an error state would otherwise
	// show up here as every model failing to load.
	restore := face.ConfiguredModel()
	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelNone}))
	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})

	t.Run("Named", func(t *testing.T) {
		// SFace is the shipped default, so the row says so whether it was named or detected: the
		// word describes the model rather than how this instance arrived at it.
		c := newSFaceTestConfig(t)

		assert.Equal(t, face.ModelDisplayName(face.ModelSFace)+" (default)", c.faceModelReport())
	})
	t.Run("Default", func(t *testing.T) {
		// Nothing is pinned, so the row names the model that is going to apply and says the
		// operator did not choose it.
		c := NewConfig(CliTestContext())
		c.options.FaceModel = ""

		assert.Equal(t, face.ModelDisplayName(c.installedFaceModel())+" (default)", c.faceModelReport())
	})
	t.Run("NotAvailable", func(t *testing.T) {
		// A configured model that cannot be loaded resolves to none everywhere else, so the row
		// is the only place an operator sees which one was asked for.
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.FaceModel = face.ModelSFace

		assert.Equal(t, face.ModelNone+" ("+face.ModelSFace+" is not available)", c.faceModelReport())
	})
	t.Run("NoneInstalled", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.ModelsPath = t.TempDir()
		c.options.FaceModel = ""

		assert.Equal(t, face.ModelNone+" (no embedding model is installed)", c.faceModelReport())
	})
	t.Run("Disabled", func(t *testing.T) {
		// The report commands never load the model, so this must follow the configuration.
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone

		assert.Equal(t, face.ModelNone+" (embeddings disabled)", c.faceModelReport())
	})
	t.Run("Paused", func(t *testing.T) {
		t.Cleanup(face.UnblockEmbeddings)
		c := newSFaceTestConfig(t)
		face.BlockEmbeddings("12 marker(s) use facenet")

		assert.Equal(t, face.ModelDisplayName(face.ModelSFace)+" (default, paused: 12 marker(s) use facenet)", c.faceModelReport())
	})
	t.Run("DetectNamesTheLibraryModel", func(t *testing.T) {
		// "faces status" connects, so it is the one report where the detected model is the one
		// the library actually holds rather than the one a fresh install would start with.
		c := TestConfig()
		setting := c.options.FaceModel
		t.Cleanup(func() { c.options.FaceModel = setting })
		c.options.FaceModel = face.ModelAuto

		assert.Contains(t, c.faceModelReport(), "(default")
	})
}

func TestConfig_faceDistReport(t *testing.T) {
	t.Run("ModelInForce", func(t *testing.T) {
		c := newSFaceTestConfig(t)

		assert.Equal(t, "0.850000", c.faceDistReport(c.FaceClusterDist))
	})
	t.Run("NoModel", func(t *testing.T) {
		// The calibrated distances are per model, so a report with none must not print
		// five numbers that will not apply to the model the instance ends up using.
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone

		assert.Equal(t, "", c.faceDistReport(c.FaceClusterDist))
	})
}

func TestConfig_faceEngineReport(t *testing.T) {
	// The row reports the configured value rather than the runtime in force, which now follows
	// the detector: a stale "none" in options.yml is the only thing that explains why detection
	// is off, and blanking it would take that away.
	c := NewConfig(CliTestContext())

	t.Run("Deprecated", func(t *testing.T) {
		c.options.FaceEngine = face.EngineNone
		assert.Equal(t, "none (deprecated)", c.faceEngineReport())
	})
	t.Run("LegacyAlias", func(t *testing.T) {
		c.options.FaceEngine = "pigo"
		assert.Equal(t, "onnx (deprecated)", c.faceEngineReport())
	})
	t.Run("Unset", func(t *testing.T) {
		c.options.FaceEngine = ""
		assert.Equal(t, "auto (deprecated)", c.faceEngineReport())
	})
}

func TestConfig_faceDetectorReport(t *testing.T) {
	c := NewConfig(CliTestContext())
	t.Cleanup(func() { c.options.FaceDetector = "" })

	t.Run("Named", func(t *testing.T) {
		// The faces report is read rather than parsed, so it names the detector the way an operator
		// would. "show config" keeps the identifier. "(default)" is a property of the value, so it
		// stands whether the detector was named or derived.
		c.options.FaceDetector = face.DetectorYuNet
		assert.Equal(t, face.DetectorDisplayName(face.DetectorYuNet)+" (default)", c.faceDetectorReport())
	})
	t.Run("NamedButNotTheDefault", func(t *testing.T) {
		c.options.FaceDetector = face.DetectorSCRFD
		t.Setenv(face.LicenseAcceptanceVar, "1")

		if c.FaceDetector() != face.DetectorSCRFD {
			t.Skip("scrfd is not installed")
		}

		assert.Equal(t, face.DetectorDisplayName(face.DetectorSCRFD), c.faceDetectorReport())
	})
	t.Run("Disabled", func(t *testing.T) {
		c.options.FaceDetector = face.DetectorNone
		assert.Equal(t, face.DetectorNone, c.faceDetectorReport())
	})
	t.Run("Derived", func(t *testing.T) {
		// Nothing was configured, so the row names the detector that is going to run rather than
		// the word "auto", which says nothing about whether anything will be detected.
		c.options.FaceDetector = ""
		assert.Equal(t, fmt.Sprintf("%s (default)", face.DetectorDisplayName(c.FaceDetector())), c.faceDetectorReport())
		assert.Equal(t, face.DefaultDetectorName(), c.FaceDetector())
	})
	t.Run("NotAvailable", func(t *testing.T) {
		// A named detector that cannot run disables detection, so the row has to name it: "none"
		// alone reads as a setting rather than as a detector that could not be loaded.
		c.options.ModelsPath = t.TempDir()
		t.Cleanup(func() { c.options.ModelsPath = "" })
		c.options.FaceDetector = face.DetectorYuNet
		assert.Equal(t, face.DetectorNone+" ("+face.DetectorYuNet+" is not available)", c.faceDetectorReport())
		assert.Equal(t, face.DetectorNone, face.DetectorDisplayName(face.DetectorNone), "a disabled detector has no display name to show")
	})
}

// TestConfig_faceConfigRows pins the row order to the order flags.go declares the options, which
// is what makes both reports diffable against the flag list rather than against each other.
func TestConfig_faceConfigRows(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Run("FlagOrder", func(t *testing.T) {
		want := []string{
			"xmp-faces",
			"face-run",
			"face-detector",
			"face-detector-path",
			"face-detector-threads",
			"face-size",
			"face-size-retry",
			"face-score",
			"face-migrate-size",
			"face-migrate-score",
			"face-overlap",
			"face-model",
			"face-model-path",
			"face-model-threads",
			"face-cluster-size",
			"face-cluster-score",
			"face-cluster-core",
			"face-cluster-dist",
			"face-cluster-radius",
			"face-match-dist",
			"face-collision-dist",
			"face-epsilon-dist",
		}

		rows := c.faceConfigRows()
		flags := make([]string, 0, len(rows))

		for _, row := range rows {
			flags = append(flags, row.Flag)
			assert.NotEmpty(t, row.Section, row.Flag)
		}

		assert.Equal(t, want, flags)
	})
	t.Run("EveryFlagIsRegistered", func(t *testing.T) {
		// A row naming an option that no longer exists reads as a setting an operator can change.
		// The two derived paths have no flag of their own.
		derived := map[string]bool{"face-detector-path": true, "face-model-path": true}

		registered := make(map[string]bool, len(Flags))

		for _, flag := range Flags {
			registered[flag.Name()] = true
		}

		for _, row := range c.faceConfigRows() {
			if derived[row.Flag] {
				continue
			}

			assert.True(t, registered[row.Flag], row.Flag)
		}
	})
	t.Run("ValuesOnly", func(t *testing.T) {
		// "show config" runs without a database, so it can neither detect the model nor see that
		// one is blocked. Stating a resolution it cannot check is how it came to report "ok" on a
		// paused instance, so every resolution qualifier belongs in a note beside the table.
		//
		// face-engine is the exception and has to be forced on to be reached at all: its marker
		// names the option as deprecated, which is a property of the setting rather than a
		// resolution, and it is the only row that carries one.
		engine := c.options.FaceEngine
		t.Cleanup(func() { c.options.FaceEngine = engine })
		c.options.FaceEngine = face.EngineNone

		rows := c.faceConfigRows()
		flags := make([]string, 0, len(rows))

		for _, row := range rows {
			flags = append(flags, row.Flag)

			if row.Flag == "face-engine" {
				assert.Equal(t, "none (deprecated)", row.Value)
				continue
			}

			assert.NotContains(t, row.Value, "(", row.Flag)
		}

		require.Contains(t, flags, "face-engine", "the exempt row must be reached, or this asserts nothing")
	})
	t.Run("ResolvedScores", func(t *testing.T) {
		// Zero is what the two raw options hold in the ordinary case, and it reads as "nothing is
		// filtered" - the one thing it does not mean.
		values := make(map[string]string)

		for _, row := range c.faceConfigRows() {
			values[row.Flag] = row.Value
		}

		assert.Equal(t, fmt.Sprintf("%g", c.FaceScoreEffective()), values["face-score"])
		assert.Equal(t, fmt.Sprintf("%d", c.FaceClusterScoreEffective()), values["face-cluster-score"])
		assert.NotEqual(t, "0", values["face-score"])
		assert.NotEqual(t, "0", values["face-cluster-score"])
	})
	t.Run("EngineReportedWhenItDecides", func(t *testing.T) {
		engine := c.options.FaceEngine
		t.Cleanup(func() { c.options.FaceEngine = engine })

		flags := func() []string {
			var names []string

			for _, row := range c.faceConfigRows() {
				names = append(names, row.Flag)
			}

			return names
		}

		c.options.FaceEngine = face.EngineAuto
		assert.NotContains(t, flags(), "face-engine")
		c.options.FaceEngine = face.EngineNone
		assert.Contains(t, flags(), "face-engine")
	})
}

func TestConfig_faceDetectionNote(t *testing.T) {
	t.Run("CalibratedScore", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceScore = 0

		assert.Contains(t, c.faceDetectionNote(), "Detector: ")
		assert.Contains(t, c.faceDetectionNote(), fmt.Sprintf("minimum score of %g is calibrated for %s", c.FaceScoreEffective(), c.FaceDetector()))
	})
	t.Run("ConfiguredScore", func(t *testing.T) {
		// An operator's own value has no calibration to attribute it to.
		c := NewConfig(CliTestContext())
		c.options.FaceScore = 42

		assert.NotContains(t, c.faceDetectionNote(), "calibrated")
	})
	t.Run("Disabled", func(t *testing.T) {
		// A cutoff no detector enforces must not be attributed to one.
		c := NewConfig(CliTestContext())
		c.options.FaceDetector = face.DetectorNone

		assert.NotContains(t, c.faceDetectionNote(), "calibrated")
	})
}

func TestConfig_faceRecognitionNote(t *testing.T) {
	t.Run("CalibratedClusterScore", func(t *testing.T) {
		c := newSFaceTestConfig(t)
		c.options.FaceClusterScore = 0

		assert.Contains(t, c.faceRecognitionNote(), "Model: ")
		assert.Contains(t, c.faceRecognitionNote(), fmt.Sprintf("%d where none is recorded", face.ClusterScoreThresholdDefault),
			"a library of markers predating the provenance column is filtered at the shared default, not the detector's own bar")
	})
	t.Run("ConfiguredClusterScore", func(t *testing.T) {
		c := newSFaceTestConfig(t)
		c.options.FaceClusterScore = 42

		assert.NotContains(t, c.faceRecognitionNote(), "calibrated")
	})
	t.Run("NoModel", func(t *testing.T) {
		// The distance rows are blank without a model, so the note has to say why rather than
		// leaving five empty cells to be read as zero.
		c := NewConfig(CliTestContext())
		c.options.FaceModel = face.ModelNone

		assert.Contains(t, c.faceRecognitionNote(), "once a model is in force")
	})
}

// TestConfig_faceClusterStatus pins that the report names the bar holding clustering back. A
// library that never forms a cluster otherwise looks exactly like one that clustered everything.
func TestConfig_faceClusterStatus(t *testing.T) {
	c := TestConfig()

	t.Run("SizeIsTheGate", func(t *testing.T) {
		size := c.options.FaceClusterSize
		t.Cleanup(func() { c.options.FaceClusterSize = size })
		c.options.FaceClusterSize = 10000

		status := c.faceClusterStatus()

		assert.Contains(t, status, "Automatic clustering needs")
		assert.Contains(t, status, "face-cluster-size of 10000 px")
		assert.Contains(t, status, "face-cluster-core")
	})
	t.Run("RequiredMatchesTheWorker", func(t *testing.T) {
		// A non-default core, because at the default the getter and the package variable coincide
		// and the test cannot see which one the report read. This command never propagates, so
		// reading the global here reported the shipped default whatever FACE_CLUSTER_CORE said.
		size, core := c.options.FaceClusterSize, c.options.FaceClusterCore
		t.Cleanup(func() { c.options.FaceClusterSize, c.options.FaceClusterCore = size, core })
		c.options.FaceClusterSize = 10000
		c.options.FaceClusterCore = 7

		require.NotEqual(t, c.FaceSampleThreshold(), face.SampleThreshold, "the fixture must tell the two apart")
		assert.Contains(t, c.faceClusterStatus(), fmt.Sprintf("needs %d new markers", c.FaceSampleThreshold()))
	})
	t.Run("ScoreBarFollowsTheFloorMapping", func(t *testing.T) {
		// Unset, the counts are taken per marker, so naming one number would state the bar of the
		// detector in force for markers a different one scored.
		size, score := c.options.FaceClusterSize, c.options.FaceClusterScore
		t.Cleanup(func() { c.options.FaceClusterSize, c.options.FaceClusterScore = size, score })
		c.options.FaceClusterSize = 10000
		c.options.FaceClusterScore = 0

		assert.Contains(t, c.faceClusterStatus(), "the detector that scored each one")

		c.options.FaceClusterScore = 42
		assert.Contains(t, c.faceClusterStatus(), "face-cluster-score of 42")
	})
	t.Run("NoDatabase", func(t *testing.T) {
		// It runs one query per bar, so a report without a connection must skip it rather than
		// count zero and blame a threshold.
		assert.Empty(t, NewConfig(CliTestContext()).faceClusterStatus())
	})
	t.Run("NilConfig", func(t *testing.T) {
		assert.Empty(t, (*Config)(nil).faceClusterStatus())
	})
}

func TestFaceClusterStatusFor(t *testing.T) {
	const bar = "the face-cluster-score of 70"

	t.Run("Enough", func(t *testing.T) {
		// A line every healthy instance prints is one nobody reads on the instance that is not.
		assert.Empty(t, faceClusterStatusFor(query.FaceClusterGates{Unclustered: 20, Recent: 20, SizeOK: 20, ScoreOK: 20, Eligible: 20}, 8, 60, bar, 4))
	})
	t.Run("NothingUnclustered", func(t *testing.T) {
		assert.Empty(t, faceClusterStatusFor(query.FaceClusterGates{}, 8, 60, bar, 4))
	})
	t.Run("StrandedBehindTheLastCluster", func(t *testing.T) {
		// The shortfall no threshold explains: every marker predates the newest cluster, so none
		// counts toward the trigger again and the instance sits idle looking healthy.
		status := faceClusterStatusFor(query.FaceClusterGates{Unclustered: 6, Clusterable: 4}, 8, 60, bar, 4)

		assert.Contains(t, status, "6 markers are unclustered")
		assert.Contains(t, status, "none was added since the last cluster")
		// What a forced run would take, or the remedy cannot be weighed.
		assert.Contains(t, status, "4 of them clear both thresholds")
		assert.Contains(t, status, "faces update --force")
	})
	t.Run("VolumeOnly", func(t *testing.T) {
		// Every marker clears every bar, so naming thresholds would send an operator to tune bars
		// that are excluding nothing.
		status := faceClusterStatusFor(query.FaceClusterGates{Unclustered: 3, Recent: 3, SizeOK: 3, ScoreOK: 3, Eligible: 3}, 8, 60, bar, 4)

		assert.Contains(t, status, "needs 8 new markers (2 x face-cluster-core 4) and has 3")
		assert.Contains(t, status, "no threshold is excluding")
		assert.NotContains(t, status, "face-cluster-size")
	})
	t.Run("SizeIsTheGate", func(t *testing.T) {
		status := faceClusterStatusFor(query.FaceClusterGates{Unclustered: 40, Recent: 40, SizeOK: 2, ScoreOK: 40, Eligible: 2}, 8, 60, bar, 4)

		// The two numbers look contradictory unless the derivation is named, and the size and
		// score counts overlap, so the eligible one has to read as their intersection.
		assert.Contains(t, status, "needs 8 new markers (2 x face-cluster-core 4)")
		assert.Contains(t, status, "has 2 clearing both")
		assert.Contains(t, status, "of the 40 added since the last cluster")
		assert.Contains(t, status, "2 clear the face-cluster-size of 60 px")
		assert.Contains(t, status, "40 clear "+bar)
	})
	t.Run("ScoreIsTheGate", func(t *testing.T) {
		status := faceClusterStatusFor(query.FaceClusterGates{Unclustered: 40, Recent: 40, SizeOK: 40, ScoreOK: 1, Eligible: 1}, 8, 60, bar, 4)

		assert.Contains(t, status, "40 clear the face-cluster-size of 60 px")
		assert.Contains(t, status, "1 clear "+bar)
	})
	t.Run("OlderMarkersDoNotCountTowardTheTrigger", func(t *testing.T) {
		// Recent is what the worker sees; Unclustered is the whole pool. Reporting the pool as
		// though it were the trigger is what hid the stranded case.
		status := faceClusterStatusFor(query.FaceClusterGates{Unclustered: 90, Recent: 3, SizeOK: 3, ScoreOK: 3, Eligible: 3}, 8, 60, bar, 4)

		assert.Contains(t, status, "has 3")
		assert.NotContains(t, status, "90")
	})
}

func TestConfig_faceClusterScoreFloor(t *testing.T) {
	// The option and the marker queries use opposite conventions for zero, so the mapping is where
	// "let the detector decide" would silently become "no filter at all".
	c := NewConfig(CliTestContext())

	t.Run("Detector", func(t *testing.T) {
		c.options.FaceClusterScore = 0
		assert.Equal(t, face.ClusterScoreAuto, c.faceClusterScoreFloor())
	})
	t.Run("Configured", func(t *testing.T) {
		c.options.FaceClusterScore = 42
		assert.Equal(t, 42, c.faceClusterScoreFloor())
	})
	t.Run("Disabled", func(t *testing.T) {
		c.options.FaceClusterScore = -1
		assert.Equal(t, 0, c.faceClusterScoreFloor())
	})
}

func TestConfig_faceEmbedderStatus(t *testing.T) {
	// The embedder is process-wide, so a test that read it as another test left it would report
	// every model as failing to load - which is the state this is supposed to detect.
	restore := face.ConfiguredModel()
	require.NoError(t, face.ConfigureEmbedder(face.EmbedderSettings{Name: face.ModelNone}))
	t.Cleanup(func() {
		_ = face.ConfigureEmbedder(face.EmbedderSettings{Name: restore, Model: face.FindEmbeddingModel(restore)})
	})

	c := NewConfig(CliTestContext())

	t.Run("Ok", func(t *testing.T) {
		assert.Empty(t, c.faceEmbedderStatus())
	})
	t.Run("Paused", func(t *testing.T) {
		t.Cleanup(face.UnblockEmbeddings)
		face.BlockEmbeddings("12 marker(s) use facenet")

		assert.Equal(t, "Face embeddings are paused, because 12 marker(s) use facenet.", c.faceEmbedderStatus())
	})
}

func TestFaceReportValue(t *testing.T) {
	assert.Equal(t, "sface", faceReportValue("sface"))
	assert.Equal(t, "sface (default)", faceReportValue("sface", "default"))
	assert.Equal(t, "sface (default, paused: 12 markers)", faceReportValue("sface", "default", "paused: 12 markers"))
}
