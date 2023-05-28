package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/sortby"
	"github.com/photoprism/photoprism/pkg/txt"
)

func TestNewAlbum(t *testing.T) {
	t.Run("name Christmas 2018", func(t *testing.T) {
		album := NewAlbum("Christmas 2018", AlbumManual)
		assert.Equal(t, "Christmas 2018", album.AlbumTitle)
		assert.Equal(t, "christmas-2018", album.AlbumSlug)
	})
	t.Run("name empty", func(t *testing.T) {
		album := NewAlbum("", AlbumManual)

		defaultName := time.Now().Format("January 2006")
		defaultSlug := txt.Slug(defaultName)

		assert.Equal(t, defaultName, album.AlbumTitle)
		assert.Equal(t, defaultSlug, album.AlbumSlug)
	})
	t.Run("type empty", func(t *testing.T) {
		album := NewAlbum("Christmas 2018", "")
		assert.Equal(t, "Christmas 2018", album.AlbumTitle)
		assert.Equal(t, "christmas-2018", album.AlbumSlug)
	})
}

func TestAlbum_SetName(t *testing.T) {
	t.Run("valid name", func(t *testing.T) {
		album := NewAlbum("initial name", AlbumManual)
		assert.Equal(t, "initial name", album.AlbumTitle)
		assert.Equal(t, "initial-name", album.AlbumSlug)
		album.SetTitle("New Album \"Name\"")
		assert.Equal(t, "New Album “Name“", album.AlbumTitle)
		assert.Equal(t, "new-album-name", album.AlbumSlug)
	})
	t.Run("empty name", func(t *testing.T) {
		album := NewAlbum("initial name", AlbumManual)
		assert.Equal(t, "initial name", album.AlbumTitle)
		assert.Equal(t, "initial-name", album.AlbumSlug)

		album.SetTitle("")
		expected := album.CreatedAt.Format("January 2006")
		assert.Equal(t, expected, album.AlbumTitle)
		assert.Equal(t, txt.Slug(expected), album.AlbumSlug)
	})
	t.Run("long name", func(t *testing.T) {
		longName := `A value in decimal degrees to a precision of 4 decimal places is precise to 11.132 meters at the 
equator. A value in decimal degrees to 5 decimal places is precise to 1.1132 meter at the equator. Elevation also 
introduces a small error. At 6,378 m elevation, the radius and surface distance is increased by 0.001 or 0.1%. 
Because the earth is not flat, the precision of the longitude part of the coordinates increases 
the further from the equator you get. The precision of the latitude part does not increase so much, 
more strictly however, a meridian arc length per 1 second depends on the latitude at the point in question. 
The discrepancy of 1 second meridian arc length between equator and pole is about 0.3 metres because the earth 
is an oblate spheroid.`
		expected := txt.Shorten(longName, txt.ClipDefault, txt.Ellipsis)
		slugExpected := txt.Clip(longName, txt.ClipSlug)
		album := NewAlbum(longName, AlbumManual)
		assert.Equal(t, expected, album.AlbumTitle)
		assert.Contains(t, album.AlbumSlug, txt.Slug(slugExpected))
	})
}

