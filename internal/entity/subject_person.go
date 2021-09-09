package entity

import (
	"encoding/json"

	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	SubjectPerson = "person"
)

// People represents a list of people.
type People []Person

// Person represents a subject with type person.
type Person struct {
	SubjectUID   string `json:"UID"`
	SubjectName  string `json:"Name"`
	SubjectAlias string `json:"Alias,omitempty"`
	Favorite     bool   `json:"Favorite,omitempty"`
	Thumb        string `json:",omitempty"`
}

// NewPerson returns a new entity.
func NewPerson(subj Subject) *Person {
	result := &Person{
		SubjectUID:   subj.SubjectUID,
		SubjectName:  subj.SubjectName,
		SubjectAlias: subj.SubjectAlias,
		Favorite:     subj.Favorite,
		Thumb:        subj.Thumb,
	}

	return result
}

// MarshalJSON returns the JSON encoding.
func (m *Person) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UID      string
		Name     string
		Keywords []string `json:",omitempty"`
		Favorite bool     `json:",omitempty"`
		Thumb    string   `json:",omitempty"`
	}{
		UID:      m.SubjectUID,
		Name:     m.SubjectName,
		Keywords: txt.NameKeywords(m.SubjectName, m.SubjectAlias),
		Favorite: m.Favorite,
		Thumb:    m.Thumb,
	})
}
