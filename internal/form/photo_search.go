package form

import (
	"time"
)

// PhotoSearch represents search form fields for "/api/v1/photos".
type PhotoSearch struct {
	Query       string    `form:"q"`
	ID          string    `form:"id"`
	Title       string    `form:"title"`
	Description string    `form:"description"`
	Notes       string    `form:"notes"`
	Artist      string    `form:"artist"`
	Hash        string    `form:"hash"`
	Duplicate   bool      `form:"duplicate"`
	Archived    bool      `form:"archived"`
	Error       bool      `form:"error"`
	Lat         float64   `form:"lat"`
	Lng         float64   `form:"lng"`
	Dist        uint      `form:"dist"`
	Fmin        float64   `form:"fmin"`
	Fmax        float64   `form:"fmax"`
	Chroma      uint      `form:"chroma"`
	Mono        bool      `form:"mono"`
	Portrait    bool      `form:"portrait"`
	Location    bool      `form:"location"`
	Album       string    `form:"album"`
	Label       string    `form:"label"`
	Country     string    `form:"country"`
	Year        uint      `form:"year"`
	Month       uint      `form:"month"`
	Color       string    `form:"color"`
	Camera      int       `form:"camera"`
	Lens        int       `form:"lens"`
	Before      time.Time `form:"before" time_format:"2006-01-02"`
	After       time.Time `form:"after" time_format:"2006-01-02"`
	Favorites   bool      `form:"favorites"`
	Public      bool      `form:"public"`
	Story       bool      `form:"story"`
	Safe        bool      `form:"safe"`
	Nsfw        bool      `form:"nsfw"`
	Count       int       `form:"count" binding:"required"`
	Offset      int       `form:"offset"`
	Order       string    `form:"order"`
}

func (f *PhotoSearch) GetQuery() string {
	return f.Query
}

func (f *PhotoSearch) SetQuery(q string) {
	f.Query = q
}

func (f *PhotoSearch) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewPhotoSearch(query string) PhotoSearch {
	return PhotoSearch{Query: query}
}
