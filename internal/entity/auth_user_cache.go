package entity

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

// webDAVUserCache retains authenticated users between WebDAV requests.
var webDAVUserCache = gc.New(time.Minute, time.Minute)

// CachedWebDAVUser returns the user authenticated by the cached credential, if any.
func CachedWebDAVUser(key string) *User {
	if value, found := webDAVUserCache.Get(key); found {
		user, _ := value.(*User)
		return user
	}

	return nil
}

// CacheWebDAVUser caches a successful WebDAV credential check for one minute.
func CacheWebDAVUser(key string, user *User) {
	if key != "" && user != nil {
		webDAVUserCache.SetDefault(key, user)
	}
}
