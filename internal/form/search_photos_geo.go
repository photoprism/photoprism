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
	Type      string    `form:"type"`
	Path      string    `form:"path"`
	Folder    string    `form:"folder"` // Alias for Path
	Name      string    `form:"name"`
	Title     string    `form:"title"`
	Before    time.Time `form:"before" time_format:"2006-01-02"`
	After     time.Time `form:"after" time_format:"2006-01-02"`
	Favorite  string    `form:"favorite" example:"favorite:yes" notes:"Finds favorites only"`
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
	Quality   int       `form:"quality" notes:"Minimum quality score (1-7)"`
	Face      string    `form:"face" notes:"Face ID, yes, no, new, or kind"`
	Faces     string    `form:"faces"` // Find or exclude faces if detected.
	Subject   string    `form:"subject"`
	Near      string    `form:"near" example:"near:pqbcf5j446s0futy" notes:"Finds nearby pictures (UID)"`
	S2        string    `form:"s2" example:"s2:4799e370ca54c8b9"  notes:"S2 Position (Cell ID)"`
	Olc       string    `form:"olc" example:"olc:8FWCHX7W+" notes:"OLC Position (Open Location Code)"`
	Lat       float64   `form:"lat" example:"lat:41.894043" notes:"GPS Position (Latitude)"`
	Lng       float64   `form:"lng" example:"lng:-87.62448" notes:"GPS Position (Longitude)"`
	Alt       string    `form:"alt" example:"alt:300-500" notes:"GPS Altitude (m)"`
	Dist      float64   `form:"dist" example:"dist:50" notes:"Distance to Position (km)"`
	Latlng    string    `form:"latlng" notes:"GPS Bounding Box (Lat N, Lng E, Lat S, Lng W)"`
	Camera    int       `form:"camera"`
	Lens      int       `form:"lens"`
	Iso       string    `form:"iso" example:"iso:200-400" notes:"ISO Number (light sensitivity)"`
	Mm        string    `form:"mm" example:"mm:28-35" notes:"Focal Length (35mm equivalent)"`
	F         string    `form:"f" example:"f:2.8-4.5" notes:"Aperture (f-number)"`
	Color     string    `form:"color"`
	Chroma    int16     `form:"chroma" example:"chroma:70" notes:"Chroma (0-100)"`
	Mono      bool      `form:"mono" notes:"Finds pictures with few or no colors"`
	Person    string    `form:"person"`   // Alias for Subject
	Subjects  string    `form:"subjects"` // Text
	People    string    `form:"people"`   // Alias for Subjects
	Keywords  string    `form:"keywords" example:"keywords:\"sand&water\"" notes:"Keywords (combinable with & and |)"`
	Label     string    `form:"label" example:"label:cat|dog" notes:"Label Names (separate with |)"`
	Category  string    `form:"category" example:"category:airport" notes:"Location Category"`
	Album     string    `form:"album" example:"album:berlin" notes:"Album UID or Name, supports * wildcards"`
	Albums    string    `form:"albums" example:"albums:\"South Africa & Birds\"" notes:"Album Names (combinable with & and |)"`
	Country   string    `form:"country"`
	State     string    `form:"state"` // Moments
	City      string    `form:"city"`
	Year      string    `form:"year"`  // Moments
	Month     string    `form:"month"` // Moments
	Day       string    `form:"day"`   // Moments
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
