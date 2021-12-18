package form

// SearchLabels represents search form fields for "/api/v1/labels".
type SearchLabels struct {
	Query    string `form:"q"`
	UID      string `form:"uid"`
	Slug     string `form:"slug"`
	Name     string `form:"name"`
	All      bool   `form:"all"`
	Favorite bool   `form:"favorite"`
	Count    int    `form:"count" binding:"required" serialize:"-"`
	Offset   int    `form:"offset" serialize:"-"`
	Order    string `form:"order" serialize:"-"`
}

func (f *SearchLabels) GetQuery() string {
	return f.Query
}

func (f *SearchLabels) SetQuery(q string) {
	f.Query = q
}

func (f *SearchLabels) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewLabelSearch(query string) SearchLabels {
	return SearchLabels{Query: query}
}
