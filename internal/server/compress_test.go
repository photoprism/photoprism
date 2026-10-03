package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/http/header"
)

// newTestCompressWriter returns a gzip compressWriter around a Gin response recorder.
func newTestCompressWriter(status int) (*compressWriter, *gzip.Writer, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Status(status)

	gz := gzip.NewWriter(c.Writer)

	return &compressWriter{ResponseWriter: c.Writer, encoder: gz, encoding: EncodingGzip}, gz, w
}

func TestCompressWriter_WriteString(t *testing.T) {
	t.Run("Encoded", func(t *testing.T) {
		cw, gz, w := newTestCompressWriter(http.StatusOK)

		n, err := cw.WriteString("hello world")
		require.NoError(t, err)
		assert.Equal(t, len("hello world"), n)
		assert.True(t, cw.encoded)
		require.NoError(t, gz.Close())

		zr, err := gzip.NewReader(w.Body)
		require.NoError(t, err)
		decoded, err := io.ReadAll(zr)
		require.NoError(t, err)
		assert.Equal(t, "hello world", string(decoded))
	})
	t.Run("BypassedForErrors", func(t *testing.T) {
		cw, _, w := newTestCompressWriter(http.StatusNotFound)
		cw.Header().Set(header.ContentEncoding, EncodingGzip)

		_, err := cw.WriteString("not found")
		require.NoError(t, err)

		assert.True(t, cw.bypass)
		assert.False(t, cw.encoded)
		assert.Empty(t, w.Header().Get(header.ContentEncoding))
		assert.Equal(t, "not found", w.Body.String())
	})
}

func TestCompressWriter_Flush(t *testing.T) {
	t.Run("FlushesEncoder", func(t *testing.T) {
		cw, _, w := newTestCompressWriter(http.StatusOK)

		_, err := cw.WriteString(strings.Repeat("a", 64))
		require.NoError(t, err)

		before := w.Body.Len()
		cw.Flush()

		assert.Greater(t, w.Body.Len(), before)
		assert.True(t, w.Flushed)
	})
	t.Run("Bypassed", func(t *testing.T) {
		cw, _, w := newTestCompressWriter(http.StatusInternalServerError)

		_, err := cw.WriteString("error")
		require.NoError(t, err)
		cw.Flush()

		assert.Equal(t, "error", w.Body.String())
		assert.True(t, w.Flushed)
	})
}
