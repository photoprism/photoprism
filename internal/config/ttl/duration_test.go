package ttl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDuration_Int(t *testing.T) {
	t.Run("Hour", func(t *testing.T) {
		assert.Equal(t, 3600, Duration(3600).Int())
	})
	t.Run("Month", func(t *testing.T) {
		assert.Equal(t, 2592000, Duration(2592000).Int())
	})
}

func TestDuration_String(t *testing.T) {
	t.Run("Hour", func(t *testing.T) {
		assert.Equal(t, "3600", Duration(3600).String())
	})
	t.Run("Month", func(t *testing.T) {
		assert.Equal(t, "2592000", Duration(2592000).String())
	})
}
