package header

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteForwarded(t *testing.T) {
	t.Run("NilHeader", func(t *testing.T) {
		DeleteForwarded(nil, "X-Custom-IP")
	})
	t.Run("Success", func(t *testing.T) {
		h := http.Header{}

		for _, name := range ForwardedHeaders {
			h.Set(name, "203.0.113.7")
		}

		h.Set("X-Custom-IP", "203.0.113.7")
		h.Set("X-Other", "keep")
		h.Set(Accept, "text/html")

		DeleteForwarded(h, " X-Custom-IP ", "")

		assert.Equal(t, http.Header{"X-Other": {"keep"}, Accept: {"text/html"}}, h)
	})
}
