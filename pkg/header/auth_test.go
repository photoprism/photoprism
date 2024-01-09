package header

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthToken(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// No headers have been set, so no token should be returned.
		token := AuthToken(c)
		assert.Equal(t, "", token)
	})
	t.Run("BearerToken", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// Add authorization header.
		SetAuthorization(c.Request, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")

		// Check result.
		authToken := AuthToken(c)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", authToken)
		bearerToken := BearerToken(c)
		assert.Equal(t, authToken, bearerToken)
	})
	t.Run("Header", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// Add authorization header.
		c.Request.Header.Add(XAuthToken, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")

		// Check result.
		authToken := AuthToken(c)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", authToken)
		bearerToken := BearerToken(c)
		assert.Equal(t, "", bearerToken)
	})
}

func TestBearerToken(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// No headers have been set, so no token should be returned.
		token := BearerToken(c)
		assert.Equal(t, "", token)
	})
	t.Run("Found", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// Add authorization header.
		SetAuthorization(c.Request, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")

		// Check result.
		token := BearerToken(c)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", token)
	})
}

func TestAuthorization(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// No headers have been set, so no token should be returned.
		authType, authToken := Authorization(c)
		assert.Equal(t, "", authType)
		assert.Equal(t, "", authToken)
	})
	t.Run("BearerToken", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// Add authorization header.
		c.Request.Header.Add(Auth, "Bearer 69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")

		// Check result.
		authType, authToken := Authorization(c)
		assert.Equal(t, AuthBearer, authType)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", authToken)
	})
}

func TestBasicAuth(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// No headers have been set, so no token should be returned.
		user, pass, key := BasicAuth(c)
		assert.Equal(t, "", user)
		assert.Equal(t, "", pass)
		assert.Equal(t, "", key)
	})
	t.Run("Found", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// Add authorization header.
		c.Request.Header.Add(Auth, AuthBasic+" QWxhZGRpbjpvcGVuIHNlc2FtZQ==")

		// Check result.
		user, pass, key := BasicAuth(c)
		assert.Equal(t, "Aladdin", user)
		assert.Equal(t, "open sesame", pass)
		assert.Equal(t, "0cdb723383eb144043424a4a254461658d887396", key)
	})
}
