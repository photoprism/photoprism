package txt

import (
	"fmt"
	"time"
)

// TimeStamp converts a time to a timestamp string for reporting.
func TimeStamp(t *time.Time) string {
	if t == nil {
		return ""
	} else if t.IsZero() {
		return ""
	}

	return t.UTC().Format("2006-01-02 15:04:05")
}

// NTimes converts an integer to a string in the format "n times" or returns an empty string if n is 0.
func NTimes(n int) string {
	if n < 2 {
		return ""
	} else {
		return fmt.Sprintf("%d times", n)
	}
}
