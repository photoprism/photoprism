package forms

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Query parameters for GET /api/v1/photos
type PhotoSearchForm struct {
	Query     string    `form:"q"`
	Location  bool      `form:"location"`
	Tags      string    `form:"tags"`
	Cat       string    `form:"cat"`
	Country   string    `form:"country"`
	Color     string    `form:"color"`
	Camera    int       `form:"camera"`
	Order     string    `form:"order"`
	Count     int       `form:"count" binding:"required"`
	Offset    int       `form:"offset"`
	Before    time.Time `form:"before" time_format:"2006-01-02"`
	After     time.Time `form:"after" time_format:"2006-01-02"`
	Favorites bool      `form:"favorites"`
}

func (f *PhotoSearchForm) ParseQueryString() (result error) {
	var key, value []byte
	var escaped, isKeyValue bool

	query := f.Query

	f.Query = ""

	formValues := reflect.ValueOf(f).Elem()

	query = strings.TrimSpace(query) + "\n"

	for _, char := range query {
		if unicode.IsSpace(char) && !escaped {
			if isKeyValue {
				fieldName := string(bytes.Title(bytes.ToLower(key)))

				field := formValues.FieldByName(fieldName)
				valueString := string(bytes.ToLower(value))

				if field.CanSet() {
					switch field.Interface().(type) {
					case int, int64:
						if i, err := strconv.Atoi(valueString); err == nil {
							field.SetInt(int64(i))
						} else {
							result = err
						}
					case uint, uint64:
						if i, err := strconv.Atoi(valueString); err == nil {
							field.SetUint(uint64(i))
						} else {
							result = err
						}
					case string:
						field.SetString(valueString)
					case bool:
						if valueString == "1" || valueString == "true" || valueString == "yes" {
							field.SetBool(true)
						} else if valueString == "0" || valueString == "false" || valueString == "no" {
							field.SetBool(false)
						} else {
							result = fmt.Errorf("not a bool value: %s", fieldName)
						}
					default:
						result = fmt.Errorf("unsupported field type: %s", fieldName)
					}
				} else {
					result = fmt.Errorf("unknown form field: %s", fieldName)
				}
			} else {
				f.Query = string(bytes.ToLower(key))
			}

			escaped = false
			isKeyValue = false
			key = key[:0]
			value = value[:0]
		} else if char == ':' {
			isKeyValue = true
		} else if char == '"' {
			escaped = !escaped
		} else if isKeyValue {
			value = append(value, byte(char))
		} else {
			key = append(key, byte(char))
		}
	}

	return result
}
