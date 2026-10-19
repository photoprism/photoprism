package form

import (
	"errors"
	"reflect"

	"github.com/photoprism/photoprism/pkg/clean"
)

// SearchForm defines the minimal interface for query string parsing.
type SearchForm interface {
	GetQuery() string
	SetQuery(q string)
}

// ErrNoMatch is returned when removing NUL bytes and invalid UTF-8 from a search filter value leaves nothing
// usable, so it can match nothing; callers return an empty result.
var ErrNoMatch = errors.New("search filter matches nothing")

// ParseQueryString removes NUL bytes and invalid UTF-8 from the string fields of the search form, and
// populates its fields from its query string. It returns ErrNoMatch if a filter value had nothing usable left.
func ParseQueryString(f SearchForm) (result error) {
	q := f.GetQuery()
	ok := bindableStrings(f)

	if noMatch, err := unserialize(f, q); err != nil {
		return err
	} else if noMatch || !ok {
		return ErrNoMatch
	}

	return nil
}

// bindableStrings applies clean.SearchBindable to every settable string field of a pointer to a struct, and
// reports false if a field other than the sort order had nothing usable left.
func bindableStrings(f any) (ok bool) {
	ok = true
	v := reflect.ValueOf(f)

	if v.Kind() != reflect.Pointer || v.IsNil() {
		return ok
	} else if v = v.Elem(); v.Kind() != reflect.Struct {
		return ok
	}

	for i := range v.NumField() {
		if field := v.Field(i); field.Kind() == reflect.String && field.CanSet() {
			if s, valid := clean.SearchFilter(field.String()); !valid {
				field.SetString(s)

				if v.Type().Field(i).Tag.Get("form") != "order" {
					ok = false
				}
			} else if s != field.String() {
				field.SetString(s)
			}
		}
	}

	return ok
}
