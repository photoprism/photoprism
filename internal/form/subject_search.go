package form

// SubjectSearch represents search form fields for "/api/v1/subjects".
type SubjectSearch struct {
	Query    string `form:"q"`
	ID       string `form:"id"`
	Type     string `form:"type"`
	Name     string `form:"name"`
	All      bool   `form:"all"`
	Hidden   string `form:"hidden"`
	Favorite string `form:"favorite"`
	Private  string `form:"private"`
	Excluded string `form:"excluded"`
	Files    int    `form:"files"`
	Photos   int    `form:"photos"`
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
