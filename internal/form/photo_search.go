package form

import (
	"time"
)

// PhotoSearch represents search form fields for "/api/v1/photos".
type PhotoSearch struct {
	Query     string    `form:"q"`
	Filter    string    `form:"filter"`
	ID        string    `form:"id"`
	Type      string    `form:"type"`
	Path      string    `form:"path"`
	Folder    string    `form:"folder"` // Alias for Path
	Name      string    `form:"name"`
	Title     string    `form:"title"`
	Hash      string    `form:"hash"`
	Video     bool      `form:"video"`
	Photo     bool      `form:"photo"`
	Duplicate bool      `form:"duplicate"`
	Archived  bool      `form:"archived"`
	Error     bool      `form:"error"`
	Lat       float32   `form:"lat"`
	Lng       float32   `form:"lng"`
	Dist      uint      `form:"dist"`
	Fmin      float32   `form:"fmin"`
	Fmax      float32   `form:"fmax"`
	Chroma    uint8     `form:"chroma"`
	Diff      uint32    `form:"diff"`
	Mono      bool      `form:"mono"`
	Portrait  bool      `form:"portrait"`
	Location  bool      `form:"location"`
	Album     string    `form:"album"`
	Label     string    `form:"label"`
	Category  string    `form:"category"` // Moments
	Country   string    `form:"country"`  // Moments
	State     string    `form:"state"`    // Moments
	Year      int       `form:"year"`     // Moments
	Month     int       `form:"month"`    // Moments
	Color     string    `form:"color"`
	Quality   int       `form:"quality"`
	Review    bool      `form:"review"`
	Camera    int       `form:"camera"`
	Lens      int       `form:"lens"`
	Before    time.Time `form:"before" time_format:"2006-01-02"`
	After     time.Time `form:"after" time_format:"2006-01-02"`
	Favorite  bool      `form:"favorite"`
	Public    bool      `form:"public"`
	Private   bool      `form:"private"`
	Safe      bool      `form:"safe"`
	Count     int       `form:"count" binding:"required" serialize:"-"`
	Offset    int       `form:"offset" serialize:"-"`
	Order     string    `form:"order" serialize:"-"`
	Merged    bool      `form:"merged" serialize:"-"`
}

func (f *PhotoSearch) GetQuery() string {
	return f.Query
}

func (f *PhotoSearch) SetQuery(q string) {
	f.Query = q
}

func (f *PhotoSearch) ParseQueryString() error {
	if err := ParseQueryString(f); err != nil {
		return err
	}

	if f.Path == "" && f.Folder != "" {
		f.Path = f.Folder
	}

	if f.Filter != "" {
		if err := Unserialize(f, f.Filter); err != nil {
			return err
		}
	}

	return nil
}

// Serialize returns a string containing non-empty fields and values of a struct.
func (f *PhotoSearch) Serialize() string {
	return Serialize(f, false)
}

// SerializeAll returns a string containing all non-empty fields and values of a struct.
func (f *PhotoSearch) SerializeAll() string {
	return Serialize(f, true)
}

func NewPhotoSearch(query string) PhotoSearch {
	return PhotoSearch{Query: query}
}