func TestAlbum_UpdateSlug(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		album := NewMonthAlbum("Foo ", "foo", 2002, 11)

		assert.Equal(t, "Foo", album.AlbumTitle)
		assert.Equal(t, "foo", album.AlbumSlug)
		assert.Equal(t, "", album.AlbumDescription)
		assert.Equal(t, 2002, album.AlbumYear)
		assert.Equal(t, 11, album.AlbumMonth)

		if err := album.Create(); err != nil {
			t.Fatal(err)
		}

		if err := album.UpdateTitleAndLocation("November / 2002", "", "", "", "november-2002"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "November / 2002", album.AlbumTitle)
		assert.Equal(t, "november-2002", album.AlbumSlug)
		assert.Equal(t, "", album.AlbumDescription)
		assert.Equal(t, 2002, album.AlbumYear)
		assert.Equal(t, 11, album.AlbumMonth)

		if err := album.DeletePermanently(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAlbum_UpdateState(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		album := NewAlbum("Any State", AlbumState)

		assert.Equal(t, "Any State", album.AlbumTitle)
		assert.Equal(t, "any-state", album.AlbumSlug)

		if err := album.Create(); err != nil {
			t.Fatal(err)
		}

		if err := album.UpdateTitleAndState("Alberta", "canada-alberta", "Alberta", "ca"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Alberta", album.AlbumTitle)
		assert.Equal(t, "", album.AlbumDescription)
		assert.Equal(t, "Canada", album.AlbumLocation)
		assert.Equal(t, "Alberta", album.AlbumState)
		assert.Equal(t, "ca", album.AlbumCountry)

		if err := album.DeletePermanently(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAlbum_SaveForm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		album := NewAlbum("Old Name", AlbumManual)

		assert.Equal(t, "Old Name", album.AlbumTitle)
		assert.Equal(t, "old-name", album.AlbumSlug)

		album2 := Album{ID: 123, AlbumTitle: "New name", AlbumDescription: "new description", AlbumCategory: "family"}

		albumForm, err := form.NewAlbum(album2)

		if err != nil {
			t.Fatal(err)
		}

		err = album.SaveForm(albumForm)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "New name", album.AlbumTitle)
		assert.Equal(t, "new description", album.AlbumDescription)
		assert.Equal(t, "Family", album.AlbumCategory)

	})
}

func TestAddPhotoToAlbums(t *testing.T) {
	t.Run("success one album", func(t *testing.T) {
		err := AddPhotoToAlbums("pt9jtxrexxvl0yh0", []string{"at6axuzitogaaiax"})

		if err != nil {
			t.Fatal(err)
		}

		a := Album{AlbumUID: "at6axuzitogaaiax"}

		if found := a.Find(); found == nil {
			t.Fatal("should find album")
		}

		var entries PhotoAlbums

		if err = Db().Where("album_uid = ? AND photo_uid = ?", "at6axuzitogaaiax", "pt9jtxrexxvl0yh0").Find(&entries).Error; err != nil {
			t.Fatal(err)
		}

		if len(entries) < 1 {
			t.Error("at least one album entry expected")
		}

		// t.Logf("photo album entries: %+v", entries)
	})

	t.Run("empty photo", func(t *testing.T) {
		err := AddPhotoToAlbums("", []string{"at6axuzitogaaiax"})

		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("invalid photo uid", func(t *testing.T) {
		assert.Error(t, AddPhotoToAlbums("xxx", []string{"at6axuzitogaaiax"}))
	})

	t.Run("success two album", func(t *testing.T) {
		err := AddPhotoToAlbums("pt9jtxrexxvl0yh0", []string{"at6axuzitogaaiax", ""})

		if err != nil {
			t.Fatal(err)
		}

		a := Album{AlbumUID: "at6axuzitogaaiax"}

		if found := a.Find(); found == nil {
			t.Fatal("should find album")
		}

		var entries PhotoAlbums

		if err = Db().Where("album_uid = ? AND photo_uid = ?", "at6axuzitogaaiax", "pt9jtxrexxvl0yh0").Find(&entries).Error; err != nil {
			t.Fatal(err)
		}

		if len(entries) < 1 {
			t.Error("at least one album entry expected")
		}

		// t.Logf("photo album entries: %+v", entries)
	})
}

func TestNewFolderAlbum(t *testing.T) {
	t.Run("name Christmas 2018", func(t *testing.T) {
		album := NewFolderAlbum("Dogs", "dogs", "label:dog")
		assert.Equal(t, "Dogs", album.AlbumTitle)
		assert.Equal(t, "dogs", album.AlbumSlug)
		assert.Equal(t, AlbumFolder, album.AlbumType)
		assert.Equal(t, sortby.Added, album.AlbumOrder)
		assert.Equal(t, "label:dog", album.AlbumFilter)
	})
	t.Run("title empty", func(t *testing.T) {
		album := NewFolderAlbum("", "dogs", "label:dog")
		assert.Nil(t, album)
	})
}

func TestNewMomentsAlbum(t *testing.T) {
	t.Run("name Christmas 2018", func(t *testing.T) {
		album := NewMomentsAlbum("Dogs", "dogs", "label:dog")
		assert.Equal(t, "Dogs", album.AlbumTitle)
		assert.Equal(t, "dogs", album.AlbumSlug)
		assert.Equal(t, AlbumMoment, album.AlbumType)
		assert.Equal(t, sortby.Oldest, album.AlbumOrder)
		assert.Equal(t, "label:dog", album.AlbumFilter)
	})
	t.Run("title empty", func(t *testing.T) {
		album := NewMomentsAlbum("", "dogs", "label:dog")
		assert.Nil(t, album)
	})
}

func TestNewStateAlbum(t *testing.T) {
	t.Run("name Christmas 2018", func(t *testing.T) {
		album := NewStateAlbum("Dogs", "dogs", "label:dog")
		assert.Equal(t, "Dogs", album.AlbumTitle)
		assert.Equal(t, "dogs", album.AlbumSlug)
		assert.Equal(t, AlbumState, album.AlbumType)
		assert.Equal(t, sortby.Newest, album.AlbumOrder)
		assert.Equal(t, "label:dog", album.AlbumFilter)
	})
	t.Run("title empty", func(t *testing.T) {
		album := NewStateAlbum("", "dogs", "label:dog")
		assert.Nil(t, album)
	})
}

func TestNewMonthAlbum(t *testing.T) {
	t.Run("name Christmas 2018", func(t *testing.T) {
		album := NewMonthAlbum("Dogs", "dogs", 2020, 7)
		assert.Equal(t, "Dogs", album.AlbumTitle)
		assert.Equal(t, "dogs", album.AlbumSlug)
		assert.Equal(t, AlbumMonth, album.AlbumType)
		assert.Equal(t, sortby.Oldest, album.AlbumOrder)
		assert.Equal(t, "public:true year:2020 month:7", album.AlbumFilter)
		assert.Equal(t, 7, album.AlbumMonth)
		assert.Equal(t, 2020, album.AlbumYear)
	})
	t.Run("title empty", func(t *testing.T) {
		album := NewMonthAlbum("", "dogs", 2020, 8)
		assert.Nil(t, album)
	})
}

func TestFindAlbumBySlug(t *testing.T) {
	t.Run("1 result", func(t *testing.T) {
		result := FindAlbumBySlug("holiday-2030", AlbumManual)

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "Holiday 2030", result.AlbumTitle)
		assert.Equal(t, "holiday-2030", result.AlbumSlug)
	})
	t.Run("state album", func(t *testing.T) {
		result := FindAlbumBySlug("california-usa", AlbumState)

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "California / USA", result.AlbumTitle)
		assert.Equal(t, "california-usa", result.AlbumSlug)
	})
	t.Run("no result", func(t *testing.T) {
		result := FindAlbumBySlug("holiday-2030", AlbumMonth)

		if result != nil {
			t.Fatal("album should be nil")
		}
	})
}

func TestAlbum_String(t *testing.T) {
	t.Run("return slug", func(t *testing.T) {
		album := Album{
			AlbumUID:   "abc123",
			AlbumSlug:  "test-slug",
			AlbumType:  AlbumManual,
			AlbumTitle: "Test Title",
		}
		assert.Equal(t, "test-slug", album.String())
	})
	t.Run("return title", func(t *testing.T) {
		album := Album{
			AlbumUID:   "abc123",
			AlbumSlug:  "",
			AlbumType:  AlbumManual,
			AlbumTitle: "Test Title",
		}
		assert.Contains(t, album.String(), "Test Title")
	})
	t.Run("return uid", func(t *testing.T) {
		album := Album{
			AlbumUID:   "abc123",
			AlbumSlug:  "",
			AlbumType:  AlbumManual,
			AlbumTitle: "",
		}
		assert.Equal(t, "abc123", album.String())
	})
	t.Run("return unknown", func(t *testing.T) {
		album := Album{
			AlbumUID:   "",
			AlbumSlug:  "",
			AlbumType:  AlbumManual,
			AlbumTitle: "",
		}
		assert.Equal(t, "[unknown album]", album.String())
	})
}

func TestAlbum_IsMoment(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		album := Album{
			AlbumUID:   "abc123",
			AlbumSlug:  "test-slug",
			AlbumType:  AlbumManual,
			AlbumTitle: "Test Title",
		}
		assert.False(t, album.IsMoment())
	})
	t.Run("true", func(t *testing.T) {
		album := Album{
			AlbumUID:   "abc123",
			AlbumSlug:  "test-slug",
			AlbumType:  AlbumMoment,
			AlbumTitle: "Test Title",
		}
		assert.True(t, album.IsMoment())
	})
}

