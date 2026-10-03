package entity

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
)

// Values is a shorthand alias for map[string]interface{}.
type Values = map[string]any

// Map is retained for backward compatibility.
// TODO: Remove when no longer needed.
type Map = Values

// ModelValues extracts exported struct fields into a Values map, optionally omitting selected names.
// Byte slices such as json.RawMessage columns are included; other slices, maps, and relations are not.
func ModelValues(m any, omit ...string) (result Values, omitted []any, err error) {
	mustOmit := func(name string) bool {
		return slices.Contains(omit, name)
	}

	r := reflect.ValueOf(m)

	if r.Kind() != reflect.Pointer {
		return result, omitted, fmt.Errorf("model interface expected")
	}

	values := r.Elem()

	if kind := values.Kind(); kind != reflect.Struct {
		return result, omitted, fmt.Errorf("model expected")
	}

	t := values.Type()
	num := t.NumField()

	omitted = make([]any, 0, len(omit))
	result = make(Values, num)

	// Add exported fields to result.
	for i := range num {
		field := t.Field(i)

		// Skip non-exported fields.
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name

		// Skip timestamps.
		if fieldName == "" || fieldName == "UpdatedAt" || fieldName == "CreatedAt" {
			continue
		}

		v := values.Field(i)

		switch v.Kind() {
		case reflect.Slice:
			if v.Type().Elem().Kind() != reflect.Uint8 {
				continue
			}
		case reflect.Chan, reflect.Func, reflect.Map, reflect.UnsafePointer:
			continue
		case reflect.Struct:
			if v.IsZero() {
				continue
			}
		}

		// Skip read-only fields.
		if !v.CanSet() {
			continue
		}

		// Skip omitted.
		if mustOmit(fieldName) {
			if !v.IsZero() {
				omitted = append(omitted, v.Interface())
			}
			continue
		}

		// Add value to result.
		result[fieldName] = v.Interface()
	}

	if len(result) == 0 {
		return result, omitted, fmt.Errorf("no values")
	}

	return result, omitted, nil
}

// reportValue formats a model value for a report, showing byte slices such as JSON columns as text.
func reportValue(v any) string {
	if b, ok := v.([]byte); ok {
		return string(b)
	} else if b, ok := v.(json.RawMessage); ok {
		return string(b)
	}

	return fmt.Sprintf("%#v", v)
}
