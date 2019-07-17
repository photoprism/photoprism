package util

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
