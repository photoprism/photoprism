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

		err := c.Load("testdata/config.yml")

		assert.Nil(t, err)

		assert.Equal(t, "lavendel", c.Theme)
		assert.Equal(t, "english", c.Language)
	})
	t.Run("not existing filename", func(t *testing.T) {
		c := NewSettings()

		err := c.Load("testdata/config123.yml")

		assert.Error(t, err)

		assert.Equal(t, "default", c.Theme)
		assert.Equal(t, "en", c.Language)
	})
}
func TestSettings_Save(t *testing.T) {
	t.Run("existing filename", func(t *testing.T) {
		c := NewSettings()
		c.Theme = "lavendel"
		c.Language = "german"

		assert.Equal(t, "lavendel", c.Theme)
		assert.Equal(t, "german", c.Language)

		err := c.Save("testdata/configEmpty.yml")

		assert.Nil(t, err)
	})
	t.Run("not existing filename", func(t *testing.T) {
		c := NewSettings()
		c.Theme = "lavendel"
		c.Language = "german"

		assert.Equal(t, "lavendel", c.Theme)
		assert.Equal(t, "german", c.Language)

		err := c.Save("testdata/configEmpty123.yml")

		if err != nil {
			t.Fatal(err)
		}

		if err := os.Remove("testdata/configEmpty123.yml"); err != nil {
			t.Fatal(err)
		}
	})
}
