package entity

import (
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
func ModelValues(m any, omit ...string) (result Values, omitted []any, err error) {
	return ModelValuesStructOption(m, true, omit...)
}

// ModelValuesStructOption extracts Values from an entity model, with the option to includeAll fields like before.
// When using this for entity Updates includeAll MUST be false, so that GormV2 is forced to behave like GormV1.
// There are two white lists which need to be maintained if new data types are used, or pointers to existing types are used.
func ModelValuesStructOption(m interface{}, includeAll bool, omit ...string) (result Values, omitted []interface{}, err error) {
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
		// log.Debugf("field %v is %v with name %v or string %v and is exported %v", fieldName, v.Kind(), v.Type().Name(), v.Type().String(), field.IsExported())

		switch v.Kind() {
		case reflect.Chan, reflect.Func, reflect.Map, reflect.UnsafePointer:
			if v.IsZero() {
				continue
			}
			if !includeAll {
				v.SetZero()
			}
			continue
		case reflect.Slice:
			if v.IsZero() {
				continue
			}
			whitelist := false
			switch v.Type().String() {
			case "json.RawMessage":
				whitelist = true
			}
			if !whitelist && !includeAll {
				v.SetZero()
				continue
			}
		case reflect.Struct:
			if v.IsZero() {
				continue
			}
			whitelist := false
			switch v.Type().String() {
			case "sql.NullTime", "time.Time", "time.Duration", "json.RawMessage", "otp.Key":
				whitelist = true
			}
			if !whitelist && !includeAll {
				v.SetZero()
				continue
			}
		case reflect.Pointer:
			whitelist := false
			switch v.Type().String() {
			case "*time.Time", "*time.Duration", "*bool", "*uint", "*uint64", "*uint32", "*int", "*int64", "*int32", "*string", "*float32", "*float64", "*otp.Key", "*sql.NullTime", "*json.RawMessage":
				whitelist = true
			}
			if !whitelist && !includeAll {
				v.SetZero()
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
