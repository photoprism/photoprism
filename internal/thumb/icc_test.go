package thumb

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestGetIccProfile(t *testing.T) {
	t.Run("Exists", func(t *testing.T) {
		profilePath, err := GetIccProfile(IccAdobeRGBCompat)
		require.NoError(t, err)
		assert.True(t, fs.FileExists(profilePath))
	})
	t.Run("Missing", func(t *testing.T) {
		originalPath := IccProfilesPath
		missingDir := t.TempDir()

		IccProfilesPath = missingDir

		t.Cleanup(func() {
			IccProfilesPath = originalPath
		})

		profilePath, err := GetIccProfile(IccAdobeRGBCompat)

		assert.Error(t, err)
		assert.Empty(t, profilePath)
	})
}

func TestIccProfiles(t *testing.T) {
	for _, profile := range IccProfiles {
		t.Run(profile, func(t *testing.T) {
			path, err := GetIccProfile(profile)

			require.NoError(t, err)
			require.True(t, fs.FileExists(path))

			//nolint:gosec // test-only: path is constrained to known profile files in assets
			data, err := os.ReadFile(path)
			require.NoError(t, err)

			if len(data) < 40 {
				t.Fatalf("profile %s too small to contain ICC header", profile)
			}

			assert.Equal(t, []byte{'a', 'c', 's', 'p'}, data[36:40], "profile %s must contain ICC signature", profile)
		})
	}
}
