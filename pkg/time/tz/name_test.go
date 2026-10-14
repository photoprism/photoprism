package tz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	t.Run("UTC", func(t *testing.T) {
		assert.Equal(t, UTC, Name("utc"))
		assert.Equal(t, UTC, Name("UTC"))
		assert.Equal(t, UTC, Name("Z"))
		assert.Equal(t, UTC, Name("Zulu"))
		assert.Equal(t, UTC, Name("UTC+0"))
		assert.Equal(t, UTC, Name("UTC-0"))
		assert.Equal(t, UTC, Name("UTC+00"))
		assert.Equal(t, UTC, Name("UTC-00"))
		assert.Equal(t, UTC, Name("UTC+00:00"))
		assert.Equal(t, UTC, Name("UTC-00:00"))
		assert.Equal(t, UTC, Name("Etc/UTC+0"))
	})
	t.Run("Local", func(t *testing.T) {
		assert.Equal(t, Local, Name(""))
		assert.Equal(t, Local, Name("local"))
		assert.Equal(t, Local, Name("Local"))
		assert.Equal(t, Local, Name("LOCAL"))
	})
	t.Run("GMT", func(t *testing.T) {
		assert.Equal(t, GMT, Name("gmt"))
		assert.Equal(t, GMT, Name("GMT"))
		assert.Equal(t, GMT, Name("GMT+0"))
		assert.Equal(t, AsiaKathmandu, Name("Etc/GMT+05:45"))
		assert.Equal(t, AsiaKathmandu, Name("GMT+05:45"))
		assert.Equal(t, "UTC+1", Name("Etc/GMT+01:00"))
		assert.Equal(t, "UTC+1", Name("GMT+1"))
		assert.Equal(t, "UTC+2", Name("GMT+2"))
		assert.Equal(t, "UTC-10", Name("Etc/GMT+10"))
		assert.Equal(t, "UTC-3", Name("GMT-3"))
		assert.Equal(t, "UTC+12", Name("Etc/GMT-12"))
	})
}
