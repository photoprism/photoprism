package form

// SearchCameras represents search form fields for command or "/api/v1/cameras".
type SearchCameras struct {
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
func (f *SearchCameras) GetQuery() string {
	return f.Query
}

// SetQuery stores the raw query string.
func (f *SearchCameras) SetQuery(q string) {
	f.Query = q
}

// ParseQueryString deserializes the query string into form fields.
func (f *SearchCameras) ParseQueryString() error {
	return ParseQueryString(f)
}

// NewCameraSearch creates a SearchCameras form with the provided query.
func NewCameraSearch(query string) SearchCameras {
	return SearchCameras{Query: query}
}
