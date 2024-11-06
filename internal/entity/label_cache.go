package entity

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var labelCache = gc.New(15*time.Minute, 15*time.Minute)

// FlushLabelCache resets the label cache.
func FlushLabelCache() {
	labelCache.Flush()
}
