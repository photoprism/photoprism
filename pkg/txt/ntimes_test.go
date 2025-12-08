package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNTimes(t *testing.T) {
	t.Run("Two", func(t *testing.T) {
		assert.Equal(t, "", NTimes(-2))
	})
	t.Run("One", func(t *testing.T) {
		assert.Equal(t, "", NTimes(-1))
	})
	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, "", NTimes(0))
	})
	t.Run("One", func(t *testing.T) {
		assert.Equal(t, "", NTimes(1))
	})
	t.Run("Num999", func(t *testing.T) {
		assert.Equal(t, "999 times", NTimes(999))
	})
}
