package entity

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var lensCache = gc.New(time.Hour, 15*time.Minute)

func FlushLensCache() {
	lensCache.Flush()
}
