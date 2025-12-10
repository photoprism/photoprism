package form

// SearchSubjects represents search form fields for "/api/v1/subjects".
type SearchSubjects struct {
	Query    string `form:"q"`
	UID      string `form:"uid"`
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
	Reverse  bool   `form:"reverse" serialize:"-"`
}

// GetQuery returns the current search query string.
func (f *SearchSubjects) GetQuery() string {
	return f.Query
}

// SetQuery stores the raw query string.
func (f *SearchSubjects) SetQuery(q string) {
	f.Query = q
}

// ParseQueryString deserializes the query string into form fields.
func (f *SearchSubjects) ParseQueryString() error {
	return ParseQueryString(f)
}

// NewSubjectSearch creates a SearchSubjects form with the provided query.
func NewSubjectSearch(query string) SearchSubjects {
	return SearchSubjects{Query: query}
}
