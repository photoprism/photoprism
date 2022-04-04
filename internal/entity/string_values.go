package entity

import (
	"reflect"
)

// Values is a shortcut for map[string]interface{}
type Values map[string]interface{}

// GetValues extracts entity Values.
func GetValues(m interface{}, omit ...string) (result Values) {
	skip := func(name string) bool {
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

	for i := 0; i < relType.NumField(); i++ {
		name := relType.Field(i).Name

		if skip(name) {
			continue
		}

		result[name] = elem.Field(i).Interface()
	}

	return result
}
