package event

import (
	"fmt"
	"slices"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
)

// MessageSep separates event topic segments when rendered as text.
var MessageSep = " › "

// segmentSepMarker replaces a field separator carried by a segment.
const segmentSepMarker = "?"

// Format joins the segments of an event with MessageSep and, when arguments are supplied,
// renders the result as their format string. Segments are structure and arguments are values,
// so the separator is removed from the former and the latter stay the caller's to sanitize.
func Format(ev []string, args ...any) string {
	s := strings.Join(formatSegments(ev), MessageSep)

	if len(args) == 0 {
		return s
	}

	return fmt.Sprintf(s, args...)
}

// formatSegments replaces the field separator wherever a segment carries one, so that a segment
// cannot carry it into the joined text. The slice the caller passed is left as it is.
func formatSegments(ev []string) []string {
	out := ev
	cloned := false

	for i, s := range ev {
		if !strings.ContainsRune(s, clean.FieldSep) {
			continue
		}

		if !cloned {
			out = slices.Clone(ev)
			cloned = true
		}

		out[i] = strings.ReplaceAll(s, string(clean.FieldSep), segmentSepMarker)
	}

	return out
}
