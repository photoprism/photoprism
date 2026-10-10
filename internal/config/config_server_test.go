package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config/ttl"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/proxy"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/txt"
)

func TestConfig_HttpServerHost(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "0.0.0.0", c.HttpHost())
	c.options.HttpHost = "test"
	assert.Equal(t, "test", c.HttpHost())
	c.options.HttpHost = "unix:/tmp/photoprism.sock"
	assert.Equal(t, "unix:/tmp/photoprism.sock", c.HttpHost())
}

func TestConfig_HttpSocket(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Nil(t, c.HttpSocket())

	t.Run("Empty", func(t *testing.T) {
		c.options.HttpSocket = nil
		c.options.HttpHost = ""

		result := c.HttpSocket()

		assert.Nil(t, result)
	})
	t.Run("Invalid", func(t *testing.T) {
		c.options.HttpSocket = nil
		c.options.HttpHost = "unix:http.sock"

		result := c.HttpSocket()

		assert.Nil(t, result)
	})
	t.Run("UnixHost", func(t *testing.T) {
		c.options.HttpSocket = nil
		c.options.HttpHost = "unix://http.sock"

		result := c.HttpSocket()

		assert.NotNil(t, result)
		assert.Equal(t, scheme.Unix, result.Scheme)
		assert.Contains(t, result.Path, "/internal/config/http.sock")
		assert.False(t, txt.Bool(result.Query().Get("force")))
		assert.Equal(t, fs.ModeSocket, fs.ParseMode(result.Query().Get("mode"), fs.ModeSocket))
	})
	t.Run("UnixPath", func(t *testing.T) {
		c.options.HttpSocket = nil
		c.options.HttpHost = "unix:/var/run/photoprism.sock?force=false&mode=0640"

		result := c.HttpSocket()

		assert.NotNil(t, result)
		assert.Equal(t, scheme.Unix, result.Scheme)
		assert.Equal(t, "/var/run/photoprism.sock", result.Path)
		assert.Equal(t, "false", result.Query().Get("force"))
		assert.False(t, txt.Bool(result.Query().Get("force")))
		assert.Equal(t, os.FileMode(0o640), fs.ParseMode(result.Query().Get("mode"), fs.ModeSocket))
	})
	t.Run("Force", func(t *testing.T) {
		c.options.HttpSocket = nil
		c.options.HttpHost = "unix:/tmp/photoprism.sock?force=true&mode=660"

		result := c.HttpSocket()

		assert.NotNil(t, result)
		assert.Equal(t, scheme.Unix, result.Scheme)
		assert.Equal(t, "/tmp/photoprism.sock", result.Path)
		assert.Equal(t, "true", result.Query().Get("force"))
		assert.True(t, txt.Bool(result.Query().Get("force")))
		assert.Equal(t, os.FileMode(0o660), fs.ParseMode(result.Query().Get("mode"), 0o000))
	})
}

func TestConfig_HttpServerPort(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, 2342, c.HttpPort())
	c.options.HttpPort = 1234
	assert.Equal(t, 1234, c.HttpPort())
}

func TestConfig_HttpServerMode(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, HttpModeProd, c.HttpMode())
	c.options.Debug = true
	assert.Equal(t, HttpModeDebug, c.HttpMode())
	c.options.HttpMode = "test"
	assert.Equal(t, "test", c.HttpMode())
}

func TestConfig_TemplateName(t *testing.T) {
	c := NewConfig(CliTestContext())
	c.initSettings()

	assert.Equal(t, "index.gohtml", c.TemplateName())
	c.settings.Templates.Default = "rainbow.gohtml"
	assert.Equal(t, "rainbow.gohtml", c.TemplateName())
	c.settings.Templates.Default = "xxx"
	assert.Equal(t, "index.gohtml", c.TemplateName())

}

func TestConfig_HttpCompression(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.HttpCompression())

	c.Options().HttpCompression = "  Zstd, GZIP "
	assert.Equal(t, "zstd, gzip", c.HttpCompression())
}

func TestConfig_HttpCompressionPreferences(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Run("Empty", func(t *testing.T) {
		c.Options().HttpCompression = ""
		assert.Empty(t, c.HttpCompressionPreferences())
	})
	t.Run("None", func(t *testing.T) {
		c.Options().HttpCompression = "none"
		assert.Empty(t, c.HttpCompressionPreferences())
	})
	t.Run("LegacyGzip", func(t *testing.T) {
		c.Options().HttpCompression = "gzip"
		assert.Equal(t, []string{"gzip"}, c.HttpCompressionPreferences())
	})
	t.Run("ZstdThenGzip", func(t *testing.T) {
		c.Options().HttpCompression = "zstd,gzip"
		assert.Equal(t, []string{"zstd", "gzip"}, c.HttpCompressionPreferences())
	})
	t.Run("PreservesOrderAndDeduplicates", func(t *testing.T) {
		c.Options().HttpCompression = "gzip, zstd, gzip"
		assert.Equal(t, []string{"gzip", "zstd"}, c.HttpCompressionPreferences())
	})
	t.Run("DropsUnknownTokens", func(t *testing.T) {
		c.Options().HttpCompression = "br, zstd, gzip"
		assert.Equal(t, []string{"zstd", "gzip"}, c.HttpCompressionPreferences())
	})
	t.Run("WhitespaceAndCaseInsensitive", func(t *testing.T) {
		c.Options().HttpCompression = "  ZSTD ,  GZip  "
		assert.Equal(t, []string{"zstd", "gzip"}, c.HttpCompressionPreferences())
	})
	t.Run("IdentityIsOff", func(t *testing.T) {
		c.Options().HttpCompression = "identity"
		assert.Empty(t, c.HttpCompressionPreferences())
	})
}

