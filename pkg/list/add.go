package list

// Add adds a string to the list if it does not exist yet.
func Add(list []string, s string) []string {
	switch {
	case s == "":
		return list
	case len(list) == 0:
		return []string{s}
	case Contains(list, s):
		return list
	default:
		return append(list, s)
	}
}
