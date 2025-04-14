package download

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var expires = time.Minute * 15
var cache = gc.New(expires, 5*time.Minute)

// Flush resets the download cache.
func Flush() {
	cache.Flush()
}
