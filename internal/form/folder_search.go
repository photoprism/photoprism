package form

// FolderSearch represents search form fields for "/api/v1/folders".
type FolderSearch struct {
	Query     string `form:"q"`
	Recursive bool   `form:"recursive"`
	Files     bool   `form:"files"`
	Uncached  bool   `form:"uncached"`
	Count     int    `form:"count" serialize:"-"`
	Offset    int    `form:"offset" serialize:"-"`
}

func (f *FolderSearch) GetQuery() string {
	return f.Query
}

func (f *FolderSearch) SetQuery(q string) {
	f.Query = q
}

func (f *FolderSearch) ParseQueryString() error {
	return ParseQueryString(f)
}

// Serialize returns a string containing non-empty fields and values of a struct.
func (f *FolderSearch) Serialize() string {
	return Serialize(f, false)
}

// SerializeAll returns a string containing all non-empty fields and values of a struct.
func (f *FolderSearch) SerializeAll() string {
	return Serialize(f, true)
}
