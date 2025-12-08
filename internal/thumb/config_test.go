package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestConfig(t *testing.T) {
	t.Run("ExamplesPath", func(t *testing.T) {
		t.Logf("examples-path: %s", ExamplesPath)
		assert.Equal(t, fs.Abs("../../assets/examples"), ExamplesPath)
		assert.True(t, fs.PathExists(ExamplesPath))
	})
	t.Run("IccProfilesPath", func(t *testing.T) {
		t.Logf("icc-profiles-path: %s", ExamplesPath)
		assert.Equal(t, fs.Abs("../../assets/profiles/icc"), IccProfilesPath)
		// assert.True(t, fs.PathExists(IccProfilesPath))
	})
}
