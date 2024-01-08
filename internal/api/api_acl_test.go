package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/session"
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
		AddRequestAuthorizationHeader(c.Request, session.PublicAuthToken)

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
		AddRequestAuthorizationHeader(c.Request, session.PublicAuthToken)

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
