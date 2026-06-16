package txt

// JoinAnd formats a slice of strings as a list with commas and "and" before the final element.
// Examples:
//
//	[]string{} => ""
//	[]string{"a"} => "a"
//	[]string{"a","b"} => "a and b"
//	[]string{"a","b","c"} => "a, b, and c"
func JoinAnd(values []string) string {
	return joinConjunction(values, "and")
}

// JoinOr formats a slice of strings using commas and "or" before the final element.
// Examples:
//
//	[]string{} => ""
//	[]string{"a"} => "a"
//	[]string{"a","b"} => "a or b"
//	[]string{"a","b","c"} => "a, b, or c"
func JoinOr(values []string) string {
	return joinConjunction(values, "or")
}

// joinConjunction joins values into a grammatical list, placing conj before the
// final element and an Oxford comma after the penultimate one for three or more.
func joinConjunction(values []string, conj string) string {
	length := len(values)

	switch length {
	case 0:
		return ""
	case 1:
		return values[0]
	case 2:
		return values[0] + " " + conj + " " + values[1]
	}

	// length >= 3
	result := ""
	for i := range length {
		switch i {
		case 0:
			result = values[i]
		case length - 1:
			result += ", " + conj + " " + values[i]
		default:
			result += ", " + values[i]
		}
	}

	return result
}
