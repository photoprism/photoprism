package rnd

import (
	"fmt"
	"hash/crc32"
	"strconv"
)

// CrcToken returns a string token with checksum.
func CrcToken() string {
	token := make([]byte, 0, 14)

	token = append(token, []byte(Base36(4))...)
	token = append(token, '-')
	token = append(token, []byte(Base36(4))...)

	checksum := crc32.ChecksumIEEE(token)

	sum := strconv.FormatInt(int64(checksum), 16)

	return fmt.Sprintf("%s-%.4s", token, sum)
}

// ValidateCrcToken tests if the token string is valid.
func ValidateCrcToken(s string) bool {
	if len(s) != 14 {
		return false
	}

	token := []byte(s[:9])

	checksum := crc32.ChecksumIEEE(token)

	sum := strconv.FormatInt(int64(checksum), 16)

	return s == fmt.Sprintf("%s-%.4s", token, sum)
}
