package form

// SearchFaces represents search form fields for "/api/v1/faces".
type SearchFaces struct {
	Query   string `form:"q"`
	UID     string `form:"uid"`
	Subject string `form:"subject"`
	Unknown string `form:"unknown"`
	Hidden  string `form:"hidden"`
	Markers bool   `form:"markers"`
	Count   int    `form:"count" binding:"required" serialize:"-"`
	Offset  int    `form:"offset" serialize:"-"`
	Order   string `form:"order" serialize:"-"`
}

func (f *SearchFaces) GetQuery() string {
	return f.Query
}

func (f *SearchFaces) SetQuery(q string) {
	f.Query = q
}

func (f *SearchFaces) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewFaceSearch(query string) SearchFaces {
	return SearchFaces{Query: query}
}
