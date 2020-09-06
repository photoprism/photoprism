package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCredentials(t *testing.T) {
	c := NewCredentials()

	assert.IsType(t, &Credentials{}, c)
}

func TestCredentials_Load(t *testing.T) {
	t.Run("existing filename", func(t *testing.T) {
		c := NewCredentials()

		if err := c.Load("testdata/credentials.yml"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "f60f5b25d59c397989e3cd374f81cdd7710a4fca", c.Key)
		assert.Equal(t, "photoprism", c.Secret)
		assert.Equal(t, "Zm9vYmFy", c.Session)
	})
	t.Run("not existing filename", func(t *testing.T) {
		c := NewCredentials()

		if err := c.Load("testdata/credentials_xxx.yml"); err == nil {
			t.Fatal("file should not exist")
		}

		assert.Equal(t, "", c.Key)
		assert.Equal(t, "", c.Secret)
		assert.Equal(t, "", c.Session)
	})
}
func TestCredentials_Save(t *testing.T) {
	t.Run("existing filename", func(t *testing.T) {
		assert.FileExists(t, "testdata/credentials.yml")

		c := NewCredentials()

		if err := c.Load("testdata/credentials.yml"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "f60f5b25d59c397989e3cd374f81cdd7710a4fca", c.Key)
		assert.Equal(t, "photoprism", c.Secret)
		assert.Equal(t, "Zm9vYmFy", c.Session)

		if err := c.Save("testdata/credentials.yml"); err != nil {
			t.Fatal(err)
		}

		assert.FileExists(t, "testdata/credentials.yml")
	})
	t.Run("not existing filename", func(t *testing.T) {
		c := NewCredentials()
		c.Key = "F60F5B25D59C397989E3CD374F81CDD7710A4FCA"
		c.Secret = "foo"
		c.Session = "bar"

		assert.Equal(t, "F60F5B25D59C397989E3CD374F81CDD7710A4FCA", c.Key)
		assert.Equal(t, "foo", c.Secret)
		assert.Equal(t, "bar", c.Session)

		if err := c.Save("testdata/credentials_new.yml"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "f60f5b25d59c397989e3cd374f81cdd7710a4fca", c.Key)
		assert.Equal(t, "", c.Secret)
		assert.Equal(t, "", c.Session)

		assert.FileExists(t, "testdata/credentials_new.yml")

		if err := os.Remove("testdata/credentials_new.yml"); err != nil {
			t.Fatal(err)
		}
	})
}
