package txt

import (
	"fmt"
)

// NTimes converts an integer to a string in the format "n times" or returns an empty string if n is 0.
func NTimes(n int) string {
	if n < 2 {
		return ""
	} else {
		return fmt.Sprintf("%d times", n)
	}
}
