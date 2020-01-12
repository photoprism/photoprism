package rnd

import (
	"strconv"
	"time"
)

// PPID returns a unique id with prefix as string.
func PPID(prefix rune) string {
	result := make([]byte, 0, 17)
	result = append(result, byte(prefix))
	result = append(result, strconv.FormatInt(time.Now().UTC().Unix(), 36)[0:6]...)
	result = append(result, Token(10)...)

	return string(result)
}
