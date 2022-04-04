package entity

import (
	"reflect"
)

// Values is a shortcut for map[string]interface{}
type Values map[string]interface{}

// GetValues extracts entity Values.
func GetValues(m interface{}, omit ...string) (result Values) {
	skip := func(name string) bool {
		if name == "" || name == "UpdatedAt" || name == "CreatedAt" {
			return true
		}

		for _, s := range omit {
			if name == s {
				return true
			}
		}

		return false
	}

	result = make(map[string]interface{})

	elem := reflect.ValueOf(m).Elem()
	relType := elem.Type()
	num := relType.NumField()

	result = make(map[string]interface{}, num)

	// Add exported fields to result.
	for i := 0; i < num; i++ {
		n := relType.Field(i).Name
		v := elem.Field(i)

		if !v.CanSet() {
			continue
		} else if skip(n) {
			continue
		}

		result[n] = elem.Field(i).Interface()
	}

	return result
}
