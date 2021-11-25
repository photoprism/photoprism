package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_AppName(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "config.test", c.AppName())
}

func TestConfig_AppMode(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "standalone", c.AppMode())
}

func TestConfig_AppIcon(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, "default", c.AppIcon())
	c.options.AppIcon = "foo"
	assert.Equal(t, "default", c.AppIcon())
	c.options.AppIcon = "lens"
	assert.Equal(t, "lens", c.AppIcon())
	c.options.AppIcon = "default"
	assert.Equal(t, "default", c.AppIcon())
}

func TestConfig_AppIconsPath(t *testing.T) {
	c := NewConfig(CliTestContext())

	if p := c.AppIconsPath(); !strings.HasSuffix(p, "photoprism/assets/static/icons") {
		t.Fatal("path .../photoprism/assets/static/icons expected")
	}

	if p := c.AppIconsPath("lens"); !strings.HasSuffix(p, "photoprism/assets/static/icons/lens") {
		t.Fatal("path .../pphotoprism/assets/static/icons/lens expected")
	}

	if p := c.AppIconsPath("lens", "512.png"); !strings.HasSuffix(p, "photoprism/assets/static/icons/lens/512.png") {
		t.Fatal("path .../photoprism/assets/static/icons/lens/512.png expected")
	}
}
