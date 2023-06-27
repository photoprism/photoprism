package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	t.Run("existing image", func(t *testing.T) {
		hash := Hash("testdata/test.jpg")
		assert.Equal(t, "516cb1fefbfd9fa66f1db50b94503a480cee30db", hash)
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
		assert.Equal(t, false, IsHash("123"))
	})
	t.Run("true", func(t *testing.T) {
		assert.Equal(t, true, IsHash("516cb1fefbfd9fa66f1db50b94503a480cee30db"))
	})
}
