package photoprism

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/fs"

	"github.com/photoprism/photoprism/internal/entity"
)

func FileName(fileRoot, fileName string) string {
	switch fileRoot {
	case entity.RootSidecar:
		return path.Join(Config().SidecarPath(), fileName)
	case entity.RootImport:
		return path.Join(Config().ImportPath(), fileName)
	case entity.RootExamples:
		return path.Join(Config().ExamplesPath(), fileName)
	default:
		return path.Join(Config().OriginalsPath(), fileName)
	}
}

// CachePath returns a cache directory name based on the base path, file hash and cache namespace.
func CachePath(fileHash, namespace string) (cachePath string, err error) {
	return fs.CachePath(Config().CachePath(), fileHash, namespace, true)
}

// CacheName returns an absolute cache file name based on the base path, file hash and cache namespace.
func CacheName(fileHash, namespace, cacheKey string) (cacheName string, err error) {
	if cacheKey == "" {
		return "", fmt.Errorf("cache: key for hash '%s' is empty", fileHash)
	}

	cachePath, err := CachePath(fileHash, namespace)

	if err != nil {
		return "", err
	}

	cacheName = filepath.Join(cachePath, fmt.Sprintf("%s_%s", fileHash, cacheKey))

	return cacheName, nil
}
