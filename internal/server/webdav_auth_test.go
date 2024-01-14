package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestWebDAVAuth(t *testing.T) {
	conf := config.TestConfig()
	webdavHandler := WebDAVAuth(conf)

	t.Run("Unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		webdavAuthCache.Flush()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceToken", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		webdavAuthCache.Flush()

		sess := entity.SessionFixtures.Get("alice_token")
		header.SetAuthorization(c.Request, sess.AuthToken())

		webdavHandler(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceTokenWebdav", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		sess := entity.SessionFixtures.Get("alice_token_webdav")
		basicAuth := []byte(fmt.Sprintf("alice:%s", sess.AuthToken()))
		c.Request.Header.Add(header.Auth, fmt.Sprintf("%s %s", header.AuthBasic, base64.StdEncoding.EncodeToString(basicAuth)))

		webdavAuthCache.Flush()
		webdavHandler(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceTokenWebdavWrongUsername", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		sess := entity.SessionFixtures.Get("alice_token_webdav")
		basicAuth := []byte(fmt.Sprintf("bob:%s", sess.AuthToken()))
		c.Request.Header.Add(header.Auth, fmt.Sprintf("%s %s", header.AuthBasic, base64.StdEncoding.EncodeToString(basicAuth)))

		webdavAuthCache.Flush()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceTokenWebdavWithoutUsername", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		sess := entity.SessionFixtures.Get("alice_token_webdav")
		basicAuth := []byte(fmt.Sprintf(":%s", sess.AuthToken()))
		c.Request.Header.Add(header.Auth, fmt.Sprintf("%s %s", header.AuthBasic, base64.StdEncoding.EncodeToString(basicAuth)))

		webdavAuthCache.Flush()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceTokenScope", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		sess := entity.SessionFixtures.Get("alice_token_scope")
		header.SetAuthorization(c.Request, sess.AuthToken())

		webdavAuthCache.Flush()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("InvalidAuthToken", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		header.SetAuthorization(c.Request, rnd.AuthToken())

		webdavAuthCache.Flush()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("InvalidAuthSecret", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		header.SetAuthorization(c.Request, rnd.AuthSecret())

		webdavAuthCache.Flush()
		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
}

func TestWebDAVAuthSession(t *testing.T) {
	t.Run("AliceTokenWebdav", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		s := entity.SessionFixtures.Get("alice_token_webdav")

		// Get session with authorized user and webdav scope.
		webdavAuthCache.Flush()
		sess, user, sid, cached := WebDAVAuthSession(c, s.AuthToken())

		// Check result.
		assert.NotNil(t, sess)
		assert.NotNil(t, user)
		assert.True(t, sess.HasUser())
		assert.Equal(t, user.UserUID, sess.UserUID)
		assert.Equal(t, entity.UserFixtures.Get("alice").UserUID, sess.UserUID)
		assert.True(t, sess.HasScope(acl.ResourceWebDAV.String()))
		assert.False(t, cached)

		assert.Equal(t, s.ID, sid)
		assert.Equal(t, entity.UserFixtures.Get("alice").UserUID, user.UserUID)
		assert.True(t, user.CanUseWebDAV())

		// WebDAVAuthSession should not set a status code or any headers.
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))

		// Cache authentication.
		webdavAuthCache.SetDefault(sid, user)

		// Get cached user.
		sess, user, sid, cached = WebDAVAuthSession(c, s.AuthToken())

		// Check result.
		assert.Nil(t, sess)
		assert.NotNil(t, user)
		assert.True(t, cached)

		assert.Equal(t, s.ID, sid)
		assert.Equal(t, entity.UserFixtures.Get("alice").UserUID, user.UserUID)
		assert.True(t, user.CanUseWebDAV())

		// WebDAVAuthSession should not set a status code or any headers.
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceTokenScope", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		s := entity.SessionFixtures.Get("alice_token_scope")

		// Get session without sufficient authorization scope.
		sess, user, sid, cached := WebDAVAuthSession(c, s.AuthToken())

		// Check result.
		assert.NotNil(t, sess)
		assert.NotNil(t, user)
		assert.Equal(t, s.ID, sid)
		assert.False(t, cached)
		assert.True(t, sess.HasUser())
		assert.Equal(t, user.UserUID, sess.UserUID)
		assert.Equal(t, entity.UserFixtures.Get("alice").UserUID, user.UserUID)
		assert.Equal(t, entity.UserFixtures.Get("alice").UserUID, sess.UserUID)
		assert.True(t, user.CanUseWebDAV())
		assert.False(t, sess.HasScope(acl.ResourceWebDAV.String()))

		// WebDAVAuthSession should not set a status code or any headers.
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("InvalidAuthSecret", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		authToken := rnd.AuthSecret()
		authId := rnd.SessionID(authToken)

		// Get session with invalid auth secret.
		sess, user, sid, cached := WebDAVAuthSession(c, authToken)

		// Check result.
		assert.Nil(t, sess)
		assert.Nil(t, user)
		assert.Equal(t, authId, sid)
		assert.False(t, cached)

		// WebDAVAuthSession should not set a status code or any headers.
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
	})
}