func TestConfig_HttpCompressionUnknown(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Run("None", func(t *testing.T) {
		c.Options().HttpCompression = "zstd, gzip"
		assert.Empty(t, c.HttpCompressionUnknown())
	})
	t.Run("ReportsUnknown", func(t *testing.T) {
		c.Options().HttpCompression = "br, zstd, deflate, gzip"
		assert.Equal(t, []string{"br", "deflate"}, c.HttpCompressionUnknown())
	})
	t.Run("DeduplicatesUnknown", func(t *testing.T) {
		c.Options().HttpCompression = "br, br, deflate"
		assert.Equal(t, []string{"br", "deflate"}, c.HttpCompressionUnknown())
	})
	t.Run("EmptyDoesNotReportUnknown", func(t *testing.T) {
		c.Options().HttpCompression = ""
		assert.Empty(t, c.HttpCompressionUnknown())
	})
}

func TestConfig_HttpHeaderTimeout(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, DefaultHttpHeaderTimeout, c.HttpHeaderTimeout())

	c.Options().HttpHeaderTimeout = 17 * time.Second
	assert.Equal(t, 17*time.Second, c.HttpHeaderTimeout())

	c.Options().HttpHeaderTimeout = -1 * time.Second
	assert.Equal(t, DefaultHttpHeaderTimeout, c.HttpHeaderTimeout())
}

func TestConfig_HttpHeaderBytes(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, DefaultHttpHeaderBytes, c.HttpHeaderBytes())

	c.Options().HttpHeaderBytes = 2048
	assert.Equal(t, 2048, c.HttpHeaderBytes())

	c.Options().HttpHeaderBytes = 0
	assert.Equal(t, DefaultHttpHeaderBytes, c.HttpHeaderBytes())
}

func TestConfig_HttpIdleTimeout(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, DefaultHttpIdleTimeout, c.HttpIdleTimeout())

	c.Options().HttpIdleTimeout = 2 * time.Minute
	assert.Equal(t, 2*time.Minute, c.HttpIdleTimeout())

	c.Options().HttpIdleTimeout = -1 * time.Second
	assert.Equal(t, DefaultHttpIdleTimeout, c.HttpIdleTimeout())
}

func TestConfig_HttpCachePublic(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.HttpCachePublic())
	c.Options().CdnUrl = "https://cdn.com/"
	assert.True(t, c.HttpCachePublic())
	c.Options().CdnUrl = ""
	assert.False(t, c.HttpCachePublic())
	c.Options().HttpCachePublic = true
	assert.True(t, c.HttpCachePublic())
	c.Options().HttpCachePublic = false
	assert.False(t, c.HttpCachePublic())
}

func TestConfig_HttpCacheMaxAge(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ttl.Duration(2592000), c.HttpCacheMaxAge())
	c.Options().HttpCacheMaxAge = 23
	assert.Equal(t, ttl.Duration(23), c.HttpCacheMaxAge())
	c.Options().HttpCacheMaxAge = 41536000
	assert.Equal(t, ttl.CacheMaxAge, c.HttpCacheMaxAge())
	c.Options().HttpCacheMaxAge = 0
	assert.Equal(t, ttl.Duration(2592000), c.HttpCacheMaxAge())
}

func TestConfig_HttpVideoMaxAge(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ttl.CacheVideo, c.HttpVideoMaxAge())
	c.Options().HttpVideoMaxAge = 23
	assert.Equal(t, ttl.Duration(23), c.HttpVideoMaxAge())
	c.Options().HttpVideoMaxAge = 41536000
	assert.Equal(t, ttl.CacheMaxAge, c.HttpVideoMaxAge())
	c.Options().HttpVideoMaxAge = 0
	assert.Equal(t, ttl.CacheVideo, c.HttpVideoMaxAge())
}

func TestConfig_TrustedProxies(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		var defaults []string

		for _, f := range Flags {
			if ssf, ok := f.Flag.(*cli.StringSliceFlag); ok && ssf.Name == "trusted-proxy" {
				defaults = ssf.Value.Value()
			}
		}

		assert.Equal(t, []string{"172.16.0.0/12", "127.0.0.0/8", "::1"}, defaults)

		// The documented default is a list that can be copied into PHOTOPRISM_TRUSTED_PROXY as it is.
		for _, f := range Flags {
			if f.Name() == "trusted-proxy" {
				assert.Equal(t, "172.16.0.0/12, 127.0.0.0/8, ::1", f.Default())
			}
		}

		c := NewConfig(CliTestContext())
		c.options.TrustedProxies = defaults
		assert.Equal(t, "172.16.0.0/12, 127.0.0.0/8, ::1", c.TrustedProxy())

		trusted := proxy.ParseTrustedProxies(c.TrustedProxies())
		assert.Len(t, trusted, 3)

		for _, ip := range []string{"127.0.0.1", "127.0.0.2", "::1", "::ffff:127.0.0.1", "172.18.0.2", "172.31.255.254"} {
			assert.True(t, proxy.TrustedIP(ip, trusted), ip)
		}

		for _, ip := range []string{"203.0.113.5", "10.0.0.1", "192.168.1.1", "fd00::1", "2001:db8::1", "::"} {
			assert.False(t, proxy.TrustedIP(ip, trusted), ip)
		}
	})
}
