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
	return strings.Join(list.Strings(), " ")
}

// Strings returns the attributes as string slice.
func (list Attr) Strings() []string {
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

	return result
}

// Sort sorts the attributes by key.
func (list Attr) Sort() Attr {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Key == list[j].Key {
			return list[i].Value < list[j].Value
		} else if list[i].Key == All {
			return false
		} else if list[j].Key == All {
			return true
		} else {
			return list[i].Key < list[j].Key
		}
	})

	return list
}

// Contains tests if the list contains the attribute provided as string.
func (list Attr) Contains(s string) bool {
	attr := list.Find(s)

	if attr.Key == "" || attr.Value == False {
		return false
	}

	return true
}

// Find returns the matching KeyValue attribute if found.
func (list Attr) Find(s string) (a KeyValue) {
	if len(list) == 0 || s == "" {
		return a
	} else if s == All {
		return KeyValue{Key: All, Value: ""}
	}

	attr := ParseKeyValue(s)

	// Return nil if key is invalid or all.
	if attr.Key == "" {
		return a
	}

	// Find and return first match.
	if attr.Value == "" || attr.Value == All {
		for i := range list {
			if strings.EqualFold(attr.Key, list[i].Key) {
				return *list[i]
			} else if list[i].Key == All {
				a = *list[i]
			}
		}
	} else {
		for i := range list {
			if strings.EqualFold(attr.Key, list[i].Key) {
				if attr.Value == True && list[i].Value == False {
					return KeyValue{Key: "", Value: ""}
				} else if attr.Value == list[i].Value {
					return *list[i]
				} else if list[i].Value == All {
					a = *list[i]
				}
			} else if list[i].Key == All && attr.Value != False {
				a = *list[i]
			}
		}
	}

	return a
}
