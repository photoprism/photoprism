package form

// SearchAlbums represents search form fields for "/api/v1/albums".
type SearchAlbums struct {
	Query    string `form:"q"`
	UID      string `form:"uid"`
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

func (f *SearchAlbums) GetQuery() string {
	return f.Query
}

func (f *SearchAlbums) SetQuery(q string) {
	f.Query = q
}

func (f *SearchAlbums) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewAlbumSearch(query string) SearchAlbums {
	return SearchAlbums{Query: query}
}
