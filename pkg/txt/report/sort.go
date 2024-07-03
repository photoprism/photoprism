package report

import (
	"sort"
)

// Sort sorts the report rows.
func Sort(rows [][]string) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i][0] == rows[j][0] {
			return rows[i][1] < rows[j][1]
		} else {
			return rows[i][0] < rows[j][0]
		}
	})
}
