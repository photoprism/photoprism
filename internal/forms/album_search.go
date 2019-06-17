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

	"github.com/araddon/dateparse"
)

// Query parameters for GET /api/v1/albums
type AlbumSearchForm struct {
	Query string `form:"q"`

	Slug      string `form:"slug"`
	Name      string `form:"name"`
	Favorites bool   `form:"favorites"`

	Count  int    `form:"count" binding:"required"`
	Offset int    `form:"offset"`
	Order  string `form:"order"`
}

func (f *AlbumSearchForm) ParseQueryString() (result error) {
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
						if timeValue, err := dateparse.ParseAny(stringValue); err != nil {
							result = err
						} else {
							field.Set(reflect.ValueOf(timeValue))
						}
					case float64:
						if floatValue, err := strconv.ParseFloat(stringValue, 64); err != nil {
							result = err
						} else {
							field.SetFloat(floatValue)
						}
					case int, int64:
						if intValue, err := strconv.Atoi(stringValue); err != nil {
							result = err
						} else {
							field.SetInt(int64(intValue))
						}
					case uint, uint64:
						if intValue, err := strconv.Atoi(stringValue); err != nil {
							result = err
						} else {
							field.SetUint(uint64(intValue))
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
						result = fmt.Errorf("unsupported type: %s", fieldName)
					}
				} else {
					result = fmt.Errorf("unknown filter: %s", fieldName)
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
		log.Errorf("error while parsing album form: %s", result)
	}

	return result
}
