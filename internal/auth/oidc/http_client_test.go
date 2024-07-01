package oidc

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpClient(t *testing.T) {
	t.Run("DefaultClient", func(t *testing.T) {
		client := HttpClient(false)

		assert.IsType(t, http.DefaultClient, client)
		assert.IsType(t, nil, client.Transport)
	})
	t.Run("LoggingProxyClient", func(t *testing.T) {
		client := HttpClient(true)

		assert.IsType(t, LoggingRoundTripper{}, client.Transport)
	})
	t.Run("GetRequest", func(t *testing.T) {
		req, err := http.NewRequest("GET", "https://www.photoprism.app/", nil)
		assert.Nil(t, err)
		rt := LoggingRoundTripper{http.DefaultTransport}
		_, err = rt.RoundTrip(req)
		assert.Nil(t, err)
	})
}
