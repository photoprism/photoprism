package entity

import (
	"time"
)

// Day specified as time.Duration to improve readability.
const Day = time.Hour * 24

// UTC returns the current Coordinated Universal Time (UTC).
func UTC() time.Time {
	return time.Now().UTC()
}

// Now returns the current time in UTC, truncated to seconds.
func Now() time.Time {
	return UTC().Truncate(time.Second)
}

// TimeStamp returns a reference to the current time in UTC, truncated to seconds.
func TimeStamp() *time.Time {
	t := Now()
	return &t
}

// Time returns a reference to the specified time in UTC, truncated to seconds, or nil if it is invalid.
func Time(s string) *time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		t = t.UTC().Truncate(time.Second)
		return &t
	}

	return nil
}

// Seconds converts an int to a duration in seconds.
func Seconds(s int) time.Duration {
	return time.Duration(s) * time.Second
}

// Yesterday returns the time 24 hours ago.
func Yesterday() time.Time {
	return UTC().Add(-24 * time.Hour)
}
