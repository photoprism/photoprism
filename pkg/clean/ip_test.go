package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIP(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0", IP("", "0.0.0.0"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0", IP("0.0.0.0", "0.0.0.0"))
	})
	t.Run("Localhost", func(t *testing.T) {
		assert.Equal(t, "127.0.0.1", IP("127.0.0.1", "0.0.0.0"))
	})
	t.Run("IPv6", func(t *testing.T) {
		assert.Equal(t, "2001:0:130f::9c0:876a:130b", IP("2001:0000:130F:0000:0000:09C0:876A:130B", "0.0.0.0"))
	})
	t.Run("IPv6", func(t *testing.T) {
		assert.Equal(t, "2001:0:130f::9c0:876a:130b", IP("    2001:0000:130F:0000:0000:09C0:876A:130B    ", "0.0.0.0"))
	})
	t.Run("PublicIPv4", func(t *testing.T) {
		assert.Equal(t, "8.8.8.8", IP("8.8.8.8", "0.0.0.0"))
	})
	t.Run("PrivateIPv4", func(t *testing.T) {
		assert.Equal(t, "192.168.1.128", IP("192.168.1.128", "0.0.0.0"))
	})
	t.Run("UUID", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0", IP("123e4567-e89b-12d3-A456-426614174000", "0.0.0.0"))
	})
	t.Run("Hello", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0", IP("Hello", "0.0.0.0"))
	})
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, "default", IP("Hello", "default"))
	})
	t.Run("EmptyDefault", func(t *testing.T) {
		assert.Equal(t, "", IP("Hello", ""))
	})
}
