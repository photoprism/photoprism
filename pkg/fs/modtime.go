package fs

import (
	"time"

	"github.com/djherbis/times"
)

// ModTime returns the last modification time of a file or directory in UTC.
func ModTime(filePath string) time.Time {
	stat, err := times.Stat(filePath)

	// Return the current time if Stat() call failed.
	if err != nil {
		return time.Now().UTC()
	}

	// Return modification time as reported by Stat().
	return stat.ModTime().UTC()
}
