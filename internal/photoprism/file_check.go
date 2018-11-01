package photoprism

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)

	return err == nil && !info.IsDir()
}

func fileHash(filename string) string {
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
