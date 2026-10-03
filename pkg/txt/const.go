package txt

import (
	"github.com/photoprism/photoprism/pkg/enum"
)

// Masked replaces a credential that must not be shown, in a log line, a report or an API
// response. Asterisks rather than letters, which could themselves be somebody's password.
const Masked = "***"

// True and False specify boolean string representations.
const (
	True  = enum.True
	False = enum.False
)

// Additional English language strings.
const (
	EnOr   = "or"
	EnAnd  = "and"
	EnWith = "with"
	EnIn   = "in"
	EnAt   = "at"
	EnNew  = "new"
)
