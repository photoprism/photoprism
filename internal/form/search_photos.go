package form

import (
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SearchPhotos represents search form fields for "/api/v1/photos".
type SearchPhotos struct {
	Query       string    `form:"q"`
	Scope       string    `form:"s" serialize:"-" example:"s:ariqwb43p5dh9h13" notes:"Restricts results to the specified album UID or other supported scopes"`
	Filter      string    `form:"filter" serialize:"-" notes:"-"`
	ID          string    `form:"id" example:"id:123e4567-e89b-..." notes:"Finds content with the specified Image, Document or Instance IDs, separated by |"`
	UID         string    `form:"uid" example:"uid:pqbcf5j446s0futy" notes:"Finds content with the specified internal UIDs, separated by |"`
	Type        string    `form:"type" example:"type:image|raw|live" notes:"Finds specific media types, such as image, raw, live, video, animated, audio, vector, or document, separated by |"`
	Path        string    `form:"path" example:"path:2020/Holiday" notes:"Path names separated by |, supports * wildcards"`
	Folder      string    `form:"folder" example:"folder:\"*/2020\"" notes:"Alias for the path filter"` // Alias for Path
	Name        string    `form:"name" example:"name:\"IMG_9831-112*\"" notes:"File names without path and extension, separated by |"`
	Filename    string    `form:"filename" example:"filename:\"2021/07/12345.jpg\"" notes:"File names including path and extension, separated by |"`
	Original    string    `form:"original" example:"original:\"IMG_9831-112*\"" notes:"Original file names of imported files, separated by |"`
	Title       string    `form:"title" example:"title:\"Lake*\"" notes:"Searches text in titles separated by |, or specify false to find content without a title"`
	Caption     string    `form:"caption" example:"caption:\"Lake*\"" notes:"Searches text in captions separated by |, or specify false to find content without a caption"`
	Description string    `form:"description" example:"description:\"Lake*\"" notes:"Searches text in titles or captions separated by |, or specify false to find content without a title or caption"`
	Hash        string    `form:"hash" example:"hash:2fd4e1c67a2d" notes:"SHA1 file hashes, separated by |"`
	Primary     bool      `form:"primary" notes:"Finds primary JPEG or PNG files only"`
	Stack       bool      `form:"stack" notes:"Finds content with more than one media file"`
	Unstacked   bool      `form:"unstacked" notes:"Finds content with a file that has been removed"`
	Stackable   bool      `form:"stackable" notes:"Finds content that can be stacked with additional files"`
	Photo       bool      `form:"photo" notes:"Finds regular photos and images, as well as RAW and Live Photos"`
	Image       bool      `form:"image" notes:"Finds regular photos and images only"`
	Raw         bool      `form:"raw" notes:"Finds RAW images only"`
	Media       bool      `form:"media" notes:"Finds live, video, audio, and animated content only"`
	Animated    bool      `form:"animated" notes:"Finds animated images only"`
	Audio       bool      `form:"audio" notes:"Finds audio content only"`
	Video       bool      `form:"video" notes:"Finds video content only"`
	Live        bool      `form:"live" notes:"Finds Motion and Live Photos only"`
	Vector      bool      `form:"vector" notes:"Finds vector graphics only"`
	Document    bool      `form:"document" notes:"Finds PDF documents only"`
	Scan        string    `form:"scan" example:"scan:true scan:false" notes:"Finds scanned photos and documents"`
	Mp          string    `form:"mp" example:"mp:3-6" notes:"Resolution in Megapixels (MP)"`
	Panorama    bool      `form:"panorama" notes:"Finds panorama pictures only (aspect ratio 1.9:1 or more)"`
	Portrait    bool      `form:"portrait" notes:"Finds portrait pictures only"`
	Landscape   bool      `form:"landscape" notes:"Finds landscape pictures only"`
	Square      bool      `form:"square" notes:"Finds square pictures only (aspect ratio 1:1)"`
	Archived    bool      `form:"archived" notes:"Finds archived content"`
	Public      bool      `form:"public" notes:"Excludes private content"`
	Private     bool      `form:"private" notes:"Finds private content"`
	Review      bool      `form:"review" notes:"Finds content in review"`
	Error       bool      `form:"error" notes:"Finds content with errors"`
	Hidden      bool      `form:"hidden" notes:"Finds hidden content (broken or unsupported)"`
	Favorite    string    `form:"favorite" example:"favorite:true favorite:false" notes:"Finds favorite content"`
	Unsorted    bool      `form:"unsorted" notes:"Finds content that is not in an album"`
	Near        string    `form:"near" example:"near:pqbcf5j446s0futy" notes:"Finds nearby pictures (UID)"`
	S2          string    `form:"s2" example:"s2:4799e370ca54c8b9"  notes:"Position, specified as S2 Cell ID"`
	Olc         string    `form:"olc" example:"olc:8FWCHX7W+" notes:"Open Location Code (OLC)"`
	Lat         float64   `form:"lat" example:"lat:41.894043" notes:"Position latitude (-90.0 to 90.0 deg)"`
	Lng         float64   `form:"lng" example:"lng:-87.62448" notes:"Position longitude (-180.0 to 180.0 deg)"`
	Alt         string    `form:"alt" example:"alt:300-500" notes:"Altitude (m)"`
	Dist        float64   `form:"dist" example:"dist:50" notes:"Maximum distance to position in km"`
	Latlng      string    `form:"latlng" example:"latlng:49.4,13.41,46.5,2.331" notes:"Position bounding box (Lat N, Lng E, Lat S, Lng W)"`
	Camera      string    `form:"camera" example:"camera:canon" notes:"Camera make or model"`
	Lens        string    `form:"lens" example:"lens:ef24" notes:"Lens make or model"`
	Iso         string    `form:"iso" example:"iso:200-400" notes:"ISO number (light sensitivity)"`
	Mm          string    `form:"mm" example:"mm:28-35" notes:"Focal length (35mm equivalent)"`
	F           string    `form:"f" example:"f:2.8-4.5" notes:"Aperture (F-Number)"`
	Color       string    `form:"color" example:"color:\"red|blue\"" notes:"Color name separated by |, e.g. purple, magenta, pink, red, orange, gold, yellow, lime, green, teal, cyan, blue, brown, white, grey, or black"` // Main color
	Codec       string    `form:"codec" example:"codec:avc1" notes:"Media codec types separated by |, e.g. jpeg, avc1, or hvc1"`
	Chroma      int16     `form:"chroma" example:"chroma:70" notes:"Chroma (0-100)"`
	Mono        bool      `form:"mono" notes:"Pictures with few or no colors"`
	Diff        uint32    `form:"diff" notes:"Differential Perceptual Hash (000000-FFFFFF)"`
	Geo         string    `form:"geo" example:"geo:yes" notes:"Finds content with or without latitude and longitude"`
	Keywords    string    `form:"keywords" example:"keywords:\"sand&water\"" notes:"Keywords, combinable with & and |"`
	Label       string    `form:"label" example:"label:cat|dog" notes:"Label names, separated by |"`
	Category    string    `form:"category" example:"category:airport" notes:"Location category type"`
	Country     string    `form:"country" example:"country:\"de|us\"" notes:"Country codes, separated by |"`                                                                             // Moments
	State       string    `form:"state" example:"state:\"Baden-Württemberg\"" notes:"State or province names, separated by |"`                                                           // Moments
	City        string    `form:"city" example:"city:\"Berlin\"" notes:"City names, separated by |"`                                                                                     // Moments
	Year        string    `form:"year" example:"year:1990|2003" notes:"Years, separated by |"`                                                                                           // Moments
	Month       string    `form:"month" example:"month:7|10" notes:"Months from 1-12, separated by |"`                                                                                   // Moments
	Day         string    `form:"day" example:"day:3|13" notes:"Days 1-31, separated by |"`                                                                                              // Moments
	Face        string    `form:"face" example:"face:PN6QO5INYTUSAATOFL43LL2ABAV5ACZG" notes:"Find pictures with a specific face ID, you can also specify yes, no, new, or a face type"` // UIDs
	Faces       string    `form:"faces" example:"faces:yes faces:3" notes:"Minimum number of detected faces (yes means 1)"`                                                              // Find or exclude faces if detected.
	Person      string    `form:"person" example:"person:\"Jane Doe & John Doe\"" notes:"Subject names, will be matched exactly and can be combined using & or |"`                       // Alias for Subject
	Subject     string    `form:"subject" example:"subject:\"Jane Doe & John Doe\"" notes:"Alias for person"`                                                                            // UIDs
	People      string    `form:"people" example:"people:\"Jane & John\"" notes:"Subject names, combinable with & or |"`                                                                 // Alias for Subjects
	Subjects    string    `form:"subjects" example:"subjects:\"Jane & John\"" notes:"Alias for people"`                                                                                  // People names
	Album       string    `form:"album" example:"album:berlin" notes:"Album UID or name, supports * wildcards"`                                                                          // Album UIDs or name
	Albums      string    `form:"albums" example:"albums:\"South Africa & Birds\"" notes:"Album names, combinable with & or |"`                                                          // Multi search with and/or
	Quality     int       `form:"quality" notes:"Minimum quality score (1-7)"`                                                                                                           // Photo quality score
	Added       time.Time `form:"added" example:"added:\"2006-01-02T15:04:05Z\"" time_format:"2006-01-02T15:04:05Z07:00" notes:"Finds content added at or after this time"`              // Pictures added at or after this time
	Updated     time.Time `form:"updated" example:"updated:\"2006-01-02T15:04:05Z\"" time_format:"2006-01-02T15:04:05Z07:00" notes:"Finds content updated at or after this time"`        // Pictures updated at or after this time
	Edited      time.Time `form:"edited" example:"edited:\"2006-01-02T15:04:05Z\"" time_format:"2006-01-02T15:04:05Z07:00" notes:"Finds content edited at or after this time"`           // Pictures edited at or after this time
	Taken       time.Time `form:"taken" time_format:"2006-01-02" notes:"Finds content created on the specified date"`                                                                    // Pictures taken on the specified date
	Before      time.Time `form:"before" time_format:"2006-01-02" notes:"Finds content created before this date"`                                                                        // Pictures taken on or before this date"
	After       time.Time `form:"after" time_format:"2006-01-02" notes:"Finds content created on or after this date"`                                                                    // Pictures taken on or after this date
	Count       int       `form:"count" binding:"required" serialize:"-"`                                                                                                                // Result FILE limit
	Offset      int       `form:"offset" serialize:"-"`                                                                                                                                  // Result FILE offset
	Order       string    `form:"order" serialize:"-"`                                                                                                                                   // Sort order
	Reverse     bool      `form:"reverse" serialize:"-"`                                                                                                                                 // Merge FILES in response
	Merged      bool      `form:"merged" serialize:"-"`                                                                                                                                  // Merge FILES in response
	Details     bool      `form:"-" serialize:"-"`                                                                                                                                       // Include additional information from details table
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

	// Try to parse remaining query into latitude and longitude.
	if q := f.GetQuery(); q == "" {
		// No remaining query to parse.
	} else if lat, lng, parseErr := txt.Position(q); parseErr == nil {
		// Use coordinates only if no other coordinates are set.
		if f.Lat == 0.0 && f.Lng == 0.0 && f.Latlng == "" {
			f.Lat = lat
			f.Lng = lng
		}

		// Remove from query.
		f.SetQuery("")
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

// FindUidOnly checks if search filters other than UID may be skipped to improve performance.
func (f *SearchPhotos) FindUidOnly() bool {
	return f.UID != "" && f.Query == "" && f.Scope == "" && f.Filter == "" && f.Album == "" && f.Albums == ""
}

func NewSearchPhotos(query string) SearchPhotos {
	return SearchPhotos{Query: query}
}
