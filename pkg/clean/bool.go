package clean

import "github.com/photoprism/photoprism/pkg/txt"

// Bool interprets s as a boolean using txt.Bool, which supports strconv.ParseBool
// tokens as well as localized yes/no mappings.
func Bool(s string) bool {
	return txt.Bool(s)
}

// Yes reports whether s matches a supported affirmative token. It mirrors txt.Yes
// so callers can stay within the clean package when sanitizing inputs.
func Yes(s string) bool {
	return txt.Yes(s)
}

// No reports whether s matches a supported negative token. It mirrors txt.No.
func No(s string) bool {
	return txt.No(s)
}
