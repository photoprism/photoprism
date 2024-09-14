package api

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/header"
)

func TestAddVideoCacheHeader(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		r := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(r)
		AddVideoCacheHeader(c, true)
		h := r.Header()
		s := h[header.CacheControl][0]
		assert.Equal(t, "public, max-age=21600, immutable", s)
	})
	t.Run("Private", func(t *testing.T) {
		r := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(r)
		AddVideoCacheHeader(c, false)
		h := r.Header()
		s := h[header.CacheControl][0]
		assert.Equal(t, "private, max-age=21600, immutable", s)
	})
}
