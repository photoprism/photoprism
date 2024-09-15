package form

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestForm struct {
	Query    string    `form:"q"`
	ID       string    `form:"id"`
	Type     string    `form:"type"`
	Path     string    `form:"path"`
	Folder   string    `form:"folder"`
	Name     string    `form:"name"`
	Title    string    `form:"title"`
	Hash     string    `form:"hash"`
	Video    bool      `form:"video"`
	Photo    bool      `form:"photo"`
	Archived bool      `form:"archived"`
	Error    bool      `form:"error"`
	Lat      float64   `form:"lat"`
	Lng      float64   `form:"lng"`
	Dist     uint      `form:"dist"`
	Color    string    `form:"color"`
	Chroma   int16     `form:"chroma"`
	Mono     bool      `form:"mono"`
	Diff     uint32    `form:"diff"`
	Portrait bool      `form:"portrait"`
	Location bool      `form:"location"`
	Album    string    `form:"album"`
	Label    string    `form:"label"`
	Country  string    `form:"country"`
	Year     int       `form:"year"`
	Month    int       `form:"month"`
	Quality  int       `form:"quality"`
	Review   bool      `form:"review"`
	Camera   int       `form:"camera"`
	Lens     int       `form:"lens"`
	Before   time.Time `form:"before" time_format:"2006-01-02"`
	After    time.Time `form:"after" time_format:"2006-01-02"`
	Favorite bool      `form:"favorite"`
	Public   bool      `form:"public"`
	Private  bool      `form:"private"`
	Safe     bool      `form:"safe"`
	Count    int       `form:"count" binding:"required" serialize:"-"`
	Offset   int       `form:"offset" serialize:"-"`
	Order    string    `form:"order" serialize:"-"`
	Merged   bool      `form:"merged" serialize:"-"`
}

func (f *TestForm) GetQuery() string {
	return f.Query
}

func (f *TestForm) SetQuery(q string) {
	f.Query = q
}

func TestSerialize(t *testing.T) {
	form := TestForm{
		Query:   "foo BAR",
		Name:    "yo/ba:z.JPG",
		Private: true,
		Photo:   false,
		Lat:     1.5,
		Lng:     -10.33333,
		Year:    2002,
		Chroma:  1,
		Diff:    424242,
		Count:   100,
		Order:   "name",
		Before:  time.Date(2019, 01, 15, 0, 0, 0, 0, time.UTC),
	}

	expected := "q:\"foo BAR\" name:\"yo/ba:z.JPG\" lat:1.500000 lng:-10.333330 chroma:1 diff:424242 year:2002 before:2019-01-15 private:true"
	expectedAll := "q:\"foo BAR\" name:\"yo/ba:z.JPG\" lat:1.500000 lng:-10.333330 chroma:1 diff:424242 year:2002 before:2019-01-15 private:true count:100 order:name"

	t.Run("value", func(t *testing.T) {
		result := Serialize(form, false)
		assert.IsType(t, expected, result)
		assert.Equal(t, expected, result)
	})

	t.Run("pointer", func(t *testing.T) {
		result := Serialize(&form, false)
		assert.IsType(t, expected, result)
		assert.Equal(t, expected, result)
	})

	t.Run("all value", func(t *testing.T) {
		result := Serialize(form, true)
		assert.IsType(t, expectedAll, result)
		assert.Equal(t, expectedAll, result)
	})

	t.Run("all pointer", func(t *testing.T) {
		result := Serialize(&form, true)
		assert.IsType(t, expectedAll, result)
		assert.Equal(t, expectedAll, result)
	})

	t.Run("invalid argument", func(t *testing.T) {
		result := Serialize("string", true)
		assert.Equal(t, "", result)
	})
}

func TestUnserialize(t *testing.T) {
	form := &TestForm{}

	serialized := "q:\"foo BAR\" name:\"yo/ba:z.JPG\" lat:1.500000 lng:-10.333330 chroma:1 diff:424242 year:2002 before:2019-01-15 private:true"

	if err := Unserialize(form, serialized); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 0, form.Count)
}
