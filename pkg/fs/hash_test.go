package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	t.Run("existing image", func(t *testing.T) {
		hash := Hash("testdata/test.jpg")
		assert.Equal(t, "9e7c20fd7eec6dfb0a2f1e006113cdf8fc00d2c6790071ae0b13063d9ea9324b", hash)
	})
	t.Run("not existing image", func(t *testing.T) {
		hash := Hash("testdata/xxx.jpg")
		assert.Equal(t, "", hash)
	})
}

func TestChecksum(t *testing.T) {
	t.Run("existing image", func(t *testing.T) {
		hash := Checksum("testdata/test.jpg")
		assert.Equal(t, "5239d867", hash)
	})
	t.Run("not existing image", func(t *testing.T) {
		hash := Checksum("testdata/xxx.jpg")
		assert.Equal(t, "", hash)
	})
}

func TestIsHash(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		assert.Equal(t, false, IsHash(""))
	})
	t.Run("true", func(t *testing.T) {
		assert.Equal(t, true, IsHash("516cb1fefbfd9fa66f1db50b94503a480cee30db"))

	})
}
