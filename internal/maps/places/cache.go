package places

import (
	"time"

	gc "github.com/patrickmn/go-cache"
)

var cache = gc.New(15*time.Minute, 5*time.Minute)
