package entity

import (
	"fmt"
	"reflect"
)

// Values is a shortcut for map[string]interface{}
type Values map[string]interface{}

// ModelValues extracts Values from an entity model.
func ModelValues(m interface{}, keyNames ...string) (result Values, keys []interface{}, err error) {
	isKey := func(name string) bool {
		for _, s := range keyNames {
			if name == s {
				return true
			}
		}

		return false
	}

	r := reflect.ValueOf(m)

	if r.Kind() != reflect.Pointer {
		return result, keys, fmt.Errorf("model interface expected")
	}

	values := r.Elem()

	if kind := values.Kind(); kind != reflect.Struct {
		return result, keys, fmt.Errorf("model expected")
	}

	t := values.Type()
	num := t.NumField()

	keys = make([]interface{}, 0, len(keyNames))
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

		// Skip keys.
		if isKey(fieldName) {
			if !v.IsZero() {
				keys = append(keys, v.Interface())
			}
			continue
		}

		// Add value to result.
		result[fieldName] = v.Interface()
	}

	if len(result) == 0 {
		return result, keys, fmt.Errorf("no values")
	}

	return result, keys, nil
}
