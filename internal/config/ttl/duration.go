package ttl

import (
	"strconv"
	"time"
)

// Duration represents a cache duration in seconds.
type Duration int

// Int returns the cache Duration in seconds as signed integer.
func (a Duration) Int() int {
	return int(a)
}

// Duration returns the cache Duration as time.Duration.
func (a Duration) Duration() time.Duration {
	return time.Duration(a) * time.Second
}

// String returns the cache Duration in seconds as string.
func (a Duration) String() string {
	return strconv.Itoa(int(a))
}
