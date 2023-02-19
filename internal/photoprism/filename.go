package photoprism

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

// FileName returns the full file name based on the root folder type.
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

// ExifToolCacheName returns the ExifTool metadata cache file name.
func ExifToolCacheName(hash string) (string, error) {
	return CacheName(hash, "json", "exiftool.json")
}

// RelName returns the relative filename.
func RelName(fileName, directory string) string {
	return fs.RelName(fileName, directory)
}

// RootPath returns the file root path based on the configuration.
func RootPath(fileName string) string {
	switch Root(fileName) {
	case entity.RootSidecar:
		return Config().SidecarPath()
	case entity.RootImport:
		return Config().ImportPath()
	case entity.RootExamples:
		return Config().ExamplesPath()
	default:
		return Config().OriginalsPath()
	}
}

// Root returns the file root directory.
func Root(fileName string) string {
	originalsPath := Config().OriginalsPath()

	if originalsPath != "" && strings.HasPrefix(fileName, originalsPath) {
		return entity.RootOriginals
	}

	importPath := Config().ImportPath()

	if importPath != "" && strings.HasPrefix(fileName, importPath) {
		return entity.RootImport
	}

	sidecarPath := Config().SidecarPath()

	if sidecarPath != "" && strings.HasPrefix(fileName, sidecarPath) {
		return entity.RootSidecar
	}

	examplesPath := Config().ExamplesPath()

	if examplesPath != "" && strings.HasPrefix(fileName, examplesPath) {
		return entity.RootExamples
	}

	return entity.RootUnknown
}

// RootRelName returns the relative filename, and automatically detects the root path.
func RootRelName(fileName string) string {
	return RelName(fileName, RootPath(fileName))
}
