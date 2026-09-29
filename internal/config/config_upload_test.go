package config

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_UploadNSFW(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.UploadNSFW())
}

func TestConfig_UploadAllow(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.UploadAllow = "jpg, PNG,pdf"

	assert.Equal(t, "jpg, pdf, png", c.UploadAllow().String())

	c.options.UploadAllow = ""

	assert.Len(t, c.UploadAllow(), 0)
	assert.Equal(t, "", c.UploadAllow().String())
}

// TestConfig_UploadLimit checks that the upload limit keeps valid values, is disabled at 0 or below, and is clamped at the maximum.
func TestConfig_UploadLimit(t *testing.T) {
	c := NewConfig(CliTestContext())

	for _, tc := range []struct {
		name  string
		limit int
		want  int
		bytes int64
	}{
		{name: "Valid", limit: 800, want: 800, bytes: 838860800},
		{name: "Zero", limit: 0, want: -1, bytes: -1},
		{name: "Disabled", limit: -1, want: -1, bytes: -1},
		{name: "Large", limit: 100001, want: 100001, bytes: int64(100001) * 1024 * 1024},
		{name: "AboveMaximum", limit: math.MaxInt, want: MaxSizeLimit, bytes: int64(MaxSizeLimit) * 1024 * 1024},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c.options.UploadLimit = tc.limit
			assert.Equal(t, tc.want, c.UploadLimit())
			assert.Equal(t, tc.bytes, c.UploadLimitBytes())
		})
	}
}
