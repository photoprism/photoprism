package list

// Join combines two lists without adding duplicates.
func Join(list []string, join []string) []string {
	if len(join) == 0 {
		return list
	} else if len(list) == 0 {
		return join
	}

	for j := range join {
		if Excludes(list, join[j]) {
			list = append(list, join[j])
		}
	}

	return list
}
