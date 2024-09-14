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
	Year     string `form:"year" example:"year:1990|2003" notes:"Year (separate with |)"`
	Month    string `form:"month" example:"month:7|10" notes:"Month (1-12, separate with |)"`
	Day      string `form:"day" example:"day:3|13" notes:"Day of Month (1-31, separate with |)"`
	Favorite bool   `form:"favorite"`
	Public   bool   `form:"public"`
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
