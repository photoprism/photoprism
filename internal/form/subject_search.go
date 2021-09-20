package form

// SubjectSearch represents search form fields for "/api/v1/subjects".
type SubjectSearch struct {
	Query    string `form:"q"`
	ID       string `form:"id"`
	Type     string `form:"type"`
	Name     string `form:"name"`
	Hidden   bool   `form:"hidden"`
	Favorite bool   `form:"favorite"`
	Private  bool   `form:"private"`
	Excluded bool   `form:"excluded"`
	Files    int    `form:"files"`
	Count    int    `form:"count" binding:"required" serialize:"-"`
	Offset   int    `form:"offset" serialize:"-"`
	Order    string `form:"order" serialize:"-"`
}

func (f *SubjectSearch) GetQuery() string {
	return f.Query
}

func (f *SubjectSearch) SetQuery(q string) {
	f.Query = q
}

func (f *SubjectSearch) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewSubjectSearch(query string) SubjectSearch {
	return SubjectSearch{Query: query}
}
