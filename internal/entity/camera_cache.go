package entity

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var cameraCache = gc.New(time.Hour, 15*time.Minute)

// FlushCameraCache removes all cached cameras.
func FlushCameraCache() {
	cameraCache.Flush()
}
