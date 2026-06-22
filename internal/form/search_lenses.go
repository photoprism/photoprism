package form

// SearchLenses represents search form fields for command or "/api/v1/lens" (future).
type SearchLenses struct {
	Query   string `form:"q"`
	ID      string `form:"id"`
	Slug    string `form:"slug"`
	Name    string `form:"name"`
	NoMake  bool   `form:"nomake"`
	Count   int    `form:"count" binding:"required" serialize:"-"`
	Offset  int    `form:"offset" serialize:"-"`
	Reverse bool   `form:"reverse" serialize:"-"`
}

// GetQuery returns the current search query string.
func (f *SearchLenses) GetQuery() string {
	return f.Query
}

// SetQuery stores the raw query string.
func (f *SearchLenses) SetQuery(q string) {
	f.Query = q
}

// ParseQueryString deserializes the query string into form fields.
func (f *SearchLenses) ParseQueryString() error {
	return ParseQueryString(f)
}

// NewLensSearch creates a SearchLens form with the provided query.
func NewLensSearch(query string) SearchLenses {
	return SearchLenses{Query: query}
}
