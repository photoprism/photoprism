package remote

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscover(t *testing.T) {
	t.Run("webdav", func(t *testing.T) {
		r, err := Discover("http://admin:photoprism@dummy-webdav/", "", "")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Dummy-Webdav", r.AccName)
		assert.Equal(t, "webdav", r.AccType)
		assert.Equal(t, "http://dummy-webdav/", r.AccURL)
		assert.Equal(t, "admin", r.AccUser)
		assert.Equal(t, "photoprism", r.AccPass)
	})

	t.Run("webdav password", func(t *testing.T) {
		r, err := Discover("http://admin@dummy-webdav/", "", "photoprism")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Dummy-Webdav", r.AccName)
		assert.Equal(t, "webdav", r.AccType)
		assert.Equal(t, "http://dummy-webdav/", r.AccURL)
		assert.Equal(t, "admin", r.AccUser)
		assert.Equal(t, "photoprism", r.AccPass)
	})

	t.Run("facebook", func(t *testing.T) {
		r, err := Discover("https://www.facebook.com/terms", "test", "")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Facebook", r.AccName)
		assert.Equal(t, "facebook", r.AccType)
		assert.Equal(t, "https://www.facebook.com/terms", r.AccURL)
		assert.Equal(t, "test", r.AccUser)
		assert.Equal(t, "", r.AccPass)
	})
	t.Run("empty raw url", func(t *testing.T) {
		r, err := Discover("", "", "")

		assert.Equal(t, err.Error(), "service URL is empty")
		assert.Equal(t, "", r.AccName)
		assert.Equal(t, "", r.AccType)
		assert.Equal(t, "", r.AccURL)
		assert.Equal(t, "", r.AccUser)
		assert.Equal(t, "", r.AccPass)
	})

	t.Run("invalid raw url", func(t *testing.T) {
		r, err := Discover("xxx", "", "")

		assert.Equal(t, err.Error(), "could not connect")
		assert.Equal(t, "", r.AccName)
		assert.Equal(t, "", r.AccType)
		assert.Equal(t, "", r.AccURL)
		assert.Equal(t, "", r.AccUser)
		assert.Equal(t, "", r.AccPass)
	})
}
