package list

import (
	"sort"
	"strings"
)

// Attr represents a list of key-value attributes.
type Attr []*KeyValue

// ParseAttr parses a string into a new Attr since and returns it.
func ParseAttr(s string) Attr {
	fields := strings.Fields(s)
	result := make(Attr, 0, len(fields))

	// Append an attribute for each field.
	for _, v := range fields {
		f := ParseKeyValue(v)
		if f != nil {
			result = append(result, f)
		}
	}

	return result
}

// String returns the attributes as string.
func (list Attr) String() string {
	result := make([]string, 0, len(list))

	list.Sort()

	var i int
	var l int

	for _, f := range list {
		s := f.String()

		if s == "" {
			continue
		} else if i == 0 {
			// Skip check.
		} else if result[i-1] == s {
			continue
		}

		l += len(s)

		if l > StringLengthLimit {
			break
		}

		result = append(result, s)

		i++
	}

	return strings.Join(result, " ")
}

// Sort sorts the attributes by key.
func (list Attr) Sort() {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Key == list[j].Key {
			return list[i].Value < list[j].Value
		} else {
			return list[i].Key < list[j].Key
		}
	})
}
