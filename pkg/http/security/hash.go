package security

import "hash/fnv"

// HashPath returns a lightweight, deterministic hash for an HTTP request path.
func HashPath(path string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(path))
	return h.Sum64()
}

// IsScanPath returns true if the request path matches a scanner path hash.
func IsScanPath(path string) bool {
	_, ok := ScanPathHashes[HashPath(path)]
	return ok
}
