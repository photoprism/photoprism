package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOrientation(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, KeepOrientation, ParseOrientation("foo", KeepOrientation))
		assert.Equal(t, ResetOrientation, ParseOrientation("foo", ResetOrientation))
		assert.Equal(t, ResetOrientation, ParseOrientation("", ResetOrientation))
		assert.Equal(t, "", ParseOrientation("", ""))
	})
	t.Run("Keep", func(t *testing.T) {
		result := ParseOrientation("KeEp", ResetOrientation)
		assert.Equal(t, KeepOrientation, result)
	})
	t.Run("Reset", func(t *testing.T) {
		result := ParseOrientation("reset", KeepOrientation)
		assert.Equal(t, ResetOrientation, result)
	})
}