func TestAlbum_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := NewAlbum("Test Title", AlbumManual)
		if err := album.Save(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "test-title", album.AlbumSlug)

		if err := album.Update("AlbumSlug", "new-slug"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "new-slug", album.AlbumSlug)

		if err := album.DeletePermanently(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAlbum_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := NewStateAlbum("Dogs", "dogs", "label:dog")

		initialDate := album.UpdatedAt

		err := album.Save()

		if err != nil {
			t.Fatal(err)
		}
		afterDate := album.UpdatedAt
		t.Log(initialDate)
		t.Log(afterDate)
		//TODO Why does it fail?
		//assert.True(t, afterDate.After(initialDate))
	})
}

func TestAlbum_Create(t *testing.T) {
	t.Run("Album", func(t *testing.T) {
		album := Album{
			AlbumType: AlbumManual,
		}

		err := album.Create()

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Moment", func(t *testing.T) {
		album := Album{
			AlbumType: AlbumMoment,
		}

		err := album.Create()

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Month", func(t *testing.T) {
		album := Album{
			AlbumType: AlbumMonth,
		}

		err := album.Create()

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Folder", func(t *testing.T) {
		album := Album{
			AlbumType: AlbumFolder,
		}

		err := album.Create()

		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestAlbum_Title(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := Album{
			AlbumUID:   "abc123",
			AlbumSlug:  "test-slug",
			AlbumType:  AlbumManual,
			AlbumTitle: "Test Title",
		}
		assert.Equal(t, "Test Title", album.Title())
	})
}

func TestAlbum_Links(t *testing.T) {
	t.Run("OneResult", func(t *testing.T) {
		album := AlbumFixtures.Get("christmas2030")
		links := album.Links()
		assert.Equal(t, "4jxf3jfn2k", links[0].LinkToken)
	})
}

func TestAlbum_AddPhotos(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := Album{
			ID:         1000000,
			AlbumUID:   "at9lxuqxpogaaba7",
			AlbumSlug:  "test-slug",
			AlbumType:  AlbumManual,
			AlbumTitle: "Test Title",
		}
		added := album.AddPhotos([]string{"pt9jtdre2lvl0yh7", "pt9jtdre2lvl0yh8"})
		assert.Equal(t, 2, len(added))
	})
}

