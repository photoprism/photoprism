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

	// Detached from the model, because deepcopier copies the pointer rather than what it points at
	// and binding a request decodes into the existing pointee. Sharing it edits the subject in place,
	// which hides the edit from the comparison deciding whether anything is written.
	if frm.SubjBirthday != nil {
		born := *frm.SubjBirthday
		frm.SubjBirthday = &born
	}

	return frm, err
}
