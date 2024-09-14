package config

import "strings"

// Report returns global config values as a table for reporting.
func (f CliFlags) Report() (rows [][]string, cols []string) {
	cols = []string{"Environment", "CLI Flag", "Default", "Description"}

	rows = make([][]string, 0, len(f))

	for _, flag := range Flags {
		if flag.Hidden() {
			continue
		}

		rows = append(rows, []string{strings.ReplaceAll(flag.EnvVar(), ",", ", "), flag.CommandFlag(), flag.Default(), flag.Usage()})
	}

	return rows, cols
}
