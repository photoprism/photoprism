package header

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCacheControlMaxAge(t *testing.T) {
	t.Run("Private", func(t *testing.T) {
		assert.Equal(t, CacheControlPrivateDefault, CacheControlMaxAge(0, false))
		assert.Equal(t, "no-cache", CacheControlMaxAge(-1, false))
		assert.Equal(t, "private, max-age=1", CacheControlMaxAge(1, false))
		assert.Equal(t, "private, max-age=31536000", CacheControlMaxAge(DurationYear, false))
		assert.Equal(t, "private, max-age=31536000", CacheControlMaxAge(1231536000, false))
	})
	t.Run("Public", func(t *testing.T) {
		assert.Equal(t, CacheControlPublicDefault, CacheControlMaxAge(0, true))
		assert.Equal(t, "no-cache", CacheControlMaxAge(-1, true))
		assert.Equal(t, "public, max-age=1", CacheControlMaxAge(1, true))
		assert.Equal(t, "public, max-age=31536000", CacheControlMaxAge(DurationYear, true))
		assert.Equal(t, "public, max-age=31536000", CacheControlMaxAge(1231536000, true))
	})
}

func BenchmarkTestCacheControlMaxAge(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = CacheControlMaxAge(DurationYear, false)
	}
}

func BenchmarkTestCacheControlMaxAgeImmutable(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = CacheControlMaxAge(DurationYear, false) + ", " + CacheControlImmutable
	}
}

func TestSetCacheControl(t *testing.T) {
	t.Run("Private", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		SetCacheControl(c, DurationYear, false)
		assert.Equal(t, "private, max-age=31536000", c.Writer.Header().Get(CacheControl))
	})
	t.Run("Public", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		SetCacheControl(c, DurationYear, true)
		assert.Equal(t, "public, max-age=31536000", c.Writer.Header().Get(CacheControl))
	})
	t.Run("NoCache", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		SetCacheControl(c, -1, true)
		assert.Equal(t, CacheControlNoCache, c.Writer.Header().Get(CacheControl))
	})
}

func TestSetCacheControlImmutable(t *testing.T) {
	t.Run("Private", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		SetCacheControlImmutable(c, DurationYear, false)
		assert.Equal(t, "private, max-age=31536000, immutable", c.Writer.Header().Get(CacheControl))
	})
	t.Run("Public", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		SetCacheControlImmutable(c, DurationYear, true)
		assert.Equal(t, "public, max-age=31536000, immutable", c.Writer.Header().Get(CacheControl))
	})
	t.Run("PublicDefault", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		SetCacheControlImmutable(c, 0, true)
		assert.Equal(t, CacheControlPublicDefault+", immutable", c.Writer.Header().Get(CacheControl))
	})
	t.Run("NoCache", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		SetCacheControlImmutable(c, -1, true)
		assert.Equal(t, CacheControlNoCache, c.Writer.Header().Get(CacheControl))
	})
}
