package rnd

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strconv"
)

// Token returns a random token with length of up to 10 characters.
func Token(size uint) string {
	if size > 10 || size < 1 {
		panic(fmt.Sprintf("size out of range: %d", size))
	}

	result := make([]byte, 0, 14)
	b := make([]byte, 8)

	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	randomInt := binary.BigEndian.Uint64(b)

	result = append(result, strconv.FormatUint(randomInt, 36)...)

	for i := len(result); i < cap(result); i++ {
		result = append(result, byte(123-(cap(result)-i)))
	}

	return string(result[:size])
}
