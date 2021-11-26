package form

// LabelSearch represents search form fields for "/api/v1/labels".
type LabelSearch struct {
	Query    string `form:"q"`
	ID       string `form:"id"`
	Slug     string `form:"slug"`
	Name     string `form:"name"`
	All      bool   `form:"all"`
	Favorite bool   `form:"favorite"`
	Count    int    `form:"count" binding:"required" serialize:"-"`
	Offset   int    `form:"offset" serialize:"-"`
	Order    string `form:"order" serialize:"-"`
}

func (f *LabelSearch) GetQuery() string {
	return f.Query
}

func (f *LabelSearch) SetQuery(q string) {
	f.Query = q
}

func (f *LabelSearch) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewLabelSearch(query string) LabelSearch {
	return LabelSearch{Query: query}
}
