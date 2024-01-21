package fs

import (
	"fmt"
	"path"
)

// CachePath returns a cache directory name based on the base path, file hash and cache namespace.
func CachePath(basePath, fileHash, namespace string, create bool) (cachePath string, err error) {
	if len(fileHash) < 4 {
		return "", fmt.Errorf("cache: hash '%s' is too short", fileHash)
	}

	if namespace == "" {
		return "", fmt.Errorf("cache: namespace for hash '%s' is empty", fileHash)
	}

	cachePath = path.Join(basePath, namespace, fileHash[0:1], fileHash[1:2], fileHash[2:3])

	if create {
		if err = MkdirAll(cachePath); err != nil {
			return "", err
		}
	}

	return cachePath, nil
}
