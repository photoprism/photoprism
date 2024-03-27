package list

import (
	"sort"
	"strings"
)

// Attr represents a list of key-value attributes.
type Attr []*KeyValue

// ParseAttr parses a string into a new Attr slice and returns it.
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

// Contains tests if the list contains the attribute provided as string.
func (list Attr) Contains(s string) bool {
	if len(list) == 0 || s == "" {
		return false
	} else if s == All {
		return true
	}

	attr := ParseKeyValue(s)

	// Abort if attribute is invalid.
	if attr.Key == "" {
		return false
	}

	// Find matches.
	if attr.Value == "" || attr.Value == All {
		for i := range list {
			if strings.EqualFold(attr.Key, list[i].Key) || list[i].Key == All {
				return true
			}
		}
	} else {
		for i := range list {
			if strings.EqualFold(attr.Key, list[i].Key) && (attr.Value == list[i].Value || list[i].Value == All) || list[i].Key == All {
				return true
			}
		}
	}

	return false
}
