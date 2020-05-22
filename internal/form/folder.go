package form

import "github.com/ulule/deepcopier"

// Folder represents a file system directory edit form.
type Folder struct {
	Root              string `json:"Root"`
	Path              string `json:"Path"`
	FolderType        string `json:"Type"`
	FolderOrder       string `json:"Order"`
	FolderTitle       string `json:"Title"`
	FolderDescription string `json:"Description"`
	FolderFavorite    bool   `json:"Favorite"`
	FolderIgnore      bool   `json:"Ignore"`
	FolderHidden      bool   `json:"Hidden"`
	FolderWatch       bool   `json:"Watch"`
}

func NewFolder(m interface{}) (f Folder, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
