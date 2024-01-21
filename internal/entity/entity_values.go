package entity

import (
	"fmt"
	"reflect"
)

// Map is an alias for map[string]interface{}.
type Map map[string]interface{}

// ModelValues extracts Values from an entity model.
func ModelValues(m interface{}, omit ...string) (result Map, omitted []interface{}, err error) {
	mustOmit := func(name string) bool {
		for _, s := range omit {
			if name == s {
				return true
			}
		}

		return false
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

	omitted = make([]interface{}, 0, len(omit))
	result = make(map[string]interface{}, num)

	// Add exported fields to result.
	for i := 0; i < num; i++ {
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
		case reflect.Slice, reflect.Chan, reflect.Func, reflect.Map, reflect.UnsafePointer:
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
