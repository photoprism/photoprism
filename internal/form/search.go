package form

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/photoprism/photoprism/pkg/txt"
	log "github.com/sirupsen/logrus"

	"github.com/araddon/dateparse"
)

type SearchForm interface {
	GetQuery() string
	SetQuery(q string)
}

func ParseQueryString(f SearchForm) (result error) {
	var key, value []rune
	var escaped, isKeyValue bool

	q := f.GetQuery()

	f.SetQuery("")

	formValues := reflect.ValueOf(f).Elem()

	q = strings.TrimSpace(q) + "\n"

	for _, char := range q {
		if unicode.IsSpace(char) && !escaped {
			if isKeyValue {
				fieldName := strings.Title(string(key))
				field := formValues.FieldByName(fieldName)
				stringValue := string(value)

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
						field.SetBool(txt.Bool(stringValue))
					default:
						result = fmt.Errorf("unsupported type: %s", fieldName)
					}
				} else {
					result = fmt.Errorf("unknown filter: %s", fieldName)
				}
			} else {
				f.SetQuery(string(key))
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
			value = append(value, unicode.ToLower(char))
		} else {
			key = append(key, unicode.ToLower(char))
		}
	}

	if result != nil {
		log.Errorf("error while parsing search form: %s", result)
	}

	return result
}
