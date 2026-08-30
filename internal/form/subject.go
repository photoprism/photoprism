package form

import (
	"time"

	"github.com/ulule/deepcopier"
)

// Subject represents an image subject edit form.
type Subject struct {
	SubjName     string     `json:"Name"`
	SubjAlias    string     `json:"Alias"`
	SubjBirthday *time.Time `json:"Birthday"`
	SubjAbout    string     `json:"About"`
	SubjBio      string     `json:"Bio"`
	SubjNotes    string     `json:"Notes"`
	SubjFavorite bool       `json:"Favorite"`
	SubjHidden   bool       `json:"Hidden"`
	SubjPrivate  bool       `json:"Private"`
	SubjExcluded bool       `json:"Excluded"`
	Verified     bool       `json:"Verified"`
	Thumb        string     `json:"Thumb"`
	ThumbSrc     string     `json:"ThumbSrc"`
}

// NewSubject copies values from an arbitrary model into a Subject form.
func NewSubject(m any) (*Subject, error) {
	frm := &Subject{}
	err := deepcopier.Copy(m).To(frm)

	// The form owns its own copy of each pointer value, so binding a request writes only into the
	// form and never back through a shared pointer into the model it was built from.
	if frm.SubjBirthday != nil {
		born := *frm.SubjBirthday
		frm.SubjBirthday = &born
	}

	return frm, err
}
