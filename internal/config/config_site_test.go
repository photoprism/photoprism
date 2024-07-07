package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_BaseUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.BaseUri(""))
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, "", c.BaseUri(""))
	c.options.SiteUrl = "http://foo:2342/foo bar/"
	assert.Equal(t, "/foo%20bar", c.BaseUri(""))
	assert.Equal(t, "/foo%20bar/baz", c.BaseUri("/baz"))
}

func TestConfig_StaticUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/static", c.StaticUri())
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, "/static", c.StaticUri())
	c.options.SiteUrl = "http://foo:2342/foo/"
	assert.Equal(t, "/foo/static", c.StaticUri())
	c.options.CdnUrl = "http://foo:2342/bar"
	assert.Equal(t, "http://foo:2342/bar/foo"+StaticUri, c.StaticUri())
}

func TestConfig_ApiUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ApiUri, c.ApiUri())
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, ApiUri, c.ApiUri())
	c.options.SiteUrl = "http://foo:2342/foo/"
	assert.Equal(t, "/foo"+ApiUri, c.ApiUri())
}

func TestConfig_ContentUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ApiUri, c.ContentUri())
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, ApiUri, c.ContentUri())
	c.options.CdnUrl = "http://foo:2342//"
	assert.Equal(t, "http://foo:2342"+ApiUri, c.ContentUri())
}

func TestConfig_VideoUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, ApiUri, c.VideoUri())
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, ApiUri, c.VideoUri())
	c.options.CdnUrl = "http://foo:2342//"
	c.options.CdnVideo = true
	assert.Equal(t, "http://foo:2342"+ApiUri, c.VideoUri())
	c.options.CdnVideo = false
	assert.Equal(t, ApiUri, c.VideoUri())
}

func TestConfig_SiteUrl(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "http://localhost:2342/", c.SiteUrl())
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, "http://superhost:2342/", c.SiteUrl())
	c.options.SiteUrl = "http://superhost"
	assert.Equal(t, "http://superhost/", c.SiteUrl())
}

func TestConfig_SiteHttps(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		c := NewConfig(CliTestContext())
		assert.False(t, c.SiteHttps())
	})
}

func TestConfig_SiteDomain(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "localhost", c.SiteDomain())
	c.options.SiteUrl = "https://foo.bar.com:2342/"
	assert.Equal(t, "foo.bar.com", c.SiteDomain())
	c.options.SiteUrl = ""
	assert.Equal(t, "localhost", c.SiteDomain())
}

func TestConfig_SitePreview(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, "https://i.photoprism.app/prism?cover=64&style=centered%20dark&caption=none&title=PhotoPrism", c.SitePreview())
	c.options.SitePreview = "http://preview.jpg"
	assert.Equal(t, "http://preview.jpg", c.SitePreview())
	c.options.SitePreview = "preview123.jpg"
	assert.Equal(t, "http://localhost:2342/preview123.jpg", c.SitePreview())
	c.options.SitePreview = "foo/preview123.jpg"
	assert.Equal(t, "http://localhost:2342/foo/preview123.jpg", c.SitePreview())
	c.options.SitePreview = "/foo/preview123.jpg"
	assert.Equal(t, "http://localhost:2342/foo/preview123.jpg", c.SitePreview())
}

func TestConfig_SiteAuthor(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.SiteAuthor())
	c.options.SiteAuthor = "@Jens.Mander"
	assert.Equal(t, "@Jens.Mander", c.SiteAuthor())
	c.options.SiteAuthor = ""
	assert.Equal(t, "", c.SiteAuthor())
}

func TestConfig_SiteTitle(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "PhotoPrism", c.SiteTitle())
	c.options.SiteTitle = "Cats"
	assert.Equal(t, "Cats", c.SiteTitle())
	c.options.SiteTitle = "PhotoPrism"
	assert.Equal(t, "PhotoPrism", c.SiteTitle())
}

func TestConfig_SiteCaption(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.SiteCaption())
	c.options.SiteCaption = "PhotoPrism App"
	assert.Equal(t, "PhotoPrism App", c.SiteCaption())
	c.options.SiteCaption = ""
	assert.Equal(t, "", c.SiteCaption())
}

func TestConfig_SiteDescription(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.SiteDescription())
	c.options.SiteDescription = "My Description!"
	assert.Equal(t, "My Description!", c.SiteDescription())
	c.options.SiteDescription = ""
	assert.Equal(t, "", c.SiteDescription())
}
