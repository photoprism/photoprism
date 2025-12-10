package form

// SearchLabels represents search form fields for "/api/v1/labels".
type SearchLabels struct {
	Query    string `form:"q"`
	UID      string `form:"uid"`
	Slug     string `form:"slug"`
	Name     string `form:"name"`
	All      bool   `form:"all"`
	Favorite bool   `form:"favorite"`
	NSFW     bool   `form:"nsfw"`
	Public   bool   `form:"public"`
	Count    int    `form:"count" binding:"required" serialize:"-"`
	Offset   int    `form:"offset" serialize:"-"`
	Order    string `form:"order" serialize:"-"`
	Reverse  bool   `form:"reverse" serialize:"-"`
}

// GetQuery returns the current search query string.
func (f *SearchLabels) GetQuery() string {
	return f.Query
}

// SetQuery stores the raw query string.
func (f *SearchLabels) SetQuery(q string) {
	f.Query = q
}

// ParseQueryString deserializes the query string into form fields.
func (f *SearchLabels) ParseQueryString() error {
	return ParseQueryString(f)
}

// NewLabelSearch creates a SearchLabels form with the provided query.
func NewLabelSearch(query string) SearchLabels {
	return SearchLabels{Query: query}
}
