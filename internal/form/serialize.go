package form

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
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

func Unserialize(f SearchForm, q string) (result error) {
	var key, value []rune
	var escaped, isKeyValue bool

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

	q = strings.TrimSpace(q) + "\n"

	var queryStrings []string

	for _, char := range q {
		if unicode.IsSpace(char) && !escaped {
			if isKeyValue {
				formName := strings.ToLower(string(key))
				fieldName := fieldNames[formName]

				field := formValues.FieldByNameFunc(func(name string) bool {
					return strings.EqualFold(name, fieldName)
				})

				stringValue := string(value)

				if fieldName != "" && fieldName != "-" && field.CanSet() {
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
						if intValue, err := strconv.Atoi(stringValue); err != nil {
							result = err
						} else {
							field.SetUint(uint64(intValue))
						}
					case string:
						field.SetString(clean.SearchString(stringValue))
					case bool:
						field.SetBool(txt.Bool(stringValue))
					default:
						result = fmt.Errorf("unsupported type: %s", formName)
					}
				} else {
					result = fmt.Errorf("unknown filter: %s", formName)
				}
			} else if len(strings.TrimSpace(string(key))) > 0 {
				queryStrings = append(queryStrings, strings.TrimSpace(string(key)))
			}

			escaped = false
			isKeyValue = false
			key = key[:0]
			value = value[:0]
		} else if char == ':' && !escaped {
			isKeyValue = true
		} else if char == '"' {
			escaped = !escaped
		} else if isKeyValue {
			value = append(value, char)
		} else {
			key = append(key, unicode.ToLower(char))
		}
	}

	if len(queryStrings) > 0 {
		f.SetQuery(clean.SearchQuery(strings.Join(queryStrings, " ")))
	}

	if result != nil {
		log.Warnf("form: failed parsing search query")
	}

	return result
}
