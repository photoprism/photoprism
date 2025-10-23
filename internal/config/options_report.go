package config

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// Report returns global config values as a table for reporting.
func (o Options) Report() (rows [][]string, cols []string) {
	v := reflect.ValueOf(o)

	cols = []string{"Name", "Type", "CLI Flag"}
	rows = make([][]string, 0, v.NumField())

	// Iterate through all config fields.
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		yamlName := v.Type().Field(i).Tag.Get("yaml")
		flagName := v.Type().Field(i).Tag.Get("flag")

		if yamlName == "" || yamlName == "-" || flagName == "" {
			continue
		}

		// Skip options by feature set if tags are set.
		if tags := v.Type().Field(i).Tag.Get("tags"); tags == "" {
			// Report.
		} else if !slices.Contains(strings.Split(tags, ","), Features) {
			// Skip.
			continue
		}

		fieldType := fmt.Sprintf("%T", fieldValue.Interface())

		rows = append(rows, []string{yamlName, fieldType, "--" + flagName})
	}

	return rows, cols
}
