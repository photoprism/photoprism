package form

import "github.com/ulule/deepcopier"

// Album represents an album edit form.
type Album struct {
	CoverUID         string `json:"CoverUID"`
	ParentUID        string `json:"ParentUID"`
	FolderUID        string `json:"FolderUID"`
	AlbumType        string `json:"Type"`
	AlbumName        string `json:"Name"`
	AlbumDescription string `json:"Description"`
	AlbumNotes       string `json:"Notes"`
	AlbumFilter      string `json:"Filter"`
	AlbumOrder       string `json:"Order"`
	AlbumTemplate    string `json:"Template"`
	AlbumFavorite    bool   `json:"Favorite"`
}

func NewAlbum(m interface{}) (f Album, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
