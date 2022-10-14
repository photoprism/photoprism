package entity

import (
	"time"
)

// Day specified as time.Duration to improve readability.
const Day = time.Hour * 24

// UnixHour is one hour in UnixTime.
const UnixHour int64 = 3600

// UnixDay is one day in UnixTime.
const UnixDay = UnixHour * 24

// UnixWeek is one week in UnixTime.
const UnixWeek = UnixDay * 7

// UTC returns the current Coordinated Universal Time (UTC).
func UTC() time.Time {
	return time.Now().UTC()
}

// UnixTime returns the current time in seconds since January 1, 1970 UTC.
func UnixTime() int64 {
	return UTC().Unix()
}

// TimeStamp returns the current timestamp in UTC rounded to seconds.
func TimeStamp() time.Time {
	return UTC().Truncate(time.Second)
}

// TimePointer returns a pointer to the current timestamp.
func TimePointer() *time.Time {
	t := TimeStamp()
	return &t
}

// Seconds converts an int to a duration in seconds.
func Seconds(s int) time.Duration {
	return time.Duration(s) * time.Second
}

// Yesterday returns the time 24 hours ago.
func Yesterday() time.Time {
	return UTC().Add(-24 * time.Hour)
}
