package util

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

// Returns sha1 hash of file as string
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
