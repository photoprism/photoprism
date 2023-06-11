package form

import (
	"fmt"
	"reflect"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
)

// Report returns form fields as table rows for reports.
func Report(f interface{}) (rows [][]string, cols []string) {
	cols = []string{"Filter", "Type", "Examples", "Notes"}

	v := reflect.ValueOf(f)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return rows, cols
	}

	rows = make([][]string, 0, v.NumField())

	// Iterate through all form fields.
	for i := 0; i < v.NumField(); i++ {
		// Skip unexported fields.
		if !v.Type().Field(i).IsExported() {
			continue
		}

		// Get info from struct field tags.
		fieldValue := v.Field(i)
		fieldName := v.Type().Field(i).Tag.Get("form")
		fieldInfo := v.Type().Field(i).Tag.Get("serialize")
		notes := v.Type().Field(i).Tag.Get("notes")

		// Serialize field values as string.
		if fieldName != "" && fieldName != "q" && fieldInfo != "-" && notes != "-" {
			example := v.Type().Field(i).Tag.Get("example")
			typeName := "any"

			switch t := fieldValue.Interface().(type) {
			case time.Time:
				typeName = "timestamp"
				if example == "" {
					example = fmt.Sprintf("%s:\"2022-01-30\"", fieldName)
				}
			case int, int8, int16, int32, int64:
				typeName = "number"
				if example == "" {
					example = fmt.Sprintf("%s:0 %s:3", fieldName, fieldName)
				}
			case uint, uint8, uint16, uint32, uint64:
				typeName = "number"
				if example == "" {
					example = fmt.Sprintf("%s:-1 %s:2", fieldName, fieldName)
				}
			case float32, float64:
				typeName = "decimal"
				if example == "" {
					example = fmt.Sprintf("%s:1.245", fieldName)
				}
			case string:
				typeName = "string"
				if example == "" {
					example = fmt.Sprintf("%s:\"name\"", fieldName)
				}
			case bool:
				typeName = "switch"
				if example == "" {
					example = fmt.Sprintf("%s:yes", fieldName)
				}
			default:
				log.Warnf("failed reporting on %T %s", t, clean.Token(fieldName))
				continue
			}

			rows = append(rows, []string{fieldName, typeName, example, notes})
		}
	}

	return rows, cols
}
