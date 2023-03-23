package form

import "github.com/ulule/deepcopier"

// Subject represents an image subject edit form.
type Subject struct {
	SubjName     string `json:"Name"`
	SubjAlias    string `json:"Alias"`
	SubjAbout    string `json:"About"`
	SubjBio      string `json:"Bio"`
	SubjNotes    string `json:"Notes"`
	SubjFavorite bool   `json:"Favorite"`
	SubjHidden   bool   `json:"Hidden"`
	SubjPrivate  bool   `json:"Private"`
	SubjExcluded bool   `json:"Excluded"`
}

func NewSubject(m interface{}) (f Subject, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
