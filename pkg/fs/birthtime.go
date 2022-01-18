package fs

import (
	"time"

	"github.com/djherbis/times"
)

// BirthTime returns the creation time of a file or folder.
func BirthTime(fileName string) time.Time {
	s, err := times.Stat(fileName)

	if err != nil {
		return time.Now()
	}

	if s.HasBirthTime() {
		return s.BirthTime()
	}

	return s.ModTime()
}
