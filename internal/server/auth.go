package server

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
)

var basicAuth = struct {
	user  map[string]entity.User
	mutex sync.RWMutex
}{user: make(map[string]entity.User)}

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

func BasicAuth() gin.HandlerFunc {
	realm := "Authorization Required"
	realm = "Basic realm=" + strconv.Quote(realm)

	return func(c *gin.Context) {
		invalid := true

		name, password, raw := GetCredentials(c)

		basicAuth.mutex.Lock()
		defer basicAuth.mutex.Unlock()

		if user, ok := basicAuth.user[raw]; ok {
			c.Set(gin.AuthUserKey, user.UserUID)
			return
		}

		// Check credentials and authorization.
		user := entity.FindUserByName(name)

		if user == nil {
			invalid = true
		} else if user.SyncAllowed() {
			invalid = user.InvalidPassword(password)
		}

		// Successful?
		if user == nil || invalid {
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		basicAuth.user[raw] = *user

		c.Set(gin.AuthUserKey, user.UserUID)
	}
}
