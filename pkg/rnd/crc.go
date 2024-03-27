package rnd

import (
	"fmt"
	"strconv"

	"github.com/photoprism/photoprism/pkg/checksum"
)

// CrcToken returns a string token with checksum.
func CrcToken() string {
	token := make([]byte, 0, 14)

	token = append(token, []byte(Base36(4))...)
	token = append(token, '-')
	token = append(token, []byte(Base36(4))...)

	crc := checksum.Crc32(token)
	sum := strconv.FormatInt(int64(crc), 16)

	return fmt.Sprintf("%s-%.4s", token, sum)
}

// ValidateCrcToken tests if the token string is valid.
func ValidateCrcToken(s string) bool {
	if len(s) != 14 {
		return false
	}

	token := []byte(s[:9])

	crc := checksum.Crc32(token)
	sum := strconv.FormatInt(int64(crc), 16)

	return s == fmt.Sprintf("%s-%.4s", token, sum)
}
