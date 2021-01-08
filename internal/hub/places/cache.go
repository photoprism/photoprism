package places

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var cache = gc.New(time.Hour*4, 10*time.Minute)
