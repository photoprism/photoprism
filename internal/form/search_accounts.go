package form

// SearchAccounts represents search form fields for "/api/v1/accounts".
type SearchAccounts struct {
	Query  string `form:"q"`
	Share  bool   `form:"share"`
	Sync   bool   `form:"sync"`
	Status string `form:"status"`
	Count  int    `form:"count" binding:"required" serialize:"-"`
	Offset int    `form:"offset" serialize:"-"`
	Order  string `form:"order" serialize:"-"`
}

func (f *SearchAccounts) GetQuery() string {
	return f.Query
}

func (f *SearchAccounts) SetQuery(q string) {
	f.Query = q
}

func (f *SearchAccounts) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewAccountSearch(query string) SearchAccounts {
	return SearchAccounts{Query: query}
}
