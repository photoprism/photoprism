package form

import "time"

// GeoSearch represents search form fields for "/api/v1/geo".
type GeoSearch struct {
	Query    string    `form:"q"`
	Type     string    `form:"type"`
	Path     string    `form:"path"`
	Folder   string    `form:"folder"` // Alias for Path
	Name     string    `form:"name"`
	Before   time.Time `form:"before" time_format:"2006-01-02"`
	After    time.Time `form:"after" time_format:"2006-01-02"`
	Favorite bool      `form:"favorite"`
	Video    bool      `form:"video"`
	Photo    bool      `form:"photo"`
	Archived bool      `form:"archived"`
	Public   bool      `form:"public"`
	Private  bool      `form:"private"`
	Review   bool      `form:"review"`
	Quality  int       `form:"quality"`
	Lat      float32   `form:"lat"`
	Lng      float32   `form:"lng"`
	S2       string    `form:"s2"`
	Olc      string    `form:"olc"`
	Dist     uint      `form:"dist"`
	Album    string    `form:"album"`
	Country  string    `form:"country"`
	Year     int       `form:"year"`
	Month    int       `form:"month"`
	Color    string    `form:"color"`
	Camera   int       `form:"camera"`
	Lens     int       `form:"lens"`
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
	err := ParseQueryString(f)

	if f.Path == "" && f.Folder != "" {
		f.Path = f.Folder
	}

	return err
}

// Serialize returns a string containing non-empty fields and values of a struct.
func (f *GeoSearch) Serialize() string {
	return Serialize(f, false)
}

// SerializeAll returns a string containing all non-empty fields and values of a struct.
func (f *GeoSearch) SerializeAll() string {
	return Serialize(f, true)
}

func NewGeoSearch(query string) GeoSearch {
	return GeoSearch{Query: query}
}
