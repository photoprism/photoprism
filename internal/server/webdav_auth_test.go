package server

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

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
		basicAuth := []byte(fmt.Sprintf("access-token:%s", sess.AuthToken()))
		c.Request.Header.Add(header.Auth, fmt.Sprintf("%s %s", header.AuthBasic, base64.StdEncoding.EncodeToString(basicAuth)))

		webdavHandler(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
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

		webdavHandler(c)

		assert.Equal(t, http.StatusOK, c.Writer.Status())
		assert.Equal(t, "", c.Writer.Header().Get("WWW-Authenticate"))
	})
	t.Run("AliceTokenScope", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			Header: make(http.Header),
		}

		sess := entity.SessionFixtures.Get("alice_token_scope")
		header.SetAuthorization(c.Request, sess.AuthToken())

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

		webdavHandler(c)

		assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
		assert.Equal(t, BasicAuthRealm, c.Writer.Header().Get("WWW-Authenticate"))
	})
}
