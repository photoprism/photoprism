package fs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestModTime(t *testing.T) {
	t.Run("ValidFilepath", func(t *testing.T) {
		result := ModTime("./testdata/CATYELLOW.jpg")
		assert.NotEmpty(t, result)
	})
	t.Run("InvalidFilePath", func(t *testing.T) {
		result := ModTime("/testdata/Test.jpg")
		assert.NotEmpty(t, result)
		assert.True(t, result.Before(time.Now()))
	})
}
