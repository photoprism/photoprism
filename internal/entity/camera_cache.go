package entity

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var cameraCache = gc.New(time.Hour, 15*time.Minute)

func FlushCameraCache() {
	cameraCache.Flush()
}
