package download

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Find returns the fileName for the given download id or an error if the id is invalid.
func Find(uniqueId string) (fileName string, err error) {
	if uniqueId == "" || !rnd.IsUUID(uniqueId) {
		return fileName, fmt.Errorf("id has an invalid format")
	}

	// Cached?
	if cacheData, hit := cache.Get(uniqueId); hit {
		log.Tracef("download: cache hit for %s", uniqueId)
		return cacheData.(string), nil
	}

	return "", fmt.Errorf("invalid id %s", clean.LogQuote(uniqueId))
}
