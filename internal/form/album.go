package form

import "github.com/ulule/deepcopier"

// Album represents an album edit form.
type Album struct {
	AlbumType        string `json:"Type"`
	AlbumTitle       string `json:"Title"`
	AlbumLocation    string `json:"Location"`
	AlbumCategory    string `json:"Category"`
	AlbumCaption     string `json:"Caption"`
	AlbumDescription string `json:"Description"`
	AlbumNotes       string `json:"Notes"`
	AlbumFilter      string `json:"Filter"`
	AlbumOrder       string `json:"Order"`
	AlbumTemplate    string `json:"Template"`
	AlbumCountry     string `json:"Country"`
	AlbumFavorite    bool   `json:"Favorite"`
	AlbumPrivate     bool   `json:"Private"`
	Thumb            string `json:"Thumb"`
	ThumbSrc         string `json:"ThumbSrc"`
}

// NewAlbum creates a new form struct based on the interface values.
func NewAlbum(m interface{}) (*Album, error) {
	frm := &Album{}
	err := deepcopier.Copy(m).To(frm)

	return frm, err
}
