package fs

import (
	"crypto/sha1"
	"encoding/hex"
	"hash/crc32"
	"io"
	"os"
)

// Hash returns the SHA1 hash of a file as string.
func Hash(filename string) string {
	var result []byte

	file, err := os.Open(filename)

	if err != nil {
		return ""
	}

	defer file.Close()

	hash := sha1.New()

	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(result))
}

// Checksum returns the CRC32 checksum of a file as string.
func Checksum(filename string) string {
	var result []byte

	file, err := os.Open(filename)

	if err != nil {
		return ""
	}

	defer file.Close()

	hash := crc32.New(crc32.MakeTable(crc32.Castagnoli))

	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(result))
}
