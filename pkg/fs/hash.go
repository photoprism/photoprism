package fs

import (
	"crypto/sha1" //nolint:gosec // SHA1 retained for legacy hash compatibility
	"encoding/hex"
	"hash/crc32"
	"io"
	"os"

	"github.com/photoprism/photoprism/pkg/checksum"
)

// Hash returns the SHA1 hash of a file as string.
func Hash(fileName string) string {
	var result []byte

	file, err := os.Open(fileName) //nolint:gosec // caller-controlled path; intended file read

	if err != nil {
		return ""
	}

	defer file.Close()

	hash := sha1.New() //nolint:gosec // legacy SHA1 hashes retained for compatibility

	buf := getCopyBuffer()
	defer putCopyBuffer(buf)

	if _, err = io.CopyBuffer(hash, file, buf); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(result))
}

// Checksum returns the CRC32 checksum of a file as string.
func Checksum(fileName string) string {
	var result []byte

	file, err := os.Open(fileName) //nolint:gosec // caller-controlled path; intended file read

	if err != nil {
		return ""
	}

	defer file.Close()

	hash := crc32.New(checksum.Crc32Castagnoli)

	buf := getCopyBuffer()
	defer putCopyBuffer(buf)

	if _, err = io.CopyBuffer(hash, file, buf); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(result))
}

// IsHash tests if a string looks like a hash.
func IsHash(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if (r < 48 || r > 57) && (r < 97 || r > 102) && (r < 65 || r > 70) {
			return false
		}
	}

	switch len(s) {
	case 8, 16, 32, 40, 56, 64, 80, 128, 256:
		return true
	}

	return false
}
