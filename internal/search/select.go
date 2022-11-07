package search

import (
	"reflect"
	"strings"

	"github.com/photoprism/photoprism/pkg/list"
)

// Cols represents a list of database columns.
type Cols []string

// SelectString returns the columns for a search result struct as a string.
func SelectString(f interface{}, tags []string) string {
	return strings.Join(SelectCols(f, tags), ", ")
}

// SelectCols returns the columns for a search result struct.
func SelectCols(f interface{}, tags []string) Cols {
	v := reflect.ValueOf(f)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return Cols{}
	}

	cols := make(Cols, 0, v.NumField())

	// Find matching columns for struct fields.
	for i := 0; i < v.NumField(); i++ {
		s := strings.TrimSpace(v.Type().Field(i).Tag.Get("select"))

		// Serialize field values as string.
		if s == "" || s == "-" {
			continue
		} else if c := strings.Split(s, ","); c[0] == "" {
			continue
		} else if len(tags) == 0 || list.ContainsAny(c, tags) {
			cols = append(cols, c[0])
		}
	}

	return cols
}
