package form

import (
	"time"

	"github.com/photoprism/photoprism/pkg/fs"
)

// SearchPhotos represents search form fields for "/api/v1/photos".
type SearchPhotos struct {
	Query     string    `form:"q"`
	Scope     string    `form:"s" serialize:"-" example:"s:ariqwb43p5dh9h13" notes:"Limits the results to one album or another scope, if specified"`
	Filter    string    `form:"filter" serialize:"-" notes:"-"`
	ID        string    `form:"id" example:"id:123e4567-e89b-..." notes:"Finds pictures by Exif UID, XMP Document ID or Instance ID"`
	UID       string    `form:"uid" example:"uid:pqbcf5j446s0futy" notes:"Limits results to the specified internal unique IDs"`
	Type      string    `form:"type" example:"type:raw" notes:"Media Type (image, video, raw, live, animated); separate with |"`
	Path      string    `form:"path" example:"path:2020/Holiday" notes:"Path Name (separate with |), supports * wildcards"`
	Folder    string    `form:"folder" example:"folder:\"*/2020\"" notes:"Path Name (separate with |), supports * wildcards"` // Alias for Path
	Name      string    `form:"name" example:"name:\"IMG_9831-112*\"" notes:"File Name without path and extension (separate with |)"`
	Filename  string    `form:"filename" example:"filename:\"2021/07/12345.jpg\"" notes:"File Name with path and extension (separate with |)"`
	Original  string    `form:"original" example:"original:\"IMG_9831-112*\"" notes:"Original file name of imported files (separate with |)"`
	Title     string    `form:"title" example:"title:\"Lake*\"" notes:"Title (separate with |)"`
	Hash      string    `form:"hash" example:"hash:2fd4e1c67a2d" notes:"SHA1 File Hash (separate with |)"`
	Primary   bool      `form:"primary" notes:"Finds primary JPEG files only"`
	Stack     bool      `form:"stack" notes:"Finds pictures with more than one media file"`
	Unstacked bool      `form:"unstacked" notes:"Finds pictures with a file that has been removed from a stack"`
	Stackable bool      `form:"stackable" notes:"Finds pictures that can be stacked with additional media files"`
	Video     bool      `form:"video" notes:"Finds video files only"`
	Vector    bool      `form:"vector" notes:"Finds vector graphics only"`
	Animated  bool      `form:"animated" notes:"Finds animated GIFs"`
	Photo     bool      `form:"photo" notes:"Finds only photos, no videos"`
	Raw       bool      `form:"raw" notes:"Finds pictures with RAW image file"`
	Live      bool      `form:"live" notes:"Finds Live Photos and short videos"`
	Scan      string    `form:"scan" example:"scan:true scan:false" notes:"Finds scanned photos and documents"`
	Panorama  bool      `form:"panorama" notes:"Finds pictures with an aspect ratio > 1.9:1"`
	Portrait  bool      `form:"portrait" notes:"Finds pictures in portrait format"`
	Landscape bool      `form:"landscape" notes:"Finds pictures in landscape format"`
	Square    bool      `form:"square" notes:"Finds images with an aspect ratio of 1:1"`
	Error     bool      `form:"error" notes:"Finds pictures with errors"`
	Hidden    bool      `form:"hidden" notes:"Finds hidden pictures (broken or unsupported)"`
	Archived  bool      `form:"archived" notes:"Finds archived pictures"`
	Public    bool      `form:"public" notes:"Excludes private pictures"`
	Private   bool      `form:"private" notes:"Finds private pictures"`
	Favorite  string    `form:"favorite" example:"favorite:true favorite:false" notes:"Finds images by favorite status"`
	Unsorted  bool      `form:"unsorted" notes:"Finds pictures not in an album"`
	Near      string    `form:"near" example:"near:pqbcf5j446s0futy" notes:"Finds nearby pictures (UID)"`
	S2        string    `form:"s2" example:"s2:4799e370ca54c8b9"  notes:"S2 Position (Cell ID)"`
	Olc       string    `form:"olc" example:"olc:8FWCHX7W+" notes:"OLC Position (Open Location Code)"`
	Lat       float64   `form:"lat" example:"lat:41.894043" notes:"GPS Position (Latitude)"`
	Lng       float64   `form:"lng" example:"lng:-87.62448" notes:"GPS Position (Longitude)"`
	Dist      float64   `form:"dist" example:"dist:50" notes:"Distance to Position (km)"`
	Latlng    string    `form:"latlng" notes:"GPS Bounding Box (Lat N, Lng E, Lat S, Lng W)"`
	Fmin      float32   `form:"fmin" notes:"F-number (min)"`
	Fmax      float32   `form:"fmax" notes:"F-number (max)"`
	Chroma    int16     `form:"chroma" example:"chroma:70" notes:"Chroma (0-100)"`
	Diff      uint32    `form:"diff" notes:"Differential Perceptual Hash (000000-FFFFFF)"`
	Mono      bool      `form:"mono" notes:"Finds pictures with few or no colors"`
	Geo       string    `form:"geo" example:"geo:yes" notes:"Finds pictures with or without coordinates"`
	Keywords  string    `form:"keywords" example:"keywords:\"sand&water\"" notes:"Keywords (combinable with & and |)"`
	Label     string    `form:"label" example:"label:cat|dog" notes:"Label Names (separate with |)"`
	Category  string    `form:"category" example:"category:airport" notes:"Location Category"`
	Country   string    `form:"country" example:"country:\"de|us\"" notes:"Location Country Code (separate with |)"`                                                                                                  // Moments
	State     string    `form:"state" example:"state:\"Baden-Württemberg\"" notes:"Location State (separate with |)"`                                                                                                 // Moments
	City      string    `form:"city" example:"city:\"Berlin\"" notes:"Location City (separate with |)"`                                                                                                               // Moments
	Year      string    `form:"year" example:"year:1990|2003" notes:"Year (separate with |)"`                                                                                                                         // Moments
	Month     string    `form:"month" example:"month:7|10" notes:"Month (1-12, separate with |)"`                                                                                                                     // Moments
	Day       string    `form:"day" example:"day:3|13" notes:"Day of Month (1-31, separate with |)"`                                                                                                                  // Moments
	Face      string    `form:"face" example:"face:PN6QO5INYTUSAATOFL43LL2ABAV5ACZG" notes:"Face ID, yes, no, new, or kind"`                                                                                          // UIDs
	Faces     string    `form:"faces" example:"faces:yes faces:3" notes:"Minimum number of Faces (yes = 1)"`                                                                                                          // Find or exclude faces if detected.
	Subject   string    `form:"subject" example:"subject:\"Jane Doe & John Doe\"" notes:"Alias for person"`                                                                                                           // UIDs
	Person    string    `form:"person" example:"person:\"Jane Doe & John Doe\"" notes:"Subject Names, exact matches (combinable with & and |)"`                                                                       // Alias for Subject
	Subjects  string    `form:"subjects" example:"subjects:\"Jane & John\"" notes:"Alias for people"`                                                                                                                 // People names
	People    string    `form:"people" example:"people:\"Jane & John\"" notes:"Subject Names (combinable with & and |)"`                                                                                              // Alias for Subjects
	Album     string    `form:"album" example:"album:berlin" notes:"Album UID or Name, supports * wildcards"`                                                                                                         // Album UIDs or name
	Albums    string    `form:"albums" example:"albums:\"South Africa & Birds\"" notes:"Album Names (combinable with & and |)"`                                                                                       // Multi search with and/or
	Color     string    `form:"color" example:"color:\"red|blue\"" notes:"Color Name (purple, magenta, pink, red, orange, gold, yellow, lime, green, teal, cyan, blue, brown, white, grey, black) (separate with |)"` // Main color
	Quality   int       `form:"quality" notes:"Minimum quality score (1-7)"`                                                                                                                                          // Photo quality score
	Review    bool      `form:"review" notes:"Finds pictures in review"`                                                                                                                                              // Find photos in review
	Camera    string    `form:"camera" example:"camera:canon" notes:"Camera Make/Model Name"`                                                                                                                         // Camera UID or name
	Lens      string    `form:"lens" example:"lens:ef24" notes:"Lens Make/Model Name"`                                                                                                                                // Lens UID or name
	Before    time.Time `form:"before" time_format:"2006-01-02" notes:"Finds pictures taken before this date"`                                                                                                        // Finds images taken before date
	After     time.Time `form:"after" time_format:"2006-01-02" notes:"Finds pictures taken after this date"`                                                                                                          // Finds images taken after date
	Count     int       `form:"count" binding:"required" serialize:"-"`                                                                                                                                               // Result FILE limit
	Offset    int       `form:"offset" serialize:"-"`                                                                                                                                                                 // Result FILE offset
	Order     string    `form:"order" serialize:"-"`                                                                                                                                                                  // Sort order
	Merged    bool      `form:"merged" serialize:"-"`                                                                                                                                                                 // Merge FILES in response
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
