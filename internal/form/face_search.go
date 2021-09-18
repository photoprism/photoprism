package form

// FaceSearch represents search form fields for "/api/v1/faces".
type FaceSearch struct {
	Query   string `form:"q"`
	ID      string `form:"id"`
	Subject string `form:"subject"`
	Unknown string `form:"unknown"`
	Hidden  string `form:"hidden"`
	Markers bool   `form:"markers"`
	Count   int    `form:"count" binding:"required" serialize:"-"`
	Offset  int    `form:"offset" serialize:"-"`
	Order   string `form:"order" serialize:"-"`
}

func (f *FaceSearch) GetQuery() string {
	return f.Query
}

func (f *FaceSearch) SetQuery(q string) {
	f.Query = q
}

func (f *FaceSearch) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewFaceSearch(query string) FaceSearch {
	return FaceSearch{Query: query}
}
