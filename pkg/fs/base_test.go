package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase(t *testing.T) {
	result := Base("/testdata/test.jpg")

	assert.Equal(t, "test", result)
}

func TestBaseAbs(t *testing.T) {
	result := AbsBase("/testdata/test.jpg")

	assert.Equal(t, "/testdata/test", result)
}
