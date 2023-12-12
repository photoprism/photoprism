package rnd

import (
	"strconv"
	"time"
)

const (
	PrefixNone  = byte(0)
	PrefixMixed = byte('*')
)

// GenerateUID returns a unique id with prefix as string.
func GenerateUID(prefix byte) string {
	result := make([]byte, 0, 16)
	result = append(result, prefix)
	result = append(result, strconv.FormatInt(time.Now().UTC().Unix(), 36)[0:6]...)
	result = append(result, Base36(9)...)

	return string(result)
}
