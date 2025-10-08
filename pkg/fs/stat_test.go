package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStat(t *testing.T) {
	// Success case
	info, err := Stat("./testdata/test.jpg")
	assert.NoError(t, err)
	assert.False(t, info.IsDir())
	assert.Greater(t, info.Size(), int64(0))

	// Error on empty path
	_, err = Stat("")
	assert.Error(t, err)
}
