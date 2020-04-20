package form

import "github.com/ulule/deepcopier"

// Album represents an album edit form.
type Album struct {
	AlbumName        string `json:"AlbumName"`
	AlbumDescription string `json:"AlbumDescription"`
	AlbumNotes       string `json:"AlbumNotes"`
	AlbumOrder       string `json:"AlbumOrder"`
	AlbumTemplate    string `json:"AlbumTemplate"`
	AlbumFavorite    bool   `json:"AlbumFavorite"`
}

func NewAlbum(m interface{}) (f Album, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
