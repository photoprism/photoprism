package config

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// optionField describes an option in Options as the values map sees it.
type optionField struct {
	Type reflect.Type
	// Exposed reports whether the API returns this option, which fields tagged json:"-" are not.
	Exposed bool
}

// optionFields returns the options in Options, indexed by the name they are stored under in
// "options.yml". Inline structs are flattened, because that is how they are stored and patched.
var optionFields = sync.OnceValue(func() map[string]optionField {
	fields := make(map[string]optionField)
	addOptionFields(fields, reflect.TypeFor[Options]())
	return fields
})

// addOptionFields indexes the persisted fields of a struct type by their stored name.
func addOptionFields(fields map[string]optionField, t reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name, opts, _ := strings.Cut(field.Tag.Get("yaml"), ",")

		// An inline struct is stored as if its fields were declared here.
		if strings.Contains(opts, "inline") && field.Type.Kind() == reflect.Struct {
			addOptionFields(fields, field.Type)
			continue
		}

		// The yaml tag is the name a value is stored under, so it is what a patch key matches.
		if name == "" || name == "-" {
			continue
		}

		jsonName, _, _ := strings.Cut(field.Tag.Get("json"), ",")

		fields[name] = optionField{Type: field.Type, Exposed: jsonName != "-"}
	}
}

// RemoveUnsupportedOptionValues removes the values a request may not set - names that are not
// options, and options the API does not return - and reports the names it removed, so that what a
// request may set matches what it may read.
func RemoveUnsupportedOptionValues(values Values) (removed []string) {
	fields := optionFields()

	for name := range values {
		if field, known := fields[name]; !known || !field.Exposed {
			delete(values, name)
			removed = append(removed, name)
		}
	}

	sort.Strings(removed)

	return removed
}

// CoerceOptionValues converts the numbers in an options patch to the type of the option they set,
// modifying the map in place and naming the option in any error it returns.
// JSON decodes every number into a float64 when the target is an untyped map, so an integer option
// would otherwise persist as a float that the loader silently truncates on the way back in.
func CoerceOptionValues(values Values) error {
	fields := optionFields()

	for name, value := range values {
		field, known := fields[name]

		// A name that is not an option has no type to check it against.
		if !known {
			continue
		}

		coerced, err := coerceOptionValue(name, field.Type, value)

		if err != nil {
			return err
		}

		values[name] = coerced
	}

	return nil
}

// coerceOptionValue converts a single option value to the specified type.
// Values that are not numbers pass through unchanged, since a string is how a duration is written.
func coerceOptionValue(name string, t reflect.Type, value any) (any, error) {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return coerceOptionInt(name, t, value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return coerceOptionUint(name, t, value)
	default:
		return value, nil
	}
}

// coerceOptionInt converts a numeric option value to a signed integer that fits the specified type.
func coerceOptionInt(name string, t reflect.Type, value any) (any, error) {
	var n int64

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n = v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u := v.Uint(); u > math.MaxInt64 {
			return nil, fmt.Errorf("%w: %s is out of range", ErrInvalidOptionValue, name)
		} else {
			n = int64(u)
		}
	case reflect.Float32, reflect.Float64:
		f, err := roundOptionFloat(name, v.Float())

		if err != nil {
			return nil, err
		}

		// float64(math.MaxInt64) rounds up to 2^63, so this is the exact overflow bound.
		if f >= float64(math.MaxInt64) || f < float64(math.MinInt64) {
			return nil, fmt.Errorf("%w: %s is out of range", ErrInvalidOptionValue, name)
		}

		n = int64(f)
	default:
		return value, nil
	}

	if reflect.Zero(t).OverflowInt(n) {
		return nil, fmt.Errorf("%w: %s is out of range", ErrInvalidOptionValue, name)
	}

	return n, nil
}

// coerceOptionUint converts a numeric option value to an unsigned integer that fits the specified
// type, and rejects a negative number rather than wrapping it around.
func coerceOptionUint(name string, t reflect.Type, value any) (any, error) {
	var n uint64

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n = v.Uint()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i := v.Int(); i < 0 {
			return nil, fmt.Errorf("%w: %s must not be negative", ErrInvalidOptionValue, name)
		} else {
			n = uint64(i)
		}
	case reflect.Float32, reflect.Float64:
		f, err := roundOptionFloat(name, v.Float())

		if err != nil {
			return nil, err
		}

		if f < 0 {
			return nil, fmt.Errorf("%w: %s must not be negative", ErrInvalidOptionValue, name)
		} else if f >= float64(math.MaxUint64) {
			return nil, fmt.Errorf("%w: %s is out of range", ErrInvalidOptionValue, name)
		}

		n = uint64(f)
	default:
		return value, nil
	}

	if reflect.Zero(t).OverflowUint(n) {
		return nil, fmt.Errorf("%w: %s is out of range", ErrInvalidOptionValue, name)
	}

	return n, nil
}

// roundOptionFloat rounds a value to the nearest whole number so that an option set from a slider
// keeps the number the user selected, and rejects one that has no integer representation at all.
func roundOptionFloat(name string, f float64) (float64, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, fmt.Errorf("%w: %s is not a number", ErrInvalidOptionValue, name)
	}

	return math.Round(f), nil
}
