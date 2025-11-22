package txt

// JoinAnd formats a slice of strings using commas and a localized "and" before the final element.
// Examples:
//
//	[]string{} => ""
//	[]string{"a"} => "a"
//	[]string{"a","b"} => "a and b"
//	[]string{"a","b","c"} => "a, b, and c"
func JoinAnd(values []string) string {
	length := len(values)

	switch length {
	case 0:
		return ""
	case 1:
		return values[0]
	case 2:
		return values[0] + " and " + values[1]
	}

	// length >= 3
	result := ""
	for i := 0; i < length; i++ {
		switch i {
		case 0:
			result = values[i]
		case length - 1:
			result += ", and " + values[i]
		default:
			result += ", " + values[i]
		}
	}

	return result
}
