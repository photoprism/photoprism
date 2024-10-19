package entity

import (
	"runtime"
	"strconv"
	"strings"
)

// Retreive the caller from the stack
func GetCallerFormatted(callerFileName string) string {
	programCounters := [10]uintptr{}
	// The 3rd result tends to be the function that called
	numOfCallers := runtime.Callers(3, programCounters[:])
	frames := runtime.CallersFrames(programCounters[:numOfCallers])
	for i := 0; i < numOfCallers; i++ {
		frame, _ := frames.Next()
		if strings.HasPrefix(frame.File, "/go/src/github.com/photoprism/") &&
			!strings.HasSuffix(frame.File, callerFileName) {
			return string(strconv.AppendInt(append([]byte(frame.File), ':'), int64(frame.Line), 10))
		}
	}
	return ""
}
