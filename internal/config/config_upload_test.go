package config

import (
	"flag"
	"math"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
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

func TestConfig_UploadMaxAge(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, DefaultUploadMaxAge, c.UploadMaxAge())
		c.options.UploadMaxAge = 0
		assert.Equal(t, int64(604800), c.UploadMaxAge())
	})
	t.Run("Disabled", func(t *testing.T) {
		c.options.UploadMaxAge = -1
		assert.Equal(t, int64(-1), c.UploadMaxAge())
		c.options.UploadMaxAge = -5
		assert.Equal(t, int64(-1), c.UploadMaxAge())
	})
	t.Run("Minimum", func(t *testing.T) {
		c.options.UploadMaxAge = 1
		assert.Equal(t, int64(86400), c.UploadMaxAge())
		c.options.UploadMaxAge = MinUploadMaxAge - 1
		assert.Equal(t, MinUploadMaxAge, c.UploadMaxAge())
	})
	t.Run("Custom", func(t *testing.T) {
		c.options.UploadMaxAge = MinUploadMaxAge
		assert.Equal(t, int64(86400), c.UploadMaxAge())
		c.options.UploadMaxAge = 259200
		assert.Equal(t, int64(259200), c.UploadMaxAge())
		c.options.UploadMaxAge = MaxUploadMaxAge
		assert.Equal(t, int64(3153600000), c.UploadMaxAge())
	})
	t.Run("Maximum", func(t *testing.T) {
		for _, value := range []int64{MaxUploadMaxAge + 1, 9999999999, 18446744074, math.MaxInt64} {
			c.options.UploadMaxAge = value
			assert.Equal(t, MaxUploadMaxAge, c.UploadMaxAge(), value)
			assert.Positive(t, time.Duration(c.UploadMaxAge())*time.Second, value)
		}
	})
	t.Run("Report", func(t *testing.T) {
		c.options.UploadMaxAge = 3600
		rows, _ := c.Report()
		assert.Contains(t, rows, []string{"upload-maxage", "86400"})
	})
	t.Run("Flag", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_UPLOAD_MAXAGE", "")
		require.NoError(t, os.Unsetenv("PHOTOPRISM_UPLOAD_MAXAGE"))
		var found *cli.Int64Flag
		for _, f := range Flags {
			if int64Flag, ok := f.Flag.(*cli.Int64Flag); ok && int64Flag.Name == "upload-maxage" {
				found = int64Flag
			}
		}
		require.NotNil(t, found)
		assert.Equal(t, int64(604800), found.Value)
		assert.Equal(t, []string{"PHOTOPRISM_UPLOAD_MAXAGE"}, found.EnvVars)
		for _, tc := range []struct {
			args     []string
			expected int64
		}{{nil, DefaultUploadMaxAge}, {[]string{"--upload-maxage=-1"}, -1}, {[]string{"--upload-maxage=3600"}, MinUploadMaxAge}, {[]string{"--upload-maxage=9999999999"}, MaxUploadMaxAge}} {
			set := flag.NewFlagSet("test", flag.ContinueOnError)
			require.NoError(t, found.Apply(set))
			require.NoError(t, set.Parse(tc.args))
			opt := &Options{}
			require.NoError(t, opt.ApplyCliContext(cli.NewContext(cli.NewApp(), set, nil)))
			assert.Equal(t, tc.expected, (&Config{options: opt}).UploadMaxAge(), tc.args)
		}
	})
}
