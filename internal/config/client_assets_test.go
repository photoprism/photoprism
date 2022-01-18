package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientAssets_Load(t *testing.T) {
	c := NewConfig(CliTestContext())

	t.Run("Success", func(t *testing.T) {
		a := NewClientAssets(c.StaticUri())

		err := a.Load("testdata/static/build/assets.json")

		assert.NoError(t, err)

		assert.Equal(t, "/static", a.BaseUri)
		assert.Equal(t, "app.2259c0edcc020e7af593.css", a.AppCss)
		assert.Equal(t, "/static/build/app.2259c0edcc020e7af593.css", a.AppCssUri())
		assert.Equal(t, "app.9bd7132eaee8e4c7c7e3.js", a.AppJs)
		assert.Equal(t, "/static/build/app.9bd7132eaee8e4c7c7e3.js", a.AppJsUri())
		assert.Equal(t, "share.2259c0edcc020e7af593.css", a.ShareCss)
		assert.Equal(t, "/static/build/share.2259c0edcc020e7af593.css", a.ShareCssUri())
		assert.Equal(t, "share.7aaf321a984ae545e4e5.js", a.ShareJs)
		assert.Equal(t, "/static/build/share.7aaf321a984ae545e4e5.js", a.ShareJsUri())
	})

	t.Run("Error", func(t *testing.T) {
		a := NewClientAssets(c.StaticUri())

		err := a.Load("testdata/foo/assets.json")

		assert.Error(t, err)

		assert.Equal(t, "/static", a.BaseUri)
		assert.Equal(t, "", a.AppCss)
		assert.Equal(t, "", a.AppCssUri())
		assert.Equal(t, "", a.AppJs)
		assert.Equal(t, "", a.AppJsUri())
		assert.Equal(t, "", a.ShareCss)
		assert.Equal(t, "", a.ShareCssUri())
		assert.Equal(t, "", a.ShareJs)
		assert.Equal(t, "", a.ShareJsUri())
	})
}

func TestConfig_ClientAssets(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.AssetsPath = "testdata"
	c.options.CdnUrl = "https://mycdn.com/foo/"

	a := c.ClientAssets()

	assert.Equal(t, "https://mycdn.com/foo/static", a.BaseUri)
	assert.Equal(t, "app.2259c0edcc020e7af593.css", a.AppCss)
	assert.Equal(t, "https://mycdn.com/foo/static/build/app.2259c0edcc020e7af593.css", a.AppCssUri())
	assert.Equal(t, "app.9bd7132eaee8e4c7c7e3.js", a.AppJs)
	assert.Equal(t, "https://mycdn.com/foo/static/build/app.9bd7132eaee8e4c7c7e3.js", a.AppJsUri())
	assert.Equal(t, "share.2259c0edcc020e7af593.css", a.ShareCss)
	assert.Equal(t, "https://mycdn.com/foo/static/build/share.2259c0edcc020e7af593.css", a.ShareCssUri())
	assert.Equal(t, "share.7aaf321a984ae545e4e5.js", a.ShareJs)
	assert.Equal(t, "https://mycdn.com/foo/static/build/share.7aaf321a984ae545e4e5.js", a.ShareJsUri())

	c.options.AssetsPath = "testdata/invalid"
	c.options.CdnUrl = ""
	c.options.SiteUrl = "http://myhost/foo"

	a = c.ClientAssets()

	assert.Equal(t, "/foo/static", a.BaseUri)
	assert.Equal(t, "", a.AppCss)
	assert.Equal(t, "", a.AppCssUri())
	assert.Equal(t, "", a.AppJs)
	assert.Equal(t, "", a.AppJsUri())
	assert.Equal(t, "", a.ShareCss)
	assert.Equal(t, "", a.ShareCssUri())
	assert.Equal(t, "", a.ShareJs)
	assert.Equal(t, "", a.ShareJsUri())
}

func TestClientManifestUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "/manifest.json?1638035864", c.ClientManifestUri())

	c.options.SiteUrl = ""

	assert.Equal(t, "/manifest.json?1638035864", c.ClientManifestUri())

	c.options.SiteUrl = "http://myhost/foo"

	assert.True(t, strings.HasPrefix(c.ClientManifestUri(), "/foo/manifest.json"))
}
