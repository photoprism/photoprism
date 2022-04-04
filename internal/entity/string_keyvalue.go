package entity

// KeyValue represents a string key/value pair.
type KeyValue struct {
	K string `json:"value"`
	V string `json:"text"`
}

// KeyValues represents a list of string key/value pairs.
type KeyValues []KeyValue

// Strings returns the list as a lookup map.
func (v KeyValues) Strings() Strings {
	result := make(Strings, len(v))

	for i := range v {
		result[v[i].K] = v[i].V
	}

	return result
}
