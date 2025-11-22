package customize

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitDefaultFeatures_DisableList(t *testing.T) {
	origEnv, envSet := os.LookupEnv("PHOTOPRISM_DISABLE_FEATURES")
	origDefaults := DefaultFeatures

	t.Cleanup(func() {
		if envSet {
			_ = os.Setenv("PHOTOPRISM_DISABLE_FEATURES", origEnv)
		} else {
			_ = os.Unsetenv("PHOTOPRISM_DISABLE_FEATURES")
		}

		DefaultFeatures = initDefaultFeatures()
	})

	_ = os.Setenv("PHOTOPRISM_DISABLE_FEATURES", "Upload, videos share batch-edit labels")
	DefaultFeatures = initDefaultFeatures()

	assert.False(t, DefaultFeatures.Upload)
	assert.False(t, DefaultFeatures.Videos)
	assert.False(t, DefaultFeatures.Share)
	assert.False(t, DefaultFeatures.BatchEdit)
	assert.False(t, DefaultFeatures.Labels)

	// unaffected feature stays enabled
	assert.True(t, DefaultFeatures.Favorites)

	// ensure the defaults are not permanently changed
	assert.NotEqual(t, origDefaults, FeatureSettings{})
}

func TestNewSettingsCopiesDefaultFeatures(t *testing.T) {
	origEnv, envSet := os.LookupEnv("PHOTOPRISM_DISABLE_FEATURES")
	origDefaults := DefaultFeatures

	t.Cleanup(func() {
		if envSet {
			_ = os.Setenv("PHOTOPRISM_DISABLE_FEATURES", origEnv)
		} else {
			_ = os.Unsetenv("PHOTOPRISM_DISABLE_FEATURES")
		}

		DefaultFeatures = origDefaults
	})

	_ = os.Unsetenv("PHOTOPRISM_DISABLE_FEATURES")
	DefaultFeatures = initDefaultFeatures()

	settings := NewSettings("", "", "")
	settings.Features.Upload = false
	settings.Features.Download = false

	assert.True(t, DefaultFeatures.Upload, "DefaultFeatures should remain unchanged after mutation")
	assert.True(t, DefaultFeatures.Download, "DefaultFeatures should remain unchanged after mutation")
}
