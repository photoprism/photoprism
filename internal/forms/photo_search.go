package forms

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	log "github.com/sirupsen/logrus"
)

// Query parameters for GET /api/v1/photos
type PhotoSearchForm struct {
	Query     string    `form:"q"`
	Location  bool      `form:"location"`
	Tags      string    `form:"tags"`
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
				stringValue := string(bytes.ToLower(value))

				if field.CanSet() {
					switch field.Interface().(type) {
					case time.Time:
						if timeValue, err := time.Parse("2006-01-02", stringValue); err != nil {
							result = err
						} else {
							field.Set(reflect.ValueOf(timeValue))
						}
					case int, int64:
						if i, err := strconv.Atoi(stringValue); err != nil {
							result = err
						} else {
							field.SetInt(int64(i))
						}
					case uint, uint64:
						if i, err := strconv.Atoi(stringValue); err != nil {
							result = err
						} else {
							field.SetUint(uint64(i))
						}
					case string:
						field.SetString(stringValue)
					case bool:
						if stringValue == "1" || stringValue == "true" || stringValue == "yes" {
							field.SetBool(true)
						} else if stringValue == "0" || stringValue == "false" || stringValue == "no" {
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

	if result != nil {
		log.Errorf("error while parsing search form: %s", result)
	}

	return result
}
