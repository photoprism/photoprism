package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errReader fails every read, as a connection that drops mid-body would.
type errReader struct{}

// Read returns an error.
func (errReader) Read([]byte) (int, error) { return 0, errors.New("connection reset") }

func TestReadRequestBody(t *testing.T) {
	newContext := func(body string) (*gin.Context, *httptest.ResponseRecorder) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))

		return c, w
	}

	t.Run("Success", func(t *testing.T) {
		c, _ := newContext(`{"Name":"Jane"}`)

		body, ok := readRequestBody(c, 64)
		require.True(t, ok)
		assert.Equal(t, `{"Name":"Jane"}`, string(body))
		assert.False(t, c.IsAborted())
	})
	t.Run("ExactLimit", func(t *testing.T) {
		c, _ := newContext(strings.Repeat("x", 64))

		body, ok := readRequestBody(c, 64)
		require.True(t, ok)
		assert.Len(t, body, 64)
	})
	t.Run("InvalidLimit", func(t *testing.T) {
		// Without a positive bound the body is not read at all, rather than read unbounded.
		c, w := newContext(`{"Name":"Jane"}`)

		body, ok := readRequestBody(c, 0)
		assert.False(t, ok)
		assert.Nil(t, body)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
	t.Run("TooLarge", func(t *testing.T) {
		c, w := newContext(strings.Repeat("x", 65))

		body, ok := readRequestBody(c, 64)
		assert.False(t, ok)
		assert.Nil(t, body)
		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})
	t.Run("ReadError", func(t *testing.T) {
		c, w := newContext("")
		c.Request.Body = errReaderCloser{}

		_, ok := readRequestBody(c, 64)
		assert.False(t, ok)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("NoBody", func(t *testing.T) {
		c, _ := newContext("")
		c.Request.Body = nil

		body, ok := readRequestBody(c, 64)
		assert.True(t, ok)
		assert.Empty(t, body)
	})
}

// errReaderCloser is an errReader that can be closed.
type errReaderCloser struct{ errReader }

// Close does nothing.
func (errReaderCloser) Close() error { return nil }
