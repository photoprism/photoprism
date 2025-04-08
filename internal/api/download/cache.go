package download

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var cache = gc.New(time.Minute*15, 5*time.Minute)

// Flush resets the download cache.
func Flush() {
	cache.Flush()
}
