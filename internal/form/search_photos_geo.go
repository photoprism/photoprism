package form

import (
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
)

// SearchPhotosGeo represents search form fields for "/api/v1/geo".
type SearchPhotosGeo struct {
	Query     string    `form:"q"`
	Scope     string    `form:"s" serialize:"-" example:"s:ariqwb43p5dh9h13" notes:"Limits the results to one album or another scope, if specified"`
	Filter    string    `form:"filter" serialize:"-" notes:"-"`
	ID        string    `form:"id" example:"id:123e4567-e89b-..." notes:"Finds pictures by Exif UID, XMP Document ID or Instance ID"`
	UID       string    `form:"uid" example:"uid:pqbcf5j446s0futy" notes:"Limits results to the specified internal unique IDs"`
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
	Scan      string    `form:"scan" example:"scan:true scan:false" notes:"Finds scanned photos and documents"`
	Panorama  bool      `form:"panorama"`
	Portrait  bool      `form:"portrait"`
	Landscape bool      `form:"landscape"`
	Square    bool      `form:"square"`
	Archived  bool      `form:"archived"`
	Public    bool      `form:"public"`
	Private   bool      `form:"private"`
	Review    bool      `form:"review"`
	Quality   int       `form:"quality"`
	Face      string    `form:"face" notes:"Face ID, yes, no, new, or kind"`
	Faces     string    `form:"faces"` // Find or exclude faces if detected.
	Subject   string    `form:"subject"`
	Lat       float32   `form:"lat"`
	Lng       float32   `form:"lng"`
	S2        string    `form:"s2"`
	Olc       string    `form:"olc"`
	Dist      uint      `form:"dist"`
	Person    string    `form:"person"`   // Alias for Subject
	Subjects  string    `form:"subjects"` // Text
	People    string    `form:"people"`   // Alias for Subjects
	Chroma    int16     `form:"chroma" example:"chroma:70" notes:"Chroma (0-100)"`
	Mono      bool      `form:"mono" notes:"Finds pictures with few or no colors"`
	Keywords  string    `form:"keywords"`
	Album     string    `form:"album" example:"album:berlin" notes:"Album UID or Name, supports * wildcards"`
	Albums    string    `form:"albums" example:"albums:\"South Africa & Birds\"" notes:"Album Names, can be combined with & and |"`
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

// FindUidOnly checks if search filters other than UID may be skipped to improve performance.
func (f *SearchPhotosGeo) FindUidOnly() bool {
	return f.UID != "" && f.Query == "" && f.Scope == "" && f.Filter == "" && f.Album == "" && f.Albums == ""
}

func NewSearchPhotosGeo(query string) SearchPhotosGeo {
	return SearchPhotosGeo{Query: query}
}
