package scheme

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScheme(t *testing.T) {
	t.Run("HTTP", func(t *testing.T) {
		assert.Equal(t, "http", Http)
		assert.Equal(t, "http+unix", HttpUnix)
		assert.Equal(t, "https", Https)
		assert.Equal(t, "wss", Websocket)
	})
	t.Run("Unix", func(t *testing.T) {
		assert.Equal(t, "unix", Unix)
		assert.Equal(t, "unixgram", Unixgram)
		assert.Equal(t, "unixpacket", Unixpacket)
	})
}
