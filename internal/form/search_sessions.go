package form

// SearchSessions represents a session search form.
type SearchSessions struct {
	Query  string `form:"q"`
	UID    string `form:"uid"`
	Count  int    `form:"count" binding:"required" serialize:"-"`
	Offset int    `form:"offset" serialize:"-"`
	Order  string `form:"order" serialize:"-"`
}

func (f *SearchSessions) GetQuery() string {
	return f.Query
}

func (f *SearchSessions) SetQuery(q string) {
	f.Query = q
}

func (f *SearchSessions) ParseQueryString() error {
	return ParseQueryString(f)
}
