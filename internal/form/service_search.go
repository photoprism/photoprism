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

func (f *SearchServices) GetQuery() string {
	return f.Query
}

func (f *SearchServices) SetQuery(q string) {
	f.Query = q
}

func (f *SearchServices) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewSearchServices(query string) SearchServices {
	return SearchServices{Query: query}
}
