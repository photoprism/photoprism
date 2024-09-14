package config

import (
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
	assert.Equal(t, "grayscale", c.DefaultTheme())
	c.options.Sponsor = true
	assert.Equal(t, "grayscale", c.DefaultTheme())
	c.options.Sponsor = false
	c.options.Test = true
	assert.Equal(t, "grayscale", c.DefaultTheme())
	c.options.Sponsor = false
	c.options.Test = false
	assert.Equal(t, "grayscale", c.DefaultTheme())
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

func TestConfig_DefaultTimezone(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "UTC", c.DefaultTimezone().String())

	c.options.DefaultTimezone = "Europe/Berlin"

	assert.Equal(t, "Europe/Berlin", c.DefaultTimezone().String())

	c.options.DefaultTimezone = ""

	assert.Equal(t, "UTC", c.DefaultTimezone().String())
}

func TestConfig_WallpaperUri(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "", c.WallpaperUri())
	assert.Equal(t, "", c.Options().WallpaperUri)
	c.SetWallpaperUri("default")
	assert.Equal(t, "/static/img/wallpaper/default.jpg", c.WallpaperUri())
	c.SetWallpaperUri("https://cdn.photoprism.app/wallpaper/welcome.jpg")
	assert.Equal(t, "https://cdn.photoprism.app/wallpaper/welcome.jpg", c.WallpaperUri())
	c.options.Test = false
	assert.Equal(t, "https://cdn.photoprism.app/wallpaper/welcome.jpg", c.WallpaperUri())
	c.options.Test = true
	assert.Equal(t, "https://cdn.photoprism.app/wallpaper/welcome.jpg", c.WallpaperUri())
	c.options.Sponsor = false
	assert.Equal(t, "https://cdn.photoprism.app/wallpaper/welcome.jpg", c.WallpaperUri())
	c.options.Sponsor = true
	assert.Equal(t, "https://cdn.photoprism.app/wallpaper/welcome.jpg", c.WallpaperUri())
	c.SetWallpaperUri("default")
	assert.Equal(t, "/static/img/wallpaper/default.jpg", c.WallpaperUri())
	c.SetWallpaperUri("default")
	c.options.CdnUrl = "https://bunny.net/cdn/"
	assert.Equal(t, "https://bunny.net/cdn/static/img/wallpaper/default.jpg", c.WallpaperUri())
	assert.Equal(t, "default", c.Options().WallpaperUri)
	c.SetWallpaperUri("default")
	c.options.CdnUrl = ""
	assert.Equal(t, "/static/img/wallpaper/default.jpg", c.WallpaperUri())
	c.SetWallpaperUri("")
	assert.Equal(t, "", c.WallpaperUri())
	assert.Equal(t, "", c.Options().WallpaperUri)
}
