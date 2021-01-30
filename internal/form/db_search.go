package form

import "time"

// DbSearch represents search form fields for "/api/v1/db".
type DbSearch struct {
	Query   string    `form:"q"`
	Table   string    `form:"table" binding:"required"`
	Deleted bool      `form:"deleted"`
	Since   time.Time `form:"since"`
	Count   uint16    `form:"count" binding:"required"`
}

func (f *DbSearch) GetQuery() string {
	return f.Query
}

func (f *DbSearch) SetQuery(q string) {
	f.Query = q
}

func (f *DbSearch) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewDbSearch(query string) DbSearch {
	return DbSearch{Query: query}
}
