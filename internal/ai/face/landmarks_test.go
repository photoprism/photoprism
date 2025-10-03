package face

import (
	"testing"

	pigo "github.com/esimov/pigo/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadCascadeDir(t *testing.T) {
	t.Parallel()

	plc := pigo.NewPuplocCascade()

	cascades, err := ReadCascadeDir(plc, "cascade/lps")
	require.NoError(t, err)
	require.NotNil(t, cascades)

	expected := []string{"lp312", "lp38", "lp42", "lp44", "lp46", "lp81", "lp82", "lp84", "lp93"}

	for _, name := range expected {
		entries, ok := cascades[name]
		require.Truef(t, ok, "expected cascade %s", name)
		require.NotEmptyf(t, entries, "cascade %s should contain entries", name)

		for _, entry := range entries {
			require.NotNil(t, entry)
			assert.NotNilf(t, entry.PuplocCascade, "expected cascade %s to include unpacked data", name)
			assert.Nilf(t, entry.error, "expected cascade %s to unpack without error", name)
		}
	}
}

func TestReadCascadeDirMissing(t *testing.T) {
	t.Parallel()

	plc := pigo.NewPuplocCascade()

	cascades, err := ReadCascadeDir(plc, "cascade/missing")
	require.Error(t, err)
	assert.EqualError(t, err, "the cascade directory is empty")
	assert.NotNil(t, cascades)
	assert.Len(t, cascades, 0)
}
