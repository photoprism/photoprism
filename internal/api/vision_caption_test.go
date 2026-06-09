package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostVisionCaption(t *testing.T) {
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionCaption(router)

		body := `{"images":["data:image/jpeg;base64,` + strings.Repeat("a", int(MaxVisionRequestBytes)) + `"]}`
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/caption", body)

		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
	})
}
