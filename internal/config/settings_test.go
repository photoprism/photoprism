package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSettings(t *testing.T) {
	c := NewSettings(TestConfig())

	assert.IsType(t, new(Settings), c)
}

func TestSettings_Load(t *testing.T) {
	t.Run("existing filename", func(t *testing.T) {
		c := NewSettings(TestConfig())

		if err := c.Load("testdata/settings.yml"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "onyx", c.UI.Theme)
		assert.Equal(t, "de", c.UI.Language)
	})
	t.Run("not existing filename", func(t *testing.T) {
		c := NewSettings(TestConfig())

		err := c.Load("testdata/settings_123.yml")

		assert.Error(t, err)

		assert.Equal(t, "default", c.UI.Theme)
		assert.Equal(t, "en", c.UI.Language)
	})
}
func TestSettings_Save(t *testing.T) {
	t.Run("existing filename", func(t *testing.T) {
		c := NewSettings(TestConfig())

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
		c := NewSettings(TestConfig())
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

func TestSettings_Stacks(t *testing.T) {
	c := NewSettings(TestConfig())
	assert.False(t, c.StackSequences())
	assert.True(t, c.StackUUID())
	assert.True(t, c.StackMeta())
}

func TestConfig_Settings(t *testing.T) {
	c := TestConfig()
	c.options.DisablePlaces = true
	r := c.Settings()
	assert.False(t, r.Features.Places)
}
