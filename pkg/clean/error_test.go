package clean

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogError(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "no error", Error(nil))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "unknown error", Error(errors.New("")))
	})
	t.Run("Simple", func(t *testing.T) {
		assert.Equal(t, "simple", Error(errors.New("simple")))
	})
	t.Run("Spaces", func(t *testing.T) {
		assert.Equal(t, "the quick brown fox", Error(errors.New("the quick brown fox")))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, "??https://?host?:?port?/?path??", Error(errors.New("${https://<host>:<port>/<path>}")))
	})
}
