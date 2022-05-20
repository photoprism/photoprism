package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_DefaultTheme(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "default", c.DefaultTheme())
	c.options.Demo = false
	c.options.Sponsor = false
	c.options.Test = false
	c.options.DefaultTheme = "grayscale"
	assert.Equal(t, "default", c.DefaultTheme())
	c.options.Sponsor = true
	assert.Equal(t, "grayscale", c.DefaultTheme())
	c.options.Sponsor = false
	c.options.Test = true
	assert.Equal(t, "grayscale", c.DefaultTheme())
	c.options.Sponsor = false
	c.options.Test = false
	assert.Equal(t, "default", c.DefaultTheme())
	c.options.Sponsor = true
	c.options.DefaultTheme = ""
	assert.Equal(t, "default", c.DefaultTheme())
	c.options.Sponsor = false
	assert.Equal(t, "default", c.DefaultTheme())
}

func TestConfig_DefaultLocale(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "en", c.DefaultLocale())
	c.options.DefaultLocale = "de"
	assert.Equal(t, "de", c.DefaultLocale())
	c.options.DefaultLocale = ""
	assert.Equal(t, "en", c.DefaultLocale())
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

func TestConfig_AppIconsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	if p := c.AppIconsPath(); !strings.HasSuffix(p, "photoprism/assets/static/icons") {
		t.Fatal("path .../photoprism/assets/static/icons expected")
	}

	if p := c.AppIconsPath("app"); !strings.HasSuffix(p, "photoprism/assets/static/icons/app") {
		t.Fatal("path .../pphotoprism/assets/static/icons/app expected")
	}

	if p := c.AppIconsPath("app", "512.png"); !strings.HasSuffix(p, "photoprism/assets/static/icons/app/512.png") {
		t.Fatal("path .../photoprism/assets/static/icons/app/512.png expected")
	}
}

func TestConfig_AppName(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "PhotoPrism", c.AppName())
}

func TestConfig_AppMode(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "standalone", c.AppMode())
}

func TestConfig_WallpaperUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.WallpaperUri())
	c.options.WallpaperUri = "kashmir"
	assert.Equal(t, "/static/img/wallpaper/kashmir.jpg", c.WallpaperUri())
	c.options.WallpaperUri = "https://cdn.photoprism.app/wallpaper/welcome.jpg"
	assert.Equal(t, "https://cdn.photoprism.app/wallpaper/welcome.jpg", c.WallpaperUri())
	c.options.Test = false
	assert.Equal(t, "", c.WallpaperUri())
	c.options.Test = true
	assert.Equal(t, "https://cdn.photoprism.app/wallpaper/welcome.jpg", c.WallpaperUri())
	c.options.WallpaperUri = ""
	assert.Equal(t, "", c.WallpaperUri())
}
