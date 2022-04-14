package form

import (
	"time"
)

// SearchPhotos represents search form fields for "/api/v1/photos".
type SearchPhotos struct {
	Query     string    `form:"q"`
	Filter    string    `form:"filter" notes:"-"`
	UID       string    `form:"uid" example:"uid:pqbcf5j446s0futy"`
	Type      string    `form:"type" example:"type:raw" notes:"Can be combined with |. Options: image, video, raw, live, animated"`
	Path      string    `form:"path" example:"path:2020/Holiday, path:\"*/2020\"" notes:"Same as folder. Can be combined with |."`
	Folder    string    `form:"folder" example:"folder:2020/Holiday, folder:\"*/2020\"" notes:"Same as path. Can be combined with |."` // Alias for Path
	Name      string    `form:"name" example:"name:\"IMG_9831-112\", name:\"IMG_9831-112*\"" notes:"Can be combined with |."`
	Filename  string    `form:"filename" example:"filename:\"2021/07/12345.jpg\"" notes:"Can be combined with |."`
	Original  string    `form:"original" example:"original:\"IMG_9831-112\", original:\"IMG_9831-112*\"" notes:"Can be combined with |. Only applicable when files have been imported"`
	Title     string    `form:"title" example:"title:\"Lake*\"" notes:"Can be combined with |."`
	Hash      string    `form:"hash" example:"hash:2fd4e1c67a2d" notes:"Can be combined with |."`
	Primary   bool      `form:"primary"`
	Stack     bool      `form:"stack"`
	Unstacked bool      `form:"unstacked"`
	Stackable bool      `form:"stackable"`
	Video     bool      `form:"video"`
	Vector    bool      `form:"vector" notes:"Vector Graphics"`
	Animated  bool      `form:"animated" notes:"Animated GIFs"`
	Photo     bool      `form:"photo" notes:"No Videos"`
	Raw       bool      `form:"raw" notes:"RAW Images"`
	Live      bool      `form:"live" notes:"Live Photos, Short Videos"`
	Scan      bool      `form:"scan" notes:"Scanned Images, Documents"`
	Panorama  bool      `form:"panorama" notes:"Aspect Ratio > 1.9:1"`
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
	Dist      uint      `form:"dist" notes:"Distance to coordinates (radius in kilometre). Only applicable in combination with the lat/lng filters."`
	Fmin      float32   `form:"fmin"`
	Fmax      float32   `form:"fmax"`
	Chroma    uint8     `form:"chroma"`
	Diff      uint32    `form:"diff"`
	Mono      bool      `form:"mono"`
	Geo       bool      `form:"geo"`
	Keywords  string    `form:"keywords"  example:"keywords:\"buffalo&water\"" notes:"Keywords can be combined with & and |"`                                                                                           // Filter by keyword(s)
	Label     string    `form:"label" example:"label:cat|dog" notes:"Can be combined with |."`                                                                                                                          // Label name
	Category  string    `form:"category"`                                                                                                                                                                               // Moments
	Country   string    `form:"country" example:"country:\"de|us\"" notes:"Can be combined with |."`                                                                                                                    // Moments
	State     string    `form:"state" example:"state:\"Baden-Württemberg\"" notes:"Can be combined with |."`                                                                                                            // Moments
	Year      string    `form:"year" example:"year:1990|2003" notes:"Can be combined with |."`                                                                                                                          // Moments
	Month     string    `form:"month" example:"month:7|10" notes:"Can be combined with |."`                                                                                                                             // Moments
	Day       string    `form:"day" example:"day:3|13" notes:"Can be combined with |."`                                                                                                                                 // Moments
	Face      string    `form:"face"`                                                                                                                                                                                   // UIDs
	Subject   string    `form:"subject" example:"subject:\"Jane Doe & John Doe\"" notes:"Same as person. Only exact matches. Names can be combined with & and |"`                                                       // UIDs
	Person    string    `form:"person" example:"person:\"Jane Doe & John Doe\"" notes:"Same as subject. Only exact matches. Names can be combined with & and |"`                                                        // Alias for Subject
	Subjects  string    `form:"subjects" example:"subjects:\"Jane & John\"" notes:"Same as people. Names can be combined with & and |"`                                                                                 // People names
	People    string    `form:"people" example:"people:\"Jane & John\"" notes:"Same as subjects. Names can be combined with & and |"`                                                                                   // Alias for Subjects
	Album     string    `form:"album" example:"album:berlin" notes:"Single name with * wildcard"`                                                                                                                       // Album UIDs or name
	Albums    string    `form:"albums" example:"albums:\"South Africa & Birds\"" notes:"Album names can be combined with & and |"`                                                                                      // Multi search with and/or
	Color     string    `form:"color" example:"color:\"red|blue\"" notes:"Can be combined with |. Options: purple, magenta, pink, red, orange, gold, yellow, lime, green, teal, cyan, blue, brown, white, grey, black"` // Main color
	Faces     string    `form:"faces" example:"faces:yes faces:no faces:3" notes:"3 means minimum 3 faces"`                                                                                                             // Find or exclude faces if detected.
	Quality   int       `form:"quality" notes:"Options: 0, 1, 2, 3, 4, 5"`                                                                                                                                              // Photo quality score
	Review    bool      `form:"review"`                                                                                                                                                                                 // Find photos in review
	Camera    string    `form:"camera" example:"camera:canon"`                                                                                                                                                          // Camera UID or name
	Lens      string    `form:"lens" example:"lens:ef24"`                                                                                                                                                               // Lens UID or name
	Before    time.Time `form:"before" time_format:"2006-01-02" notes:"Taken before this date"`                                                                                                                         // Finds images taken before date
	After     time.Time `form:"after" time_format:"2006-01-02" notes:"Taken after this date"`                                                                                                                           // Finds images taken after date
	Count     int       `form:"count" binding:"required" serialize:"-"`                                                                                                                                                 // Result FILE limit
	Offset    int       `form:"offset" serialize:"-"`                                                                                                                                                                   // Result FILE offset
	Order     string    `form:"order" serialize:"-"`                                                                                                                                                                    // Sort order
	Merged    bool      `form:"merged" serialize:"-"`                                                                                                                                                                   // Merge FILES in response
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
