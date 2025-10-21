package header

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SetLocation adds a Location header with a relative path based on the provided segments.
// When the first segment is non-empty it is treated as the base path;
// otherwise the request URL path is used.
func SetLocation(c *gin.Context, segments ...string) {
	// Return if context is missing.
	if c == nil {
		return
	}

	base := ""

	if len(segments) > 0 && segments[0] != "" {
		base = segments[0]
		segments = segments[1:]
	} else if c.Request != nil && c.Request.URL != nil {
		base = c.Request.URL.Path
	}

	// Return if base is missing.
	if base == "" {
		return
	}

	// Compose redirect location string.
	prefixSlash := strings.HasPrefix(base, "/")
	base = strings.Trim(base, "/")

	parts := make([]string, 0, 1+len(segments))
	if base != "" {
		parts = append(parts, base)
	}

	for _, segment := range segments {
		segment = strings.Trim(segment, "/")
		if segment == "" {
			continue
		}
		parts = append(parts, segment)
	}

	location := strings.Join(parts, "/")
	if prefixSlash {
		location = "/" + location
	}

	// Add Location header to response.
	if location == "" && prefixSlash {
		c.Header(Location, "/")
	} else {
		c.Header(Location, location)
	}
}
