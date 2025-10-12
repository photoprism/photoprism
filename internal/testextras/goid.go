package testextras

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// GetGoID returns the underlying goid (thread) for the caller.
// This is a debugging function.
func GetGoID() int64 {
	var buf [64]byte

	stackLen := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:stackLen]), "goroutine "))[0]
	id, err := strconv.ParseInt(idField, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}
