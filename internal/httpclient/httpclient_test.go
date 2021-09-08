package httpclient

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClient(t *testing.T) {
	t.Run("default client", func(t *testing.T) {
		client := Client(false)

		assert.IsType(t, http.DefaultClient, client)
		assert.IsType(t, nil, client.Transport)
	})
	t.Run("logging proxy client", func(t *testing.T) {
		client := Client(true)

		assert.IsType(t, LoggingRoundTripper{}, client.Transport)
	})
	t.Run("RoundTripper working", func(t *testing.T) {
		req, err := http.NewRequest("GET", "https://photoprism.app", nil)
		assert.Nil(t, err)
		rt := LoggingRoundTripper{http.DefaultTransport}
		_, err = rt.RoundTrip(req)
		assert.Nil(t, err)
	})
}
