package header

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetLocationWithBase(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/albums", nil)

	SetLocation(c, "/api/v1/albums", "abc123")

	assert.Equal(t, "/api/v1/albums/abc123", c.Writer.Header().Get(Location))
}

func TestSetLocationUsesRequestPath(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/services", nil)

	SetLocation(c, "", "99")

	assert.Equal(t, "/api/v1/services/99", c.Writer.Header().Get(Location))
}

func TestSetLocationTrimsSegments(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/markers/", nil)

	SetLocation(c, "/api/v1/markers/", "/m1/")

	assert.Equal(t, "/api/v1/markers/m1", c.Writer.Header().Get(Location))
}

func TestSetLocationEmpty(t *testing.T) {
	SetLocation(nil)
}
