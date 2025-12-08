package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	t.Run("ExistingImage", func(t *testing.T) {
		hash := Hash("testdata/test.jpg")
		assert.Equal(t, "516cb1fefbfd9fa66f1db50b94503a480cee30db", hash)
	})
	t.Run("NotExistingImage", func(t *testing.T) {
		hash := Hash("testdata/xxx.jpg")
		assert.Equal(t, "", hash)
	})
}

func TestChecksum(t *testing.T) {
	t.Run("ExistingImage", func(t *testing.T) {
		hash := Checksum("testdata/test.jpg")
		assert.Equal(t, "5239d867", hash)
	})
	t.Run("NotExistingImage", func(t *testing.T) {
		hash := Checksum("testdata/xxx.jpg")
		assert.Equal(t, "", hash)
	})
}

func TestIsHash(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, false, IsHash(""))
		assert.Equal(t, false, IsHash("123"))
	})
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, true, IsHash("516cb1fefbfd9fa66f1db50b94503a480cee30db"))
	})
}
