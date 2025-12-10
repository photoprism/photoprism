package form

// SearchForm defines the minimal interface for query string parsing.
type SearchForm interface {
	GetQuery() string
	SetQuery(q string)
}

// ParseQueryString populates the search form fields from its query string.
func ParseQueryString(f SearchForm) (result error) {
	q := f.GetQuery()

	return Unserialize(f, q)
}
