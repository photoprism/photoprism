package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type MaxAge string

var (
	CacheShort MaxAge = "3600"
	CacheLong  MaxAge = "86400"
)

// AddCacheHeader adds a cache control header to the response.
func AddCacheHeader(c *gin.Context, maxAge MaxAge) {
	c.Header("Cache-Control", fmt.Sprintf("private, max-age=%s, no-transform", maxAge))
}
