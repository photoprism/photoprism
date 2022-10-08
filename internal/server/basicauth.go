package server

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	gc "github.com/patrickmn/go-cache"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
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
func BasicAuth() gin.HandlerFunc {
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
		name, password, key, valid := validate(c)

		if valid {
			// Already authenticated.
			return
		} else if key == "" {
			// Incomplete credentials.
			c.Header("WWW-Authenticate", BasicAuthRealm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		basicAuthMutex.Lock()
		defer basicAuthMutex.Unlock()

		// Check authentication and authorization.
		if user := entity.FindUserByName(name); user == nil {
			// Username not found.
			message := "account not found"

			event.AuditWarn([]string{api.ClientIP(c), "webdav login as %s", message}, clean.LogQuote(name))
			event.LoginError(api.ClientIP(c), name, api.UserAgent(c), message)
		} else if !user.SyncAllowed() {
			// Sync disabled for this account.
			message := "sync disabled"

			event.AuditWarn([]string{api.ClientIP(c), "webdav login as %s", message}, clean.LogQuote(name))
			event.LoginError(api.ClientIP(c), name, api.UserAgent(c), message)
		} else if valid = !user.InvalidPassword(password); !valid {
			// Wrong password.
			message := "incorrect password"

			event.AuditErr([]string{api.ClientIP(c), "webdav login as %s", message}, clean.LogQuote(name))
			event.LoginError(api.ClientIP(c), name, api.UserAgent(c), message)
		} else {
			// Successfully authenticated.
			event.AuditInfo([]string{api.ClientIP(c), "webdav login as %s", "succeeded"}, clean.LogQuote(name))
			event.LoginSuccess(api.ClientIP(c), name, api.UserAgent(c))

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
