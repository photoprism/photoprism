package report

import (
	"encoding/json"
	"fmt"
)

// RowsToObjects converts a table (rows + column names) into a slice of
// objects keyed by canonicalized column names.
func RowsToObjects(rows [][]string, cols []string) []map[string]string {
	return RowsToObjectsKeys(rows, cols, nil)
}

// RowsToObjectsKeys converts a table into objects under the given field names, falling back to the
// canonicalized column names for any the caller left out.
func RowsToObjectsKeys(rows [][]string, cols []string, keys []string) []map[string]string {
	out := make([]map[string]string, 0, len(rows))

	if len(keys) < len(cols) {
		named := uniqueKeys(cols)
		copy(named, keys)
		keys = named
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

// uniqueKeys canonicalizes the column names and suffixes any that repeat, so a table holding two
// columns of the same name exports both. Keyed by name, one would otherwise overwrite the other and
// the export would lose a column without saying so.
func uniqueKeys(cols []string) []string {
	keys := make([]string, len(cols))
	seen := make(map[string]int, len(cols))

	for i, c := range cols {
		k := CanonKey(c)

		if n := seen[k]; n > 0 {
			seen[k] = n + 1
			k = fmt.Sprintf("%s_%d", k, n+1)
		} else {
			seen[k] = 1
		}

		keys[i] = k
	}

	return keys
}

// JSONExport returns a JSON string for a single-table report as a top-level
// array of objects keyed by canonicalized column names.
func JSONExport(rows [][]string, cols []string) (string, error) {
	return JSONExportKeys(rows, cols, nil)
}

// JSONExportKeys returns a JSON string keyed by the given field names.
func JSONExportKeys(rows [][]string, cols []string, keys []string) (string, error) {
	data := RowsToObjectsKeys(rows, cols, keys)
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
