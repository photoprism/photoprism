package form

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("Value", func(t *testing.T) {
		result := Serialize(form, false)
		assert.IsType(t, expected, result)
		assert.Equal(t, expected, result)
	})
	t.Run("Pointer", func(t *testing.T) {
		result := Serialize(&form, false)
		assert.IsType(t, expected, result)
		assert.Equal(t, expected, result)
	})
	t.Run("AllValue", func(t *testing.T) {
		result := Serialize(form, true)
		assert.IsType(t, expectedAll, result)
		assert.Equal(t, expectedAll, result)
	})
	t.Run("AllPointer", func(t *testing.T) {
		result := Serialize(&form, true)
		assert.IsType(t, expectedAll, result)
		assert.Equal(t, expectedAll, result)
	})
	t.Run("InvalidArgument", func(t *testing.T) {
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

func TestUnserializeUnsignedValidation(t *testing.T) {
	t.Run("RejectNegativeUnsigned", func(t *testing.T) {
		form := &TestForm{}

		if err := Unserialize(form, "dist:-1"); err == nil {
			t.Fatal("expected error for negative unsigned value")
		}
	})
	t.Run("RejectOverflowUnsigned", func(t *testing.T) {
		form := &TestForm{}

		// uint32 max + 1 should fail for Diff.
		if err := Unserialize(form, "diff:4294967296"); err == nil {
			t.Fatal("expected overflow error for uint32 field")
		}
	})
}

func TestUnserializeFilter(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		f := &SearchPhotos{}
		noMatch, err := UnserializeFilter(f, "label:cat public:true")
		require.NoError(t, err)
		assert.False(t, noMatch)
		assert.Equal(t, "cat", f.Label)
		assert.True(t, f.Public)
	})
	t.Run("NoMatch", func(t *testing.T) {
		for _, q := range []string{"\xff", "\x00 ", "label:\xff", "public:true path:\"\xc4\xd6\"", "before:\xff"} {
			noMatch, err := UnserializeFilter(&SearchPhotos{}, q)
			require.NoError(t, err, "%q", q)
			assert.True(t, noMatch, "%q", q)
		}
	})
	t.Run("InvalidFilter", func(t *testing.T) {
		_, err := UnserializeFilter(&SearchPhotos{}, "foo:bar")
		assert.Error(t, err)
	})
}

func TestFilterValue(t *testing.T) {
	t.Run("AsWritten", func(t *testing.T) {
		assert.Equal(t, "2020*", FilterValue("path:2020* public:true", "path"))
		assert.Equal(t, "20%20", FilterValue("public:true path:20%20", "path"))
		assert.Equal(t, "a|b", FilterValue("path:a|b", "path"))
		assert.Equal(t, "2020/My Trip", FilterValue("path:\"2020/My Trip\" public:true", "Path"))
		assert.Equal(t, "true", FilterValue("path:a public:true", "public"))
	})
	t.Run("LastValue", func(t *testing.T) {
		assert.Equal(t, "b", FilterValue("path:a path:b", "path"))
	})
	t.Run("OpenQuote", func(t *testing.T) {
		assert.Equal(t, "", FilterValue("path:\"My Trip", "path"))
		assert.Equal(t, "a", FilterValue("path:a public:\"tr", "path"))
	})
	t.Run("OtherWhitespace", func(t *testing.T) {
		// A value that ends at other whitespace may be part of a longer value.
		assert.Equal(t, "", FilterValue("path:x\u3000 public:true", "path"))
		assert.Equal(t, "", FilterValue("path:x\u00a0| public:true", "path"))
		assert.Equal(t, "", FilterValue("path:x\t${y}", "path"))
		assert.Equal(t, "", FilterValue("path:x\u3000", "path"))
		assert.Equal(t, "x", FilterValue("path:x  public:true", "path"))
		assert.Equal(t, "x", FilterValue("public:true\tpath:x", "path"))
	})
	t.Run("SerializedPath", func(t *testing.T) {
		// A path written by Serialize is read back as it was.
		for _, p := range []string{"2020/04", "2020/My Trip", "a|b*", "20%20", "x:y", "a\\b", "Tokyo\u3000Trip", "a\u00a0b", "a\tb", "Grüße/日本", "(a)[b]-c+'d'"} {
			assert.Equal(t, p, FilterValue(Serialize(SearchPhotos{Path: p, Public: true}, false), "path"), "%q", p)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		assert.Equal(t, "", FilterValue("", "path"))
		assert.Equal(t, "", FilterValue("public:true path", "path"))
		assert.Equal(t, "", FilterValue("folder:a", "path"))
	})
}
