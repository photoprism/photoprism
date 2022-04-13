package form

import (
	"time"
)

// SearchPhotos represents search form fields for "/api/v1/photos".
type SearchPhotos struct {
	Query     string    `form:"q"`
	Filter    string    `form:"filter" notes:"-"`
	UID       string    `form:"uid"`
	Type      string    `form:"type"`
	Path      string    `form:"path"`
	Folder    string    `form:"folder"` // Alias for Path
	Name      string    `form:"name"`
	Filename  string    `form:"filename"`
	Original  string    `form:"original"`
	Title     string    `form:"title"`
	Hash      string    `form:"hash" example:"hash:2fd4e1c67a2d"`
	Primary   bool      `form:"primary"`
	Stack     bool      `form:"stack"`
	Unstacked bool      `form:"unstacked"`
	Stackable bool      `form:"stackable"`
	Video     bool      `form:"video"`
	Photo     bool      `form:"photo"`
	Raw       bool      `form:"raw"`
	Live      bool      `form:"live"`
	Scan      bool      `form:"scan"`
	Panorama  bool      `form:"panorama"`
	Portrait  bool      `form:"portrait"`
	Landscape bool      `form:"landscape"`
	Square    bool      `form:"square"`
	Error     bool      `form:"error"`
	Hidden    bool      `form:"hidden"`
	Archived  bool      `form:"archived"`
	Public    bool      `form:"public"`
	Private   bool      `form:"private"`
	Favorite  bool      `form:"favorite"`
	Unsorted  bool      `form:"unsorted"`
	Lat       float32   `form:"lat"`
	Lng       float32   `form:"lng"`
	Dist      uint      `form:"dist"`
	Fmin      float32   `form:"fmin"`
	Fmax      float32   `form:"fmax"`
	Chroma    uint8     `form:"chroma"`
	Diff      uint32    `form:"diff"`
	Mono      bool      `form:"mono"`
	Geo       bool      `form:"geo"`
	Keywords  string    `form:"keywords"`                                                                              // Filter by keyword(s)
	Label     string    `form:"label"`                                                                                 // Label name
	Category  string    `form:"category"`                                                                              // Moments
	Country   string    `form:"country"`                                                                               // Moments
	State     string    `form:"state"`                                                                                 // Moments
	Year      string    `form:"year"`                                                                                  // Moments
	Month     string    `form:"month"`                                                                                 // Moments
	Day       string    `form:"day"`                                                                                   // Moments
	Face      string    `form:"face"`                                                                                  // UIDs
	Subject   string    `form:"subject"`                                                                               // UIDs
	Person    string    `form:"person"`                                                                                // Alias for Subject
	Subjects  string    `form:"subjects"`                                                                              // People names
	People    string    `form:"people"`                                                                                // Alias for Subjects
	Album     string    `form:"album" notes:"single name with * wildcard"`                                             // Album UIDs or name
	Albums    string    `form:"albums" example:"albums:\"South Africa & Birds\"" notes:"can be combined with & and |"` // Multi search with and/or
	Color     string    `form:"color"`                                                                                 // Main color
	Faces     string    `form:"faces"`                                                                                 // Find or exclude faces if detected.
	Quality   int       `form:"quality"`                                                                               // Photo quality score
	Review    bool      `form:"review"`                                                                                // Find photos in review
	Camera    string    `form:"camera" example:"camera:canon"`                                                         // Camera UID or name
	Lens      string    `form:"lens" example:"lens:ef24"`                                                              // Lens UID or name
	Before    time.Time `form:"before" time_format:"2006-01-02" notes:"taken before this date"`                        // Finds images taken before date
	After     time.Time `form:"after" time_format:"2006-01-02" notes:"taken after this date"`                          // Finds images taken after date
	Count     int       `form:"count" binding:"required" serialize:"-"`                                                // Result FILE limit
	Offset    int       `form:"offset" serialize:"-"`                                                                  // Result FILE offset
	Order     string    `form:"order" serialize:"-"`                                                                   // Sort order
	Merged    bool      `form:"merged" serialize:"-"`                                                                  // Merge FILES in response
}

func (f *SearchPhotos) GetQuery() string {
	return f.Query
}

func (f *SearchPhotos) SetQuery(q string) {
	f.Query = q
}

func (f *SearchPhotos) ParseQueryString() error {
	if err := ParseQueryString(f); err != nil {
		return err
	}

	if f.Path == "" && f.Folder != "" {
		f.Path = f.Folder
		f.Folder = ""
	}

	if f.Subject == "" && f.Person != "" {
		f.Subject = f.Person
		f.Person = ""
	}

	if f.Subjects == "" && f.People != "" {
		f.Subjects = f.People
		f.People = ""
	}

	if f.Filter != "" {
		if err := Unserialize(f, f.Filter); err != nil {
			return err
		}
	}

	return nil
}

// Serialize returns a string containing non-empty fields and values of a struct.
func (f *SearchPhotos) Serialize() string {
	return Serialize(f, false)
}

// SerializeAll returns a string containing all non-empty fields and values of a struct.
func (f *SearchPhotos) SerializeAll() string {
	return Serialize(f, true)
}

func NewPhotoSearch(query string) SearchPhotos {
	return SearchPhotos{Query: query}
}
