package form

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Serialize returns a string containing all non-empty fields and values of a struct.
func Serialize(f any, all bool) string {
	v := reflect.ValueOf(f)

	if v.Kind() == reflect.Pointer {
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
		if fieldName != "" && fieldName != "-" && (fieldInfo != "-" || all) {
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
				if val := strings.ReplaceAll(fieldValue.String(), "\"", ""); val != "" {
					if strings.ContainsAny(val, " :'()[]-+`") {
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
				log.Warnf("form: failed serializing %T %s", t, clean.Token(fieldName))
			}
		}
	}

	return strings.Join(q, " ")
}

// Unserialize parses a query string into the provided SearchForm implementation.
func Unserialize(f SearchForm, q string) (result error) {
	_, result = unserialize(f, q)
	return result
}

// UnserializeFilter parses a stored filter like Unserialize, and reports whether removing NUL bytes and invalid
// UTF-8 left nothing usable of the filter or of one of its values, so it can match nothing.
func UnserializeFilter(f SearchForm, q string) (noMatch bool, err error) {
	_, usable := clean.SearchFilter(q)
	noMatch, err = unserialize(f, q)

	return noMatch || !usable, err
}

// unserialize parses a query string into the provided SearchForm, skipping NUL bytes and invalid UTF-8, and
// reports whether the value of a string filter consisted only of those, so it can match nothing.
func unserialize(f SearchForm, q string) (noMatch bool, result error) {
	var key, value []rune
	var escaped, isKeyValue, dropped bool

	v := reflect.ValueOf(f)

	formValues := v.Elem()

	n := formValues.NumField()

	fieldNames := make(map[string]string, n)

	// Iterate through all form fields.
	for i := 0; i < formValues.NumField(); i++ {
		fieldName := formValues.Type().Field(i).Name
		formName := strings.ToLower(formValues.Type().Field(i).Tag.Get("form"))
		formSerialize := strings.ToLower(formValues.Type().Field(i).Tag.Get("serialize"))

		if fieldName == "" || formSerialize == "-" {
			continue
		} else if formName == "" {
			formName = strings.ToLower(fieldName)
		}

		fieldNames[formName] = fieldName
	}

	f.SetQuery("")

	var queryStrings []string

	// flush sets the field named by the current token, or keeps the token as a query string.
	flush := func() {
		if isKeyValue {
			formName := strings.ToLower(string(key))
			fieldName := fieldNames[formName]

			field := formValues.FieldByNameFunc(func(name string) bool {
				return strings.EqualFold(name, fieldName)
			})

			stringValue := string(value)

			if fieldName != "" && fieldName != "-" && field.CanSet() {
				// A value whose removed bytes leave nothing usable matches nothing, whatever the field type.
				if dropped && !clean.SearchUsable(stringValue) {
					noMatch = true
				} else {
					switch field.Interface().(type) {
					case time.Time:
						if timeValue := txt.ParseTimeUTC(stringValue); stringValue != "" && timeValue.IsZero() {
							result = fmt.Errorf("invalid %s date", formName)
						} else {
							field.Set(reflect.ValueOf(timeValue))
						}
					case float32, float64:
						if floatValue, err := strconv.ParseFloat(stringValue, 64); err != nil {
							result = err
						} else {
							field.SetFloat(floatValue)
						}
					case int, int8, int16, int32, int64:
						if intValue, err := strconv.Atoi(stringValue); err != nil {
							result = err
						} else {
							field.SetInt(int64(intValue))
						}
					case uint, uint8, uint16, uint32, uint64:
						bitSize := field.Type().Bits()
						if uintValue, err := strconv.ParseUint(stringValue, 10, int(bitSize)); err != nil {
							result = err
						} else {
							field.SetUint(uintValue)
						}
					case string:
						searchValue := clean.SearchString(stringValue)
						noMatch = noMatch || dropped && !clean.SearchUsable(searchValue)
						field.SetString(searchValue)
					case bool:
						field.SetBool(txt.Bool(stringValue))
					default:
						result = fmt.Errorf("unsupported type: %s", formName)
					}
				}
			} else {
				result = fmt.Errorf("unknown filter: %s", formName)
			}
		} else if len(strings.TrimSpace(string(key))) > 0 {
			queryStrings = append(queryStrings, strings.TrimSpace(string(key)))
		}

		escaped = false
		isKeyValue = false
		dropped = false
		key = key[:0]
		value = value[:0]
	}

	q = strings.TrimSpace(q)

	for i, char := range q {
		// NUL bytes and invalid UTF-8 are skipped, and an invalid byte decodes as a one-byte RuneError.
		if char == 0 || char == utf8.RuneError && !strings.HasPrefix(q[i:], string(utf8.RuneError)) {
			dropped = dropped || isKeyValue
			continue
		}

		switch {
		case unicode.IsSpace(char) && !escaped:
			flush()
		case char == ':' && !escaped:
			isKeyValue = true
		case char == '"':
			escaped = !escaped
		case isKeyValue:
			value = append(value, char)
		default:
			key = append(key, unicode.ToLower(char))
		}
	}

	// The last token ends with the string, also if its quote is not closed.
	flush()

	if len(queryStrings) > 0 {
		f.SetQuery(clean.SearchQuery(strings.Join(queryStrings, " ")))
	}

	if result != nil {
		log.Warnf("form: failed parsing search query")
	}

	return noMatch, result
}
