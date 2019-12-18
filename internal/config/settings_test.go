package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSettings(t *testing.T) {
	c := NewSettings()
	assert.IsType(t, new(Settings), c)
}

func TestSettings_SetValuesFromFile(t *testing.T) {
	t.Run("existing filename", func(t *testing.T) {
		c := NewSettings()

		err := c.SetValuesFromFile("testdata/config.yml")

		assert.Nil(t, err)

		assert.Equal(t, "lavendel", c.Theme)
		assert.Equal(t, "english", c.Language)
	})
	t.Run("not existing filename", func(t *testing.T) {
		c := NewSettings()

		err := c.SetValuesFromFile("testdata/config123.yml")

		assert.Error(t, err)

		assert.Equal(t, "", c.Theme)
		assert.Equal(t, "", c.Language)
	})
}
func TestSettings_WriteValuesToFile(t *testing.T) {
	t.Run("existing filename", func(t *testing.T) {
		c := NewSettings()
		c.Theme = "lavendel"
		c.Language = "german"

		assert.Equal(t, "lavendel", c.Theme)
		assert.Equal(t, "german", c.Language)

		err := c.WriteValuesToFile("testdata/configEmpty.yml")

		assert.Nil(t, err)
	})
	t.Run("not existing filename", func(t *testing.T) {
		c := NewSettings()
		c.Theme = "lavendel"
		c.Language = "german"

		assert.Equal(t, "lavendel", c.Theme)
		assert.Equal(t, "german", c.Language)

		err := c.WriteValuesToFile("testdata/configEmpty123.yml")

		assert.Error(t, err)
	})
}
