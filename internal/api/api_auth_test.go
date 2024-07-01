package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/session"
	"github.com/photoprism/photoprism/pkg/header"
)

func TestAuth(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// Add authorization header.
		header.SetAuthorization(c.Request, session.PublicAuthToken)

		// Check auth token.
		authToken := AuthToken(c)
		assert.Equal(t, session.PublicAuthToken, authToken)

		// Check successful authorization in public mode.
		s := Auth(c, acl.ResourceFiles, acl.ActionUpdate)
		assert.NotNil(t, s)
		assert.Equal(t, "admin", s.Username())
		assert.Equal(t, session.PublicID, s.ID)
		assert.Equal(t, http.StatusOK, s.HttpStatus())
		assert.False(t, s.Abort(c))

		// Check failed authorization in public mode.
		s = Auth(c, acl.ResourceUsers, acl.ActionUpload)
		assert.NotNil(t, s)
		assert.Equal(t, "", s.Username())
		assert.Equal(t, "", s.ID)
		assert.Equal(t, http.StatusForbidden, s.HttpStatus())
		assert.True(t, s.Abort(c))
	})
}

func TestAuthAny(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		// Add authorization header.
		header.SetAuthorization(c.Request, session.PublicAuthToken)

		// Check auth token.
		authToken := AuthToken(c)
		assert.Equal(t, session.PublicAuthToken, authToken)

		// Check successful authorization in public mode.
		s := AuthAny(c, acl.ResourceFiles, acl.Permissions{acl.ActionUpdate})
		assert.NotNil(t, s)
		assert.Equal(t, "admin", s.Username())
		assert.Equal(t, session.PublicID, s.ID)
		assert.Equal(t, http.StatusOK, s.HttpStatus())
		assert.False(t, s.Abort(c))

		// Check failed authorization in public mode.
		s = AuthAny(c, acl.ResourceUsers, acl.Permissions{acl.ActionUpload})
		assert.NotNil(t, s)
		assert.Equal(t, "", s.Username())
		assert.Equal(t, "", s.ID)
		assert.Equal(t, http.StatusForbidden, s.HttpStatus())
		assert.True(t, s.Abort(c))

		// Check successful authorization with multiple actions in public mode.
		s = AuthAny(c, acl.ResourceUsers, acl.Permissions{acl.ActionUpload, acl.ActionView})
		assert.NotNil(t, s)
		assert.Equal(t, "admin", s.Username())
		assert.Equal(t, session.PublicID, s.ID)
		assert.Equal(t, http.StatusOK, s.HttpStatus())
		assert.False(t, s.Abort(c))
	})
}

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
		header.SetAuthorization(c.Request, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")

		// Check result.
		authToken := AuthToken(c)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", authToken)
		bearerToken := header.BearerToken(c)
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
		c.Request.Header.Add(header.XAuthToken, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")

		// Check result.
		authToken := AuthToken(c)
		assert.Equal(t, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0", authToken)
		bearerToken := header.BearerToken(c)
		assert.Equal(t, "", bearerToken)
	})
}
