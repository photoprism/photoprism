package fsutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	assert.True(t, Exists("./_fixtures/test.jpg"))
	assert.False(t, Exists("./foo.jpg"))
}

func TestExpandedFilename(t *testing.T) {
	filename := ExpandedFilename("./_fixtures/test.jpg")

	assert.IsType(t, "", filename)

	t.Logf("ExpandedFilename: %s", filename)
}
