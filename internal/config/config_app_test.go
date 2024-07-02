package config

import (
	"strings"
	"testing"

	"github.com/photoprism/photoprism/internal/config/pwa"
	"github.com/stretchr/testify/assert"
)

func TestConfig_AppName(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "PhotoPrism", c.AppName())
}

func TestConfig_AppMode(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "standalone", c.AppMode())
}

func TestConfig_AppIcon(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "logo", c.AppIcon())
	c.options.AppIcon = "foo"
	assert.Equal(t, "logo", c.AppIcon())
	c.options.AppIcon = "app"
	assert.Equal(t, "app", c.AppIcon())
	c.options.AppIcon = "crisp"
	assert.Equal(t, "crisp", c.AppIcon())
	c.options.AppIcon = "mint"
	assert.Equal(t, "mint", c.AppIcon())
	c.options.AppIcon = "bold"
	assert.Equal(t, "bold", c.AppIcon())
	c.options.AppIcon = "logo"
	assert.Equal(t, "logo", c.AppIcon())
}

func TestConfig_AppColor(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "#000000", c.AppColor())
	c.options.AppColor = "#aBC123"
	assert.Equal(t, "#abc123", c.AppColor())
	c.options.AppColor = ""
	assert.Equal(t, "#000000", c.AppColor())
}

func TestConfig_AppIconsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	if p := c.AppIconsPath(); !strings.HasSuffix(p, "photoprism/assets/static/icons") {
		t.Fatal("path .../photoprism/assets/static/icons expected")
	}

	if p := c.AppIconsPath("app"); !strings.HasSuffix(p, "photoprism/assets/static/icons/app") {
		t.Fatal("path .../photoprism/assets/static/icons/app expected")
	}

	if p := c.AppIconsPath("app", "512.png"); !strings.HasSuffix(p, "photoprism/assets/static/icons/app/512.png") {
		t.Fatal("path .../photoprism/assets/static/icons/app/512.png expected")
	}
}

func TestConfig_AppConfig(t *testing.T) {
	c := NewConfig(CliTestContext())

	result := c.AppConfig()
	assert.NotEmpty(t, result)
	assert.Equal(t, c.AppName(), result.Name)
	assert.Equal(t, c.AppIcon(), result.Icon)
	assert.Equal(t, c.SiteDescription(), result.Description)
	assert.Equal(t, c.BaseUri("/"), result.BaseUri)
	assert.Equal(t, c.StaticUri(), result.StaticUri)
}

func TestConfig_AppManifest(t *testing.T) {
	c := NewConfig(CliTestContext())

	appConf := c.AppConfig()
	assert.NotEmpty(t, appConf)

	t.Run("Cached", func(t *testing.T) {
		result := c.AppManifest()
		assert.NotEmpty(t, result)
		assert.Equal(t, appConf.Name, result.Name)
		assert.Equal(t, appConf.Name, result.ShortName)
		assert.Equal(t, appConf.Description, result.Description)
		assert.Equal(t, appConf.BaseUri, result.Scope)
		assert.Equal(t, appConf.BaseUri+"library/", result.StartUrl)
		assert.Len(t, result.Icons, len(pwa.IconSizes))
		assert.Len(t, result.Categories, len(pwa.Categories))
		assert.Len(t, result.Permissions, len(pwa.Permissions))

		cached := c.AppManifest()
		assert.NotEmpty(t, cached)
		assert.Equal(t, appConf.Name, cached.Name)
		assert.Equal(t, appConf.Name, cached.ShortName)
		assert.Equal(t, appConf.Description, cached.Description)
		assert.Equal(t, appConf.BaseUri, cached.Scope)
		assert.Equal(t, appConf.BaseUri+"library/", cached.StartUrl)
		assert.Len(t, cached.Icons, len(pwa.IconSizes))
		assert.Len(t, cached.Categories, len(pwa.Categories))
		assert.Len(t, cached.Permissions, len(pwa.Permissions))
	})
}
