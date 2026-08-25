package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/dsn"

	"github.com/photoprism/photoprism/internal/ai/face"
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

func TestConfig_FaceReport(t *testing.T) {
	m := NewConfig(CliTestContext())
	rows, cols := m.FaceReport()

	assert.Equal(t, []string{"Name", "Value"}, cols)
	assert.GreaterOrEqual(t, len(rows), 1)

	values := make(map[string]string, len(rows))
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		values[row[0]] = row[1]
	}

	// Spot-check that the core face-related rows are present. They are named without the prefix
	// the subcommand already carries, which is what tells this report apart from "show config".
	expected := []string{
		"Enabled",
		"Vision Config",
		"Detector",
		"Detector Path",
		"Detector Threads",
		"Model",
		"Model Path",
		"Model Threads",
		"Engine",
		"Schedule",
		"Min Size",
		"Retry Size",
		"Min Score",
		"Overlap",
		"Cluster Size",
		"Cluster Score",
		"Cluster Core",
		"Cluster Distance",
		"Cluster Radius",
		"Collision Distance",
		"Epsilon Distance",
		"Match Distance",
	}

	for _, name := range expected {
		_, ok := values[name]
		assert.True(t, ok, "FaceReport missing %q", name)
	}

	// Non-face options must not leak into the focused report.
	unexpected := []string{"originals-path", "auth-mode", "ffmpeg-bin", "site-url"}
	for _, name := range unexpected {
		_, ok := values[name]
		assert.False(t, ok, "FaceReport unexpectedly includes %q", name)
	}
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
		c := newSFaceTestConfig(t)

		assert.Equal(t, face.ModelDisplayName(face.ModelSFace), c.faceModelReport())
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

		assert.Equal(t, face.ModelDisplayName(face.ModelSFace)+" (paused: 12 marker(s) use facenet)", c.faceModelReport())
	})
	t.Run("DetectNamesTheLibraryModel", func(t *testing.T) {
		// "faces status" connects, so it is the one report where the detected model is the one
		// the library actually holds rather than the one a fresh install would start with.
		c := TestConfig()
		setting := c.options.FaceModel
		t.Cleanup(func() { c.options.FaceModel = setting })
		c.options.FaceModel = face.ModelDetect

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
		// The faces report is read rather than parsed, so it names the detector the way an
		// operator would. "show config" keeps the identifier.
		c.options.FaceDetector = face.DetectorYuNet
		assert.Equal(t, face.DetectorDisplayName(face.DetectorYuNet), c.faceDetectorReport())
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
			"face-detector",
			"face-detector-path",
			"face-detector-threads",
			"face-model",
			"face-model-path",
			"face-model-threads",
			"face-run",
			"face-engine",
			"face-size",
			"face-size-retry",
			"face-score",
			"face-overlap",
			"face-cluster-size",
			"face-cluster-score",
			"face-cluster-core",
			"face-cluster-dist",
			"face-cluster-radius",
			"face-collision-dist",
			"face-epsilon-dist",
			"face-match-dist",
		}

		rows := c.faceConfigRows(false)
		flags := make([]string, 0, len(rows))

		for _, row := range rows {
			flags = append(flags, row.Flag)
			assert.NotEmpty(t, row.Label, row.Flag)
			assert.NotContains(t, row.Label, "face-", "the faces report drops the prefix its subcommand carries")
		}

		assert.Equal(t, want, flags)
	})
	t.Run("EveryFlagIsRegistered", func(t *testing.T) {
		// A row naming an option that no longer exists reads as a setting an operator can change.
		// The two derived paths and the run schedule have no flag of their own.
		derived := map[string]bool{"face-detector-path": true, "face-model-path": true}

		registered := make(map[string]bool, len(Flags))

		for _, flag := range Flags {
			registered[flag.Name()] = true
		}

		for _, row := range c.faceConfigRows(false) {
			if derived[row.Flag] {
				continue
			}

			assert.True(t, registered[row.Flag], row.Flag)
		}
	})
	t.Run("QuietWithoutVerbose", func(t *testing.T) {
		// "show config" runs without a database, so it can neither detect the model nor see that
		// one is blocked. Stating a resolution it cannot check is how it came to report "ok" on
		// a paused instance.
		for _, row := range c.faceConfigRows(false) {
			if row.Flag == "face-engine" {
				continue // The deprecation marker is a property of the option, not a resolution.
			}

			assert.NotContains(t, row.Value, "(", row.Flag)
		}
	})
}

func TestConfig_faceScoreReport(t *testing.T) {
	t.Run("Detector", func(t *testing.T) {
		// A cutoff of zero is what the raw option holds in the ordinary case, and it reads as
		// "no detection is filtered", which is the opposite of what it means.
		c := NewConfig(CliTestContext())
		c.options.FaceScore = 0

		assert.NotEqual(t, "0", c.faceScoreReport(false))
		assert.Contains(t, c.faceScoreReport(true), c.FaceDetector())
	})
	t.Run("Configured", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceScore = 42

		assert.Equal(t, "42", c.faceScoreReport(false))
		assert.Equal(t, "42", c.faceScoreReport(true), "an operator's own value has no source to name")
	})
	t.Run("Disabled", func(t *testing.T) {
		// A cutoff that was switched off must not be attributed to a detector's calibration.
		c := NewConfig(CliTestContext())
		c.options.FaceScore = -1

		assert.Equal(t, "-1", c.faceScoreReport(true))
	})
}

func TestConfig_faceClusterScoreReport(t *testing.T) {
	t.Run("Detector", func(t *testing.T) {
		// Zero is what the raw option holds in the ordinary case, and it reads as "every face
		// may cluster" - which is the opposite of what it means.
		c := NewConfig(CliTestContext())
		c.options.FaceClusterScore = 0

		assert.NotEqual(t, "0", c.faceClusterScoreReport(false))
		assert.Contains(t, c.faceClusterScoreReport(true), c.FaceDetector())
	})
	t.Run("Configured", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		c.options.FaceClusterScore = 42

		assert.Equal(t, "42", c.faceClusterScoreReport(false))
		assert.Equal(t, "42", c.faceClusterScoreReport(true), "an operator's own bar has no source to name")
	})
	t.Run("Disabled", func(t *testing.T) {
		// A bar that was switched off is not a detector calibration either.
		c := NewConfig(CliTestContext())
		c.options.FaceClusterScore = -1

		assert.Equal(t, "-1", c.faceClusterScoreReport(true))
	})
}

func TestFaceReportValue(t *testing.T) {
	assert.Equal(t, "sface", faceReportValue("sface"))
	assert.Equal(t, "sface (default)", faceReportValue("sface", "default"))
	assert.Equal(t, "sface (default, paused: 12 markers)", faceReportValue("sface", "default", "paused: 12 markers"))
}
