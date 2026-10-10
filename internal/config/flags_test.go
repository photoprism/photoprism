package config

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
)

// TestLabelModelDocDefault verifies the published default and supported modes.
func TestLabelModelDocDefault(t *testing.T) {
	for _, flag := range Flags {
		if flag.Name() != "labels-model" {
			continue
		}

		assert.Equal(t, string(classify.ModelAuto), flag.Default())
		assert.Contains(t, flag.Usage(), "auto, none")
		assert.NotContains(t, flag.Usage(), string(classify.DefaultModelName()))
		return
	}

	t.Fatal("labels-model flag not found")
}

// TestNSFWModelDocDefault verifies the published default and supported modes.
func TestNSFWModelDocDefault(t *testing.T) {
	for _, flag := range Flags {
		if flag.Name() != "nsfw-model" {
			continue
		}
		assert.Equal(t, string(nsfw.ModelAuto), flag.Default())
		assert.Contains(t, flag.Usage(), "auto, none, labels")
		assert.NotContains(t, flag.Usage(), string(nsfw.DefaultModelName()))
		return
	}
	t.Fatal("nsfw-model flag not found")
}

// TestFaceDocDefaults pins that the face options publish the number that actually applies on a
// default install. The generated end-user reference and "--help" both print a numeric option
// with no flag default as 0, and a word like "detector" tells a reader deciding what to set
// nothing at all.
func TestFaceDocDefaults(t *testing.T) {
	defaults := make(map[string]string, len(Flags))

	for _, flag := range Flags {
		defaults[flag.Name()] = flag.Default()
	}

	t.Run("ResolvedModels", func(t *testing.T) {
		assert.Equal(t, face.DefaultDetectorName(), defaults["face-detector"])
		assert.Equal(t, face.DefaultModelName(), defaults["face-model"])
	})
	// Derived from the registry, not restated: a calibration that moves would otherwise be
	// published from two places and only one of them would be updated.
	d := face.DefaultDetector()
	require.NotNil(t, d)

	t.Run("Scores", func(t *testing.T) {
		assert.Equal(t, strconv.Itoa(d.MinScore), defaults["face-score"])
		assert.Equal(t, strconv.Itoa(d.ClusterScore), defaults["face-cluster-score"])
	})
	t.Run("MigrationFloors", func(t *testing.T) {
		// The migration's own floors, which the team tunes without a rebuild.
		assert.Equal(t, strconv.Itoa(face.MinSizeThreshold), defaults["face-migrate-size"])
		assert.Equal(t, strconv.Itoa(d.MigrateScore), defaults["face-migrate-score"])
	})
	t.Run("CalibratedDistances", func(t *testing.T) {
		m := face.DefaultModel()
		require.NotNil(t, m)

		assert.Equal(t, strconv.FormatFloat(m.ClusterDist, 'g', -1, 64), defaults["face-cluster-dist"])
		assert.Equal(t, strconv.FormatFloat(m.ClusterRadius, 'g', -1, 64), defaults["face-cluster-radius"])
		assert.Equal(t, strconv.FormatFloat(m.MatchDist, 'g', -1, 64), defaults["face-match-dist"])
	})
	t.Run("FlatDistances", func(t *testing.T) {
		// The two gaps and the assignment margin do not follow the model, so "--help" states one
		// number for all of them rather than the default model's.
		assert.Equal(t, strconv.FormatFloat(face.CollisionDistDefault, 'g', -1, 64), defaults["face-collision-dist"])
		assert.Equal(t, strconv.FormatFloat(face.EpsilonDefault, 'g', -1, 64), defaults["face-epsilon-dist"])
		assert.Equal(t, strconv.FormatFloat(face.MatchMarginDefault, 'g', -1, 64), defaults["face-match-margin"])
	})
	t.Run("NoneIsZero", func(t *testing.T) {
		// The thread counts are derived from the CPU, so there is no number to publish and
		// "auto" is the honest answer.
		assert.Equal(t, "auto", defaults["face-detector-threads"])
		assert.Equal(t, "auto", defaults["face-model-threads"])
	})
}

func TestFaceDocDefault(t *testing.T) {
	assert.Equal(t, "0.85", faceDocDefault(0.85))
	assert.Equal(t, "9", faceDocDefault(9))
	assert.Empty(t, faceDocDefault(0), "an absent value must not be published as a setting")
	assert.Empty(t, faceDocDefault(-1))
}

func TestFaceModelDocDefault(t *testing.T) {
	assert.Equal(t, "0.72", faceModelDocDefault(func(m *face.EmbeddingModel) float64 { return m.ClusterDist }))
	assert.Empty(t, faceModelDocDefault(func(m *face.EmbeddingModel) float64 { return 0 }))
}

// TestFlagsSecretAnnotation pins that every flag carrying credential material sets Secret.
// A new flag matching SecretFlagPatterns must either set it or join the reviewed exceptions.
func TestFlagsSecretAnnotation(t *testing.T) {
	missing, matched := FlagsWithoutSecret(Flags)

	assert.Empty(t, missing, "these flags match a credential pattern, so they must set Secret or be a reviewed exception")

	// Guard the scan itself: without this a rename that stops matching every pattern would leave
	// the check asserting nothing and still reporting success.
	assert.GreaterOrEqual(t, matched, 12, "expected the known credential flags to match; the patterns may be stale")
}

func TestFlagsWithoutSecret(t *testing.T) {
	flags := CliFlags{
		{Flag: &cli.StringFlag{Name: "example-password", EnvVars: EnvVars("EXAMPLE_PASSWORD")}, Secret: true},
		{Flag: &cli.StringFlag{Name: "example-dsn", EnvVars: EnvVars("EXAMPLE_DSN")}},
		{Flag: &cli.StringFlag{Name: "example-path", EnvVars: EnvVars("EXAMPLE_TOKEN_PATH")}},
		{Flag: &cli.StringFlag{Name: "tls-key", EnvVars: EnvVars("TLS_KEY")}},
		{Flag: &cli.StringFlag{Name: "example-name", EnvVars: EnvVars("EXAMPLE_NAME")}},
	}

	t.Run("Missing", func(t *testing.T) {
		missing, matched := FlagsWithoutSecret(flags)
		assert.Equal(t, []string{"example-dsn", "example-path"}, missing)
		assert.Equal(t, 3, matched)
	})
	t.Run("Exceptions", func(t *testing.T) {
		missing, matched := FlagsWithoutSecret(flags, "example-path")
		assert.Equal(t, []string{"example-dsn"}, missing)
		assert.Equal(t, 2, matched)
	})
	t.Run("Empty", func(t *testing.T) {
		missing, matched := FlagsWithoutSecret(nil)
		assert.Empty(t, missing)
		assert.Zero(t, matched)
	})
}
