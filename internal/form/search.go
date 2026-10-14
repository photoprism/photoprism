package form

import (
	"reflect"

	"github.com/photoprism/photoprism/pkg/clean"
)

// SearchForm defines the minimal interface for query string parsing.
type SearchForm interface {
	GetQuery() string
	SetQuery(q string)
}

// ParseQueryString removes NUL bytes and invalid UTF-8 from the string fields of the search form, and
// populates its fields from its query string.
func ParseQueryString(f SearchForm) (result error) {
	bindableStrings(f)

	q := f.GetQuery()

	return Unserialize(f, q)
}

// bindableStrings applies clean.SearchBindable to every settable string field of a pointer to a struct.
func bindableStrings(f any) {
	v := reflect.ValueOf(f)

	if v.Kind() != reflect.Pointer || v.IsNil() {
		return
	} else if v = v.Elem(); v.Kind() != reflect.Struct {
		return
	}

	for i := range v.NumField() {
		if field := v.Field(i); field.Kind() == reflect.String && field.CanSet() {
			if s := field.String(); s != "" {
				field.SetString(clean.SearchBindable(s))
			}
		}
	}
}
