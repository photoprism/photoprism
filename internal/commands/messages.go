package commands

import (
	"fmt"

	"github.com/dustin/go-humanize/english"
)

// formatCount returns a human-readable count phrase with the correct noun form.
func formatCount(count int, singular, plural string) string {
	return english.Plural(count, singular, plural)
}

// formatFailedCount returns a human-readable failure phrase with the correct noun form.
func formatFailedCount(count int, singular, plural string) string {
	return fmt.Sprintf("%s failed", formatCount(count, singular, plural))
}
