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

// CacheWebDAVUser caches a successful credential check if its starting generation is current.
func CacheWebDAVUser(key string, user *User, generation AuthCacheGeneration) {
	authCacheState.Lock()
	defer authCacheState.Unlock()

	if key != "" && user != nil && generation.valid(user.UserUID) {
		webDAVUserCache.SetDefault(key, user)
	}
}
