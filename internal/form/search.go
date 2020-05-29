package form

type SearchForm interface {
	GetQuery() string
	SetQuery(q string)
}

func ParseQueryString(f SearchForm) (result error) {
	q := f.GetQuery()

	return Unserialize(f, q)
}
