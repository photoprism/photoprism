package event

import "time"

const Day = time.Hour * 24

// TimeStamp returns the current timestamp in UTC rounded to seconds.
func TimeStamp() time.Time {
	return time.Now().UTC().Truncate(time.Second)
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
	return time.Now().Add(-24 * time.Hour)
}
