package form

// SearchServices represents search form fields for "/api/v1/services".
type SearchServices struct {
	Query  string `form:"q"`
	Share  bool   `form:"share"`
	Sync   bool   `form:"sync"`
	Status string `form:"status"`
	Count  int    `form:"count" binding:"required" serialize:"-"`
	Offset int    `form:"offset" serialize:"-"`
	Order  string `form:"order" serialize:"-"`
}

// GetQuery returns the current search query string.
func (f *SearchServices) GetQuery() string {
	return f.Query
}

// SetQuery stores the raw query string.
func (f *SearchServices) SetQuery(q string) {
	f.Query = q
}

// ParseQueryString deserializes the query string into form fields.
func (f *SearchServices) ParseQueryString() error {
	return ParseQueryString(f)
}

// NewSearchServices creates a SearchServices form with the provided query.
func NewSearchServices(query string) SearchServices {
	return SearchServices{Query: query}
}
