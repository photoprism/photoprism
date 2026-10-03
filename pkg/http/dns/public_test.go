package dns

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPublicName(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for _, name := range []string{
			"photos.example.com",
			"Photos.Example.COM",
			"photos.example.com.",
			"photos-archive.example.org",
			strings.Repeat("a", 40) + ".example.org",
			strings.Repeat("a", 63) + ".example.org",
			"bücher.example.org",
			"xn--bcher-kva.example.org",
			"photos.internal.example.org",
			"BÜCHER.example.org",
			"photos.example.co2",
		} {
			assert.True(t, IsPublicName(name), name)
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, name := range []string{
			"",
			" ",
			".",
			"localhost",
			"photos",
			"photos.local",
			"app.localhost",
			"photos.test",
			"photos.example",
			"photos.invalid",
			"nas.internal",
			"nas.home.arpa",
			"127.1",
			"0x7f.1",
			"photos.example.123",
			"example.com",
			"test",
			"192.0.2.1",
			"2001:db8::1",
			"photos_archive.example.org",
			"-photos.example.org",
			"photos-.example.org",
			"photos..example.org",
			strings.Repeat("a", 64) + ".example.org",
			strings.Repeat("a.", 127) + "org",
		} {
			assert.False(t, IsPublicName(name), name)
		}
	})
}

func TestIsPublicLabel(t *testing.T) {
	assert.True(t, isPublicLabel("photos"))
	assert.True(t, isPublicLabel("a"))
	assert.True(t, isPublicLabel("xn--bcher-kva"))
	assert.True(t, isPublicLabel(strings.Repeat("a", 63)))
	assert.False(t, isPublicLabel(""))
	assert.False(t, isPublicLabel(strings.Repeat("a", 64)))
	assert.False(t, isPublicLabel("-a"))
	assert.False(t, isPublicLabel("a-"))
	assert.False(t, isPublicLabel("Photos"))
	assert.False(t, isPublicLabel("a_b"))
}
