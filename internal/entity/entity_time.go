package entity

import (
	"time"
)

// Day specified as time.Duration to improve readability.
const Day = time.Hour * 24

// UnixMinute is one minute in UnixTime.
const UnixMinute int64 = 60

// UnixHour is one hour in UnixTime.
const UnixHour = UnixMinute * 60

// UnixDay is one day in UnixTime.
const UnixDay = UnixHour * 24

// UnixWeek is one week in UnixTime.
const UnixWeek = UnixDay * 7

// UnixMonth is about one month in UnixTime.
const UnixMonth = UnixDay * 31

// UnixYear is about one year in UnixTime.
const UnixYear = UnixDay * 365

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
