package form

// AccountSearch represents search form fields for "/api/v1/accounts".
type AccountSearch struct {
	Query  string `form:"q"`
	Share  bool   `form:"share"`
	Sync   bool   `form:"sync"`
	Status string `form:"status"`
	Count  int    `form:"count" binding:"required" serialize:"-"`
	Offset int    `form:"offset" serialize:"-"`
	Order  string `form:"order" serialize:"-"`
}

func (f *AccountSearch) GetQuery() string {
	return f.Query
}

func (f *AccountSearch) SetQuery(q string) {
	f.Query = q
}

func (f *AccountSearch) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewAccountSearch(query string) AccountSearch {
	return AccountSearch{Query: query}
}
