package form

// SearchUsers represents a user search form.
type SearchUsers struct {
	Query   string `form:"q"`
	User    string `form:"user"`
	Name    string `form:"name"`
	Email   string `form:"email"`
	All     bool   `form:"all"`
	Deleted bool   `form:"deleted"`
	Count   int    `form:"count" binding:"required" serialize:"-"`
	Offset  int    `form:"offset" serialize:"-"`
	Order   string `form:"order" serialize:"-"`
	Reverse bool   `form:"reverse" serialize:"-"`
}

// GetQuery returns the current search query string.
func (f *SearchUsers) GetQuery() string {
	return f.Query
}

// SetQuery stores the raw query string.
func (f *SearchUsers) SetQuery(q string) {
	f.Query = q
}

// ParseQueryString deserializes the query string into form fields.
func (f *SearchUsers) ParseQueryString() error {
	return ParseQueryString(f)
}
