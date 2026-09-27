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

// formatVideoSummary returns the closing summary of a video command, which in a dry run counts the
// files that would be processed instead of reporting them as skipped.
func formatVideoSummary(action string, dryRun bool, planned, processed, skipped, failed int) string {
	if dryRun {
		return fmt.Sprintf("%s: would process %s, skipped %s", action,
			formatCount(planned, "file", "files"), formatCount(skipped, "file", "files"))
	}

	return fmt.Sprintf("%s: processed %s, skipped %s, %s", action,
		formatCount(processed, "file", "files"), formatCount(skipped, "file", "files"),
		formatFailedCount(failed, "file", "files"))
}
