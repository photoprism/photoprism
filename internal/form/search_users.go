package form

// SearchUsers represents a user search form.
type SearchUsers struct {
	User   string `form:"user"`
	Query  string `form:"q"`
	Name   string `form:"name"`
	Email  string `form:"email"`
	Count  int    `form:"count" binding:"required" serialize:"-"`
	Offset int    `form:"offset" serialize:"-"`
	Order  string `form:"order" serialize:"-"`
}

func (f *SearchUsers) GetQuery() string {
	return f.Query
}

func (f *SearchUsers) SetQuery(q string) {
	f.Query = q
}

func (f *SearchUsers) ParseQueryString() error {
	return ParseQueryString(f)
}
