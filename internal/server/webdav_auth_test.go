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
	t.Run("InvalidAppPassword", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		header.SetAuthorization(c.Request, rnd.AppPassword())

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
		assert.True(t, sess.ValidateScope(acl.ResourceWebDAV, acl.Permissions{acl.ActionView}))
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
		assert.False(t, sess.ValidateScope(acl.ResourceWebDAV, acl.Permissions{acl.ActionView}))

		// WebDAVAuthSession should not set a status code or any headers.
		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("InvalidAppPassword", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		appPassword := rnd.AppPassword()
		authId := rnd.SessionID(appPassword)

		// Get session with invalid app password.
		sess, user, sid, cached := WebDAVAuthSession(c, appPassword)

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
