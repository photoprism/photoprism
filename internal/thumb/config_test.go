package thumb

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestConfig(t *testing.T) {
	t.Run("SamplesPath", func(t *testing.T) {
		t.Logf("samples-path: %s", SamplesPath)
		assert.Equal(t, fs.Abs("../../assets/samples"), SamplesPath)
		assert.True(t, fs.PathExists(SamplesPath))
	})
	t.Run("IccProfilesPath", func(t *testing.T) {
		t.Logf("icc-profiles-path: %s", SamplesPath)
		assert.Equal(t, fs.Abs("../../assets/profiles/icc"), IccProfilesPath)
		// assert.True(t, fs.PathExists(IccProfilesPath))
	})
}
