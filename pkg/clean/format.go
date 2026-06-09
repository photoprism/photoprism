package clean

import (
	"strings"
)

// FormatSep defines the string separating multiple format identifiers in a list.
const FormatSep = ","

// Format lowercases and strips the surrounding whitespace and punctuation
// from format identifiers so that inputs like `"mp4"`, ` .avi `, or `magicyuv,`
// normalize to canonical short names suitable for map lookups.
func Format(s string) string {
	return strings.ToLower(strings.Trim(s, " .,;:\"'`"))
}
