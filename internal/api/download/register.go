package download

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Register makes the specified file available for download with the
// returned id until the cache expires, or the server is restarted.
func Register(fileName string) (string, error) {
	if !fs.FileExists(fileName) {
		return "", fmt.Errorf("%s does not exists", clean.Log(fileName))
	}

	uniqueId := rnd.UUID()
	cache.SetDefault(uniqueId, fileName)

	return uniqueId, nil
}
