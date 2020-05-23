package form

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Serialize returns a string containing all non-empty fields and values of a struct.
func Serialize(f interface{}, all bool) string {
	v := reflect.ValueOf(f)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return ""
	}

	q := make([]string, 0, v.NumField())

	// Iterate through all form fields.
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		fieldName := v.Type().Field(i).Tag.Get("form")
		fieldInfo := v.Type().Field(i).Tag.Get("serialize")

		// Serialize field values as string.
		if fieldName != "" && (fieldInfo != "-" || all) {
			switch t := fieldValue.Interface().(type) {
			case time.Time:
				if val := fieldValue.Interface().(time.Time); !val.IsZero() {
					if val.Hour() == 0 && val.Minute() == 0 {
						q = append(q, fmt.Sprintf("%s:%s", fieldName, val.Format("2006-01-02")))
					} else {
						q = append(q, fmt.Sprintf("%s:\"%s\"", fieldName, val.String()))
					}
				}
			case int, int8, int16, int32, int64:
				if val := fieldValue.Int(); val != 0 {
					q = append(q, fmt.Sprintf("%s:%d", fieldName, val))
				}
			case uint, uint8, uint16, uint32, uint64:
				if val := fieldValue.Uint(); val != 0 {
					q = append(q, fmt.Sprintf("%s:%d", fieldName, val))
				}
			case float32, float64:
				if val := fieldValue.Float(); val != 0 {
					q = append(q, fmt.Sprintf("%s:%f", fieldName, val))
				}
			case string:
				if val := strings.TrimSpace(strings.ReplaceAll(fieldValue.String(), "\"", "")); val != "" {
					if strings.Contains(val, " ") {
						q = append(q, fmt.Sprintf("%s:\"%s\"", fieldName, val))
					} else {
						q = append(q, fmt.Sprintf("%s:%s", fieldName, val))
					}
				}
			case bool:
				if val := fieldValue.Bool(); val {
					q = append(q, fmt.Sprintf("%s:%t", fieldName, fieldValue.Bool()))
				}
			default:
				log.Warnf("can't serialize value of type %s from form field %s", t, fieldName)
			}
		}
	}

	return strings.Join(q, " ")
}
