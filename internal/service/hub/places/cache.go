package places

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var clientCache = gc.New(time.Hour*4, 10*time.Minute)
