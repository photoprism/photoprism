package entity

import (
	"encoding/json"

	"github.com/photoprism/photoprism/pkg/txt"
)

const (
	SubjPerson = "person" // SubjType for people.
)

// People represents a list of people.
type People []Person

// Person represents a subject with type person.
type Person struct {
	SubjUID      string `json:"UID"`
	SubjName     string `json:"Name"`
	SubjAlias    string `json:"Alias"`
	SubjFavorite bool   `json:"Favorite"`
	SubjHidden   bool   `json:"Hidden"`
}

// AfterFind is a hook that updates the name cache after querying.
func (m *Person) AfterFind() (err error) {
	SubjNames.Set(m.SubjUID, m.SubjName)
	return
}

// NewPerson returns a new entity.
func NewPerson(subj Subject) *Person {
	result := &Person{
		SubjUID:      subj.SubjUID,
		SubjName:     subj.SubjName,
		SubjAlias:    subj.SubjAlias,
		SubjFavorite: subj.SubjFavorite,
		SubjHidden:   subj.SubjHidden,
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
		Hidden   bool     `json:",omitempty"`
	}{
		UID:      m.SubjUID,
		Name:     m.SubjName,
		Keywords: txt.NameKeywords(m.SubjName, m.SubjAlias),
		Favorite: m.SubjFavorite,
		Hidden:   m.SubjHidden,
	})
}
