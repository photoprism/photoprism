package form

import "github.com/ulule/deepcopier"

// Folder represents a file system directory edit form.
type Folder struct {
	Path              string `json:"Path"`
	Root              string `json:"Root"`
	FolderType        string `json:"Type"`
	FolderTitle       string `json:"Title"`
	FolderCategory    string `json:"Category"`
	FolderDescription string `json:"Description"`
	FolderOrder       string `json:"Order"`
	FolderCountry     string `json:"Country"`
	FolderYear        int    `json:"Year"`
	FolderMonth       int    `json:"Month"`
	FolderFavorite    bool   `json:"Favorite"`
	FolderPrivate     bool   `json:"Private"`
	FolderIgnore      bool   `json:"Ignore"`
	FolderWatch       bool   `json:"Watch"`
}

// NewFolder copies values from an arbitrary model into a Folder form.
func NewFolder(m interface{}) (f Folder, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
