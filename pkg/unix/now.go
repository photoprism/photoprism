package unix

import "time"

// Now returns the current time in seconds since January 1, 1970 UTC.
func Now() int64 {
	return time.Now().UTC().Unix()
}
