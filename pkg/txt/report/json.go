package report

import (
	"encoding/json"
)

// RowsToObjects converts a table (rows + column names) into a slice of
// objects keyed by canonicalized column names.
func RowsToObjects(rows [][]string, cols []string) []map[string]string {
	out := make([]map[string]string, 0, len(rows))
	keys := make([]string, len(cols))
	for i, c := range cols {
		keys[i] = CanonKey(c)
	}
	for _, r := range rows {
		obj := make(map[string]string, len(keys))
		for i := range keys {
			val := ""
			if i < len(r) {
				val = r[i]
			}
			obj[keys[i]] = val
		}
		out = append(out, obj)
	}
	return out
}

// JSONExport returns a JSON string for a single-table report as a top-level
// array of objects keyed by canonicalized column names.
func JSONExport(rows [][]string, cols []string) (string, error) {
	data := RowsToObjects(rows, cols)
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
