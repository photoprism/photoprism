package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNTimes(t *testing.T) {
	t.Run("-2", func(t *testing.T) {
		assert.Equal(t, "", NTimes(-2))
	})
	t.Run("-1", func(t *testing.T) {
		assert.Equal(t, "", NTimes(-1))
	})
	t.Run("0", func(t *testing.T) {
		assert.Equal(t, "", NTimes(0))
	})
	t.Run("1", func(t *testing.T) {
		assert.Equal(t, "", NTimes(1))
	})
	t.Run("999", func(t *testing.T) {
		assert.Equal(t, "999 times", NTimes(999))
	})
}
