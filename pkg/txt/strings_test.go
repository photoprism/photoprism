package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	t.Run("not empty", func(t *testing.T) {
		assert.Equal(t, true, Bool("Browse your life in pictures"))
	})
	t.Run("no", func(t *testing.T) {
		assert.Equal(t, false, Bool("no"))
	})
	t.Run("false", func(t *testing.T) {
		assert.Equal(t, false, Bool("false"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, false, Bool(""))
	})
}
