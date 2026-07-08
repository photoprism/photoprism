package projection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromName(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		result := Find("")
		assert.Equal(t, Unknown, result)
	})
	t.Run("Other", func(t *testing.T) {
		result := Find("zdfbhmdflkbhelkthn")
		assert.Equal(t, Other, result)
	})
	t.Run(Equirectangular.String(), func(t *testing.T) {
		result := Find("Equirectangular ")
		assert.Equal(t, Equirectangular, result)
	})
	t.Run(Fisheye.String(), func(t *testing.T) {
		assert.Equal(t, Fisheye, Find("Fisheye"))
	})
	t.Run(DualFisheye.String(), func(t *testing.T) {
		assert.Equal(t, DualFisheye, Find("dual-fisheye"))
	})
}
