package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSettings(t *testing.T) {
	c := NewSettings()

	assert.IsType(t, new(Settings), c)
}

func TestSettings_Load(t *testing.T) {
	t.Run("existing filename", func(t *testing.T) {
		c := NewSettings()

		if err := c.Load("testdata/settings.yml"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "onyx", c.UI.Theme)
		assert.Equal(t, "de", c.UI.Language)
	})
	t.Run("not existing filename", func(t *testing.T) {
		c := NewSettings()

		err := c.Load("testdata/settings_123.yml")

		assert.Error(t, err)

		assert.Equal(t, "default", c.UI.Theme)
		assert.Equal(t, "en", c.UI.Language)
	})
}
func TestSettings_Save(t *testing.T) {
	t.Run("existing filename", func(t *testing.T) {
		c := NewSettings()

		assert.Equal(t, "default", c.UI.Theme)
		assert.Equal(t, "en", c.UI.Language)

		c.UI.Theme = "onyx"
		c.UI.Language = "de"

		assert.Equal(t, "onyx", c.UI.Theme)
		assert.Equal(t, "de", c.UI.Language)

		if err := c.Save("testdata/settings.yml"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("not existing filename", func(t *testing.T) {
		c := NewSettings()
		c.UI.Theme = "onyx"
		c.UI.Language = "de"

		assert.Equal(t, "onyx", c.UI.Theme)
		assert.Equal(t, "de", c.UI.Language)

		if err := c.Save("testdata/settings_tmp.yml"); err != nil {
			t.Fatal(err)
		}

		if err := os.Remove("testdata/settings_tmp.yml"); err != nil {
			t.Fatal(err)
		}
	})
}
