package fs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaseInsensitive(t *testing.T) {
	t.Run("temp", func(t *testing.T) {
		if result, err := CaseInsensitive(os.TempDir()); err != nil {
			t.Fatal(err)
		} else {
			t.Logf("tmp fs case-insensitive: %t", result)
		}
	})
}

func TestIgnoreCase(t *testing.T) {
	assert.False(t, ignoreCase)
	IgnoreCase()
	assert.True(t, ignoreCase)
	ignoreCase = false
	assert.False(t, ignoreCase)
}
