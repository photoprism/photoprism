package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/http/header"
)

func TestWebDAVResponseWriter_WriteString(t *testing.T) {
	t.Run("ImplicitStatus", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		rw := &webDAVResponseWriter{ResponseWriter: c.Writer, method: header.MethodGet}

		n, err := rw.WriteString("hello")
		require.NoError(t, err)

		assert.Equal(t, 5, n)
		assert.True(t, rw.wroteHeader)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "hello", w.Body.String())
	})
	t.Run("MultiStatusContentType", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		rw := &webDAVResponseWriter{ResponseWriter: c.Writer, method: header.MethodPropfind}
		rw.Header().Set(header.ContentType, header.ContentTypeXml)

		rw.WriteHeader(http.StatusMultiStatus)
		_, err := rw.WriteString("<multistatus/>")
		require.NoError(t, err)

		assert.Equal(t, http.StatusMultiStatus, w.Code)
		assert.Equal(t, "application/xml; charset=utf-8", w.Header().Get(header.ContentType))
		assert.Equal(t, "<multistatus/>", w.Body.String())
	})
}
