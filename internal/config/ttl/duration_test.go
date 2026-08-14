package ttl

import (
	"testing"
	"time"

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

func TestDuration_Duration(t *testing.T) {
	assert.Equal(t, time.Hour, Duration(3600).Duration())
	assert.Equal(t, time.Duration(0), Duration(0).Duration())
	assert.Equal(t, -time.Second, Duration(-1).Duration())
}
