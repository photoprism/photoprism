package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	hash := Hash("testdata/test.jpg")
	assert.Equal(t, "516cb1fefbfd9fa66f1db50b94503a480cee30db", hash)
}
