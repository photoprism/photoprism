package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestAbortInvalidCredentials(t *testing.T) {
	t.Run("Response", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		AbortInvalidCredentials(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())

		var body map[string]any
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		// error/code keep their shape for OAuth2 clients and the frontend's code-based handling.
		assert.Equal(t, authn.ErrInvalidCredentials.Error(), body["error"])
		assert.Equal(t, float64(i18n.ErrInvalidCredentials), body["code"])
		assert.Equal(t, i18n.Msg(i18n.ErrInvalidCredentials), body["message"])
		// messageId carries the untranslated source so the Web UI renders it in the user's language.
		assert.Equal(t, i18n.Source(i18n.ErrInvalidCredentials), body["messageId"])
	})
	t.Run("NilContext", func(t *testing.T) {
		assert.NotPanics(t, func() { AbortInvalidCredentials(nil) })
	})
}
