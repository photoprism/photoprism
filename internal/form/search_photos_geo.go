package form

import (
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
)

// SearchPhotosGeo represents search form fields for "/api/v1/geo".
type SearchPhotosGeo struct {
	Query     string    `form:"q"`
	Filter    string    `form:"filter"`
	Near      string    `form:"near"`
	Type      string    `form:"type"`
	Path      string    `form:"path"`
	Folder    string    `form:"folder"` // Alias for Path
	Name      string    `form:"name"`
	Title     string    `form:"title"`
	Before    time.Time `form:"before" time_format:"2006-01-02"`
	After     time.Time `form:"after" time_format:"2006-01-02"`
	Favorite  bool      `form:"favorite"`
	Unsorted  bool      `form:"unsorted"`
	Video     bool      `form:"video"`
	Vector    bool      `form:"vector"`
	Animated  bool      `form:"animated"`
	Photo     bool      `form:"photo"`
	Raw       bool      `form:"raw"`
	Live      bool      `form:"live"`
	Scan      bool      `form:"scan"`
	Panorama  bool      `form:"panorama"`
	Portrait  bool      `form:"portrait"`
	Landscape bool      `form:"landscape"`
	Square    bool      `form:"square"`
	Archived  bool      `form:"archived"`
	Public    bool      `form:"public"`
	Private   bool      `form:"private"`
	Review    bool      `form:"review"`
	Quality   int       `form:"quality"`
	Faces     string    `form:"faces"` // Find or exclude faces if detected.
	Lat       float32   `form:"lat"`
	Lng       float32   `form:"lng"`
	S2        string    `form:"s2"`
	Olc       string    `form:"olc"`
	Dist      uint      `form:"dist"`
	Face      string    `form:"face"`     // UIDs
	Subject   string    `form:"subject"`  // UIDs
	Person    string    `form:"person"`   // Alias for Subject
	Subjects  string    `form:"subjects"` // Text
	People    string    `form:"people"`   // Alias for Subjects
	Keywords  string    `form:"keywords"`
	Album     string    `form:"album"`
	Albums    string    `form:"albums"`
	Country   string    `form:"country"`
	State     string    `form:"state"` // Moments
	City      string    `form:"city"`
	Year      string    `form:"year"`  // Moments
	Month     string    `form:"month"` // Moments
	Day       string    `form:"day"`   // Moments
	Color     string    `form:"color"`
	Camera    int       `form:"camera"`
	Lens      int       `form:"lens"`
	Count     int       `form:"count" serialize:"-"`
	Offset    int       `form:"offset" serialize:"-"`
}

// GetQuery returns the query parameter as string.
func (f *SearchPhotosGeo) GetQuery() string {
	return f.Query
}

// SetQuery sets the query parameter.
func (f *SearchPhotosGeo) SetQuery(q string) {
	f.Query = q
}

// ParseQueryString parses the query parameter if possible.
func (f *SearchPhotosGeo) ParseQueryString() error {
	err := ParseQueryString(f)

	if f.Path != "" {
		f.Folder = ""
	} else if f.Folder != "" {
		f.Path = f.Folder
		f.Folder = ""
	}

	if f.Subject != "" {
		f.Person = ""
	} else if f.Person != "" {
		f.Subject = f.Person
		f.Person = ""
	}

	if f.Subjects != "" {
		f.People = ""
	} else if f.People != "" {
		f.Subjects = f.People
		f.People = ""
	}

	if f.Filter != "" {
		if err := Unserialize(f, f.Filter); err != nil {
			return err
		}
	}

	// Strip file extensions if any.
	if f.Name != "" {
		f.Name = fs.StripKnownExt(f.Name)
	}

	return err
}

// Serialize returns a string containing non-empty fields and values of a struct.
func (f *SearchPhotosGeo) Serialize() string {
	return Serialize(f, false)
}

// SerializeAll returns a string containing all non-empty fields and values of a struct.
func (f *SearchPhotosGeo) SerializeAll() string {
	return Serialize(f, true)
}

func NewGeoSearch(query string) SearchPhotosGeo {
	return SearchPhotosGeo{Query: query}
}
