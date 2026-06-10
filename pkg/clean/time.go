package clean

import "time"

// TimeSet reports whether t points to a meaningful timestamp, i.e. it is non-nil
// and not the zero time. Useful for nullable *time.Time values where both nil and
// a zero value mean "unset".
func TimeSet(t *time.Time) bool {
	return t != nil && !t.IsZero()
}
