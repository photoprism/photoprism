package safe

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer creates an httptest server and closes it automatically via test cleanup.
func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	return ts
}
