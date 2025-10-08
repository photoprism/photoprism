package list

// Join combines two lists without adding duplicates.
func Join(list []string, join []string) []string {
	if len(join) == 0 {
		return list
	} else if len(list) == 0 {
		// Return a copy to avoid surprising aliasing when callers append later.
		out := make([]string, len(join))
		copy(out, join)
		return out
	}

	// Build a set of existing values for O(n+m) merging without duplicates.
	set := make(map[string]struct{}, len(list)+len(join))
	out := make([]string, 0, len(list)+len(join))
	for i := range list {
		v := list[i]
		if _, ok := set[v]; !ok {
			set[v] = struct{}{}
			out = append(out, v)
		}
	}
	for j := range join {
		v := join[j]
		if v == "" {
			continue
		}
		if _, ok := set[v]; !ok {
			set[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}
