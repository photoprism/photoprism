package rnd

import (
	"strconv"
	"time"
)

const (
	// PrefixNone indicates no UID prefix is set.
	PrefixNone = byte(0)
	// PrefixMixed marks a mixed set of UID prefixes.
	PrefixMixed = byte('*')
)

// GenerateUID returns a unique id with prefix as string.
func GenerateUID(prefix byte) string {
	return generateUID(prefix, time.Now())
}

// generateUID returns a unique id with prefix as string at a given time.
func generateUID(prefix byte, t time.Time) string {
	result := make([]byte, 0, 16)
	result = append(result, prefix)
	result = append(result, strconv.FormatInt(t.UTC().Unix(), 36)[0:6]...)
	result = append(result, Base36(9)...)

	return string(result)
}
