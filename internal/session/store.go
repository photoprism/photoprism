package session

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var cache = gc.New(72*time.Hour, 30*time.Minute)

func Create(data interface{}) string {
	token := Token()
	cache.Set(token, data, gc.DefaultExpiration)
	log.Debugf("session: created")
	return token
}

func Delete(token string) {
	cache.Delete(token)
	log.Debugf("session: deleted")
}

func Get(token string) (data interface{}, exists bool) {
	return cache.Get(token)
}

func Exists(token string) bool {
	_, found := cache.Get(token)

	return found
}
