package util

import (
	"crypto/rand"
	"encoding/binary"
	"strconv"
	"time"

	uuid "github.com/satori/go.uuid"
)

func RandomToken(size uint) string {
	if size > 10 || size < 1 {
		log.Fatalf("size out of range: %d", size)
	}

	result := make([]byte, 0, 14)
	b := make([]byte, 8)

	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}

	randomInt := binary.BigEndian.Uint64(b)

	result = append(result, strconv.FormatUint(randomInt, 36)...)

	for i := len(result); i < cap(result); i++ {
		result = append(result, byte(123-(cap(result)-i)))
	}

	return string(result[:size])
}

func RandomPassword() string {
	return RandomToken(8)
}

func ID() string {
	result := make([]byte, 0, 16)
	result = append(result, strconv.FormatInt(time.Now().UTC().Unix(), 36)[0:6]...)
	result = append(result, RandomToken(10)...)

	return string(result)
}

func UUID() string {
	return uuid.NewV4().String()
}
