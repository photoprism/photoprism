package rnd

import (
	"strconv"
	"time"
)

// PPID returns a unique id with prefix as string.
func PPID(prefix byte) string {
	result := make([]byte, 0, 16)
	result = append(result, prefix)
	result = append(result, strconv.FormatInt(time.Now().UTC().Unix(), 36)[0:6]...)
	result = append(result, Token(9)...)

	return string(result)
}

// IsPPID returns true if the id seems to be a PhotoPrism unique id.
func IsPPID(id string, prefix byte) bool {
	if len(id) != 16 {
		return false
	}

	return id[0] == prefix
}
