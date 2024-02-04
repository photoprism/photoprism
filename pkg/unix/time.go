package unix

import "time"

// Time returns the current time in seconds since January 1, 1970 UTC.
func Time() int64 {
	return time.Now().UTC().Unix()
}
