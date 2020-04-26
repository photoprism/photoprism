package form

import "time"

// GeoSearch represents search form fields for "/api/v1/geo".
type GeoSearch struct {
	Query    string    `form:"q"`
	Before   time.Time `form:"before" time_format:"2006-01-02"`
	After    time.Time `form:"after" time_format:"2006-01-02"`
	Favorite bool      `form:"favorite"`
	Lat      float32   `form:"lat"`
	Lng      float32   `form:"lng"`
	S2       string    `form:"s2"`
	Olc      string    `form:"olc"`
	Dist     uint      `form:"dist"`
	Quality  int       `form:"quality"`
	Review   bool      `form:"review"`
}

// GetQuery returns the query parameter as string.
func (f *GeoSearch) GetQuery() string {
	return f.Query
}

// SetQuery sets the query parameter.
func (f *GeoSearch) SetQuery(q string) {
	f.Query = q
}

// ParseQueryString parses the query parameter if possible.
func (f *GeoSearch) ParseQueryString() error {
	return ParseQueryString(f)
}

func NewGeoSearch(query string) GeoSearch {
	return GeoSearch{Query: query}
}
