package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/header"
)

func TestConfig_CdnUrl(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.options.SiteUrl)
	assert.Equal(t, "", c.CdnUrl(""))
	assert.True(t, c.NoCdn())
	assert.False(t, c.UseCdn())
	c.options.SiteUrl = "http://superhost:2342/"
	assert.Equal(t, "/", c.CdnUrl("/"))
	c.options.CdnUrl = "http://foo:2342/foo/"
	assert.Equal(t, "http://foo:2342/foo", c.CdnUrl(""))
	assert.Equal(t, "http://foo:2342/foo/", c.CdnUrl("/"))
	assert.False(t, c.NoCdn())
	assert.True(t, c.UseCdn())
	c.options.SiteUrl = c.options.CdnUrl
	assert.Equal(t, "/", c.CdnUrl("/"))
	assert.Equal(t, "", c.CdnUrl(""))
	assert.True(t, c.NoCdn())
	assert.False(t, c.UseCdn())
	c.options.SiteUrl = ""
	assert.False(t, c.NoCdn())
	assert.True(t, c.UseCdn())
}

func TestConfig_CdnDomain(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.options.SiteUrl)
	assert.Equal(t, "", c.CdnDomain())
	c.options.CdnUrl = "http://superhost:2342/"
	assert.Equal(t, "superhost", c.CdnDomain())
	c.options.CdnUrl = "https://foo.bar.com:2342/foo/"
	assert.Equal(t, "foo.bar.com", c.CdnDomain())
	c.options.SiteUrl = c.options.CdnUrl
	assert.Equal(t, "", c.CdnDomain())
	c.options.SiteUrl = ""
	c.options.CdnUrl = "http:/invalid:2342/foo/"
	assert.Equal(t, "", c.CdnDomain())
	c.options.CdnUrl = ""
	assert.Equal(t, "", c.CdnDomain())
}

func TestConfig_CdnVideo(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.CdnVideo())
	c.options.SiteUrl = "http://superhost:2342/"
	assert.False(t, c.CdnVideo())
	c.options.CdnUrl = "http://foo:2342/foo/"
	assert.False(t, c.CdnVideo())
	c.options.CdnVideo = true
	assert.True(t, c.CdnVideo())
	c.options.SiteUrl = c.options.CdnUrl
	assert.False(t, c.CdnVideo())
	c.options.SiteUrl = ""
	assert.True(t, c.CdnVideo())
	c.options.CdnVideo = false
	assert.False(t, c.CdnVideo())
	c.options.CdnUrl = ""
	assert.False(t, c.CdnVideo())
}

func TestConfig_CORSOrigin(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.Options().CORSOrigin = ""
	assert.Equal(t, "", c.CORSOrigin())
	c.Options().CORSOrigin = "*"
	assert.Equal(t, "*", c.CORSOrigin())
	c.Options().CORSOrigin = "https://developer.mozilla.org"
	assert.Equal(t, "https://developer.mozilla.org", c.CORSOrigin())
	c.Options().CORSOrigin = ""
	assert.Equal(t, "", c.CORSOrigin())
}

func TestConfig_CORSHeaders(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.CORSHeaders())
	c.Options().CORSHeaders = header.DefaultAccessControlAllowHeaders
	assert.Equal(t, header.DefaultAccessControlAllowHeaders, c.CORSHeaders())
	c.Options().CORSHeaders = ""
	assert.Equal(t, "", c.CORSHeaders())
}

func TestConfig_CORSMethods(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.CORSMethods())
	c.Options().CORSMethods = header.DefaultAccessControlAllowMethods
	assert.Equal(t, header.DefaultAccessControlAllowMethods, c.CORSMethods())
	c.Options().CORSMethods = ""
	assert.Equal(t, "", c.CORSMethods())
}
