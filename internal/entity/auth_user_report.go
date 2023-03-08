package entity

import (
	"fmt"
)

// Report returns the entity values as rows.
func (m *User) Report(skipEmpty bool) (rows [][]string, cols []string) {
	cols = []string{"Name", "Value"}

	// Extract model values.
	values, _, err := ModelValues(m, "ID")

	// Ok?
	if err != nil {
		return rows, cols
	}

	rows = make([][]string, 0, len(values))

	for k, v := range values {
		s := fmt.Sprintf("%#v", v)

		// Skip empty values?
		if !skipEmpty || s != "" {
			rows = append(rows, []string{k, s})
		}
	}

	return rows, cols
}
