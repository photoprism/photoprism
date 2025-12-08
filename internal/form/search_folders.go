package form

// SearchFolders represents search form fields for "/api/v1/folders".
type SearchFolders struct {
	Query     string `form:"q"`
	Recursive bool   `form:"recursive"`
	Files     bool   `form:"files"`
	Uncached  bool   `form:"uncached"`
	Public    bool   `form:"public"`
	Count     int    `form:"count" serialize:"-"`
	Offset    int    `form:"offset" serialize:"-"`
}

// GetQuery returns the current search query string.
func (f *SearchFolders) GetQuery() string {
	return f.Query
}

// SetQuery stores the raw query string.
func (f *SearchFolders) SetQuery(q string) {
	f.Query = q
}

// ParseQueryString deserializes the query string into form fields.
func (f *SearchFolders) ParseQueryString() error {
	return ParseQueryString(f)
}

// Serialize returns a string containing non-empty fields and values of a struct.
func (f *SearchFolders) Serialize() string {
	return Serialize(f, false)
}

// SerializeAll returns a string containing all non-empty fields and values of a struct.
func (f *SearchFolders) SerializeAll() string {
	return Serialize(f, true)
}
