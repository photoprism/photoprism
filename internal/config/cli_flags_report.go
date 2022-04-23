package config

// Report returns global config values as a table for reporting.
func (f CliFlags) Report() (rows [][]string, cols []string) {
	cols = []string{"Variable", "Flag", "Usage"}

	rows = make([][]string, 0, len(f))

	for _, flag := range Flags {
		if flag.Hidden() {
			continue
		}

		row := []string{flag.EnvVar(), flag.Name(), flag.Usage()}
		rows = append(rows, row)
	}

	return rows, cols
}
