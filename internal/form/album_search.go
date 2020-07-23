package form

// AlbumSearch represents search form fields for "/api/v1/albums".
type AlbumSearch struct {
	Query    string `form:"q"`
	ID       string `form:"id"`
	Type     string `form:"type"`
	Location string `form:"location"`
	Category string `form:"category"`
	Slug     string `form:"slug"`
	Title    string `form:"title"`
	Country  string `json:"country"`
	Year     int    `json:"year"`
	Month    int    `json:"month"`
	Day      int    `json:"day"`
	Favorite bool   `form:"favorite"`
	Private  bool   `form:"private"`
	Count    int    `form:"count" binding:"required" serialize:"-"`
	Offset   int    `form:"offset" serialize:"-"`
	Order    string `form:"order" serialize:"-"`
}

func (f *AlbumSearch) GetQuery() string {
	return f.Query
}

func (f *AlbumSearch) SetQuery(q string) {
	f.Query = q
}

func (f *AlbumSearch) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewAlbumSearch(query string) AlbumSearch {
	return AlbumSearch{Query: query}
}
