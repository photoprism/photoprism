package form

// Album represents an album edit form.
type Album struct {
	AlbumName        string `json:"AlbumName"`
	AlbumDescription string `json:"AlbumDescription"`
	AlbumNotes       string `json:"AlbumNotes"`
	AlbumFavorite    bool   `json:"AlbumFavorite"`
	AlbumPublic      bool   `json:"AlbumPublic"`
	AlbumOrder       string `json:"AlbumOrder"`
	AlbumTemplate    string `json:"AlbumTemplate"`
}
