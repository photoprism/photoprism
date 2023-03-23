package server

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Authentication cache with an expiration time of 5 minutes.
var basicAuthExpiration = 5 * time.Minute
var basicAuthCache = gc.New(basicAuthExpiration, basicAuthExpiration)
var basicAuthMutex = sync.Mutex{}
var BasicAuthRealm = "Basic realm=\"WebDAV Authorization Required\""

// GetAuthUser returns the authenticated user if found, nil otherwise.
func GetAuthUser(key string) *entity.User {
	user, valid := basicAuthCache.Get(key)

	if valid && user != nil {
		return user.(*entity.User)
	}

	return nil
}

// BasicAuth implements an HTTP request handler that adds basic authentication.
func BasicAuth(conf *config.Config) gin.HandlerFunc {
	var validate = func(c *gin.Context) (name, password, key string, valid bool) {
		name, password, key = GetCredentials(c)

		if name == "" || password == "" {
			return name, password, "", false
		}

		key = fmt.Sprintf("%x", sha1.Sum([]byte(key)))

		if user := GetAuthUser(key); user != nil {
			c.Set(gin.AuthUserKey, user)
			return name, password, key, true
		}

		return name, password, key, false
	}

	return func(c *gin.Context) {
		if c == nil {
			return
		}

		name, password, key, ok := validate(c)

		if ok {
			// Already authenticated.
			return
		} else if key == "" {
			// Incomplete credentials.
			c.Header("WWW-Authenticate", BasicAuthRealm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Get client IP address.
		clientIp := api.ClientIP(c)

		// Check limit for failed auth requests (max. 10 per minute).
		if limiter.Login.Reject(clientIp) {
			limiter.Abort(c)
			return
		}

		basicAuthMutex.Lock()
		defer basicAuthMutex.Unlock()

		// User credentials.
		f := form.Login{
			UserName: name,
			Password: password,
		}

		// Check credentials and authorization.
		if user, _, err := entity.Auth(f, nil, c); err != nil {
			message := err.Error()
			limiter.Login.Reserve(clientIp)
			event.AuditErr([]string{clientIp, "webdav login as %s", message}, clean.LogQuote(name))
			event.LoginError(clientIp, "webdav", name, api.UserAgent(c), message)
		} else if user == nil {
			message := "account not found"
			limiter.Login.Reserve(clientIp)
			event.AuditErr([]string{clientIp, "webdav login as %s", message}, clean.LogQuote(name))
			event.LoginError(clientIp, "webdav", name, api.UserAgent(c), message)
		} else if !user.CanUseWebDAV() {
			// Sync disabled for this account.
			message := "sync disabled"

			event.AuditWarn([]string{clientIp, "webdav login as %s", message}, clean.LogQuote(name))
			event.LoginError(clientIp, "webdav", name, api.UserAgent(c), message)
		} else if err = os.MkdirAll(filepath.Join(conf.OriginalsPath(), user.GetUploadPath()), fs.ModeDir); err != nil {
			message := "failed to create user upload path"

			event.AuditWarn([]string{clientIp, "webdav login as %s", message}, clean.LogQuote(name))
			event.LoginError(clientIp, "webdav", name, api.UserAgent(c), message)
		} else {
			// Successfully authenticated.
			event.AuditInfo([]string{clientIp, "webdav login as %s", "succeeded"}, clean.LogQuote(name))
			event.LoginInfo(clientIp, "webdav", name, api.UserAgent(c))

			// Cache successful authentication.
			basicAuthCache.SetDefault(key, user)
			c.Set(gin.AuthUserKey, user)
			return
		}

		// Abort request.
		c.Header("WWW-Authenticate", BasicAuthRealm)
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

// GetCredentials parses the "Authorization" header into username and password.
func GetCredentials(c *gin.Context) (name, password, raw string) {
	data := c.GetHeader("Authorization")

	if !strings.HasPrefix(data, "Basic ") {
		return "", "", data
	}

	data = strings.TrimPrefix(data, "Basic ")

	auth, err := base64.StdEncoding.DecodeString(data)

	if err != nil {
		return "", "", data
	}

	credentials := strings.SplitN(string(auth), ":", 2)

	if len(credentials) != 2 {
		return "", "", data
	}

	return credentials[0], credentials[1], data
}