func TestAlbum_RemovePhotos(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		album := Album{
			ID:         1000000,
			AlbumUID:   "at9lxuqxpogaaba7",
			AlbumSlug:  "test-slug",
			AlbumType:  AlbumManual,
			AlbumTitle: "Test Title",
		}
		removed := album.RemovePhotos([]string{"pt9jtdre2lvl0yh7", "pt9jtdre2lvl0yh8"})
		assert.Equal(t, 2, len(removed))
	})
}

func TestAlbum_Find(t *testing.T) {
	t.Run("existing album", func(t *testing.T) {
		a := Album{AlbumUID: "at6axuzitogaaiax"}

		if found := a.Find(); found == nil {
			t.Fatal("should find album")
		}
	})
	t.Run("invalid id", func(t *testing.T) {
		a := Album{AlbumUID: "xx"}

		if found := a.Find(); found != nil {
			t.Fatal("should not find album")
		}
	})
	t.Run("album not existing", func(t *testing.T) {
		a := Album{AlbumUID: "at6axuzitogaaxxx"}

		if found := a.Find(); found != nil {
			t.Fatal("should not find album")
		}
	})
}

func TestAlbum_UpdateFolder(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		a := Album{ID: 99999, AlbumUID: "at6axuzitogaaxxx"}
		assert.Empty(t, a.AlbumPath)
		assert.Empty(t, a.AlbumFilter)
		if err := a.UpdateFolder("2222/07", "month:07"); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "2222/07", a.AlbumPath)
		assert.Equal(t, "month:07", a.AlbumFilter)
	})
	t.Run("EmptyPath", func(t *testing.T) {
		a := Album{ID: 99999, AlbumUID: "at6axuzitogaaxxy"}
		assert.Empty(t, a.AlbumPath)
		assert.Empty(t, a.AlbumFilter)
		err := a.UpdateFolder("", "month:07")
		assert.Error(t, err)
	})
	t.Run("EmptyFilter", func(t *testing.T) {
		a := Album{ID: 99999, AlbumUID: "at6axuzitogaaxxy"}
		assert.Empty(t, a.AlbumPath)
		assert.Empty(t, a.AlbumFilter)
		err := a.UpdateFolder("2222/07", "")
		assert.Error(t, err)
	})
}

func TestFindFolderAlbum(t *testing.T) {
	/*t.Run("1 result", func(t *testing.T) {
		album := FindFolderAlbum("2023/04")

		if album == nil {
			t.Fatal("expected to find an album")
		}

		assert.Equal(t, "April 2023", album.AlbumTitle)
		assert.Equal(t, "2023-04", album.AlbumSlug)
	})*/
	t.Run("no result because slug empty", func(t *testing.T) {
		album := FindFolderAlbum("")

		if album != nil {
			t.Fatal("album should be nil")
		}
	})
	t.Run("no result because not found slug", func(t *testing.T) {
		album := FindFolderAlbum("3000/04")

		if album != nil {
			t.Fatal("album should be nil")
		}
	})
}
