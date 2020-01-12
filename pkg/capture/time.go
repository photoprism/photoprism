package capture

import (
	"fmt"
	"time"
)

// Time returns the input string with the time elapsed added.
func Time(start time.Time, label string) string {
	elapsed := time.Since(start)
	return fmt.Sprintf("%s [%s]", label, elapsed)
}
