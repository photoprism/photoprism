package clean

// SqlAliasMax is the maximum length of a table alias, which is well above any the code uses and
// far below the identifier limits the supported databases enforce.
const SqlAliasMax = 24

// SqlAlias returns a table alias that is safe to interpolate into a statement, or an empty string.
//
// An alias cannot be bound as a parameter, so it is the one part of a statement a caller may be
// tempted to concatenate. Anything that is not a bare identifier is rejected rather than stripped:
// a rejected alias yields unqualified columns and therefore an error, where a stripped one would
// silently name a different table.
func SqlAlias(s string) string {
	if s == "" || len(s) > SqlAliasMax {
		return ""
	}

	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_':
			// Always allowed.
		case i > 0 && r >= '0' && r <= '9':
			// Allowed after the first character, as SQL identifiers may not start with a digit.
		default:
			return ""
		}
	}

	return s
}
