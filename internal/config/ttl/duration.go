package ttl

import "strconv"

// Duration represents a cache duration in seconds.
type Duration int

// Int returns the cache Duration in seconds as signed integer.
func (a Duration) Int() int {
	return int(a)
}

// String returns the cache Duration in seconds as string.
func (a Duration) String() string {
	return strconv.Itoa(int(a))
}
