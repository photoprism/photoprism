package entity

import (
	"strings"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/pkg/txt"
)

func TestUpdateAlbum(t *testing.T) {
	t.Run("InvalidUID", func(t *testing.T) {
		err := UpdateAlbum("xxx", Map{"album_title": "New Title", "album_slug": "new-slug"})

		assert.Error(t, err)
	})
}

func TestAddPhotoToAlbums(t *testing.T) {
	t.Run("SuccessOneAlbum", func(t *testing.T) {
		err := AddPhotoToAlbums("ps6sg6bexxvl0yh0", []string{"as6sg6bitoga0004"})

		if err != nil {
			t.Fatal(err)
		}

		a := Album{AlbumUID: "as6sg6bitoga0004"}

		if found := a.Find(); found == nil {
			t.Fatal("should find album")
		}

		var entries PhotoAlbums

		if err = Db().Where("album_uid = ? AND photo_uid = ?", "as6sg6bitoga0004", "ps6sg6bexxvl0yh0").Find(&entries).Error; err != nil {
			t.Fatal(err)
		}

		if len(entries) < 1 {
			t.Error("at least one album entry expected")
		}

		var album Album
		if err = Db().Where("album_uid = ?", "as6sg6bitoga0004").Find(
			&album,
		).Error; err != nil {
			t.Fatal(err)
		}

		photo_updatedAt := strings.Split(entries[0].UpdatedAt.String(), ".")[0]
		album_updatedAt := strings.Split(album.UpdatedAt.String(), ".")[0]

		assert.Truef(
			t, photo_updatedAt <= album_updatedAt,
			"Expected the UpdatedAt field of an album to be updated when"+
				" new photos are added",
		)
	},
	)

	t.Run("EmptyPhoto", func(t *testing.T) {
		err := AddPhotoToAlbums("", []string{"as6sg6bitoga0004"})

		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("InvalidPhotoUid", func(t *testing.T) {
		assert.Error(t, AddPhotoToAlbums("xxx", []string{"as6sg6bitoga0004"}))
	})

	t.Run("SuccessTwoAlbums", func(t *testing.T) {
		err := AddPhotoToAlbums("ps6sg6bexxvl0yh0", []string{"as6sg6bitoga0004", ""})

		if err != nil {
			t.Fatal(err)
		}

		a := Album{AlbumUID: "as6sg6bitoga0004"}

		if found := a.Find(); found == nil {
			t.Fatal("should find album")
		}

		var entries PhotoAlbums

		if err = Db().Where("album_uid = ? AND photo_uid = ?", "as6sg6bitoga0004", "ps6sg6bexxvl0yh0").Find(&entries).Error; err != nil {
			t.Fatal(err)
		}

		if len(entries) < 1 {
			t.Error("at least one album entry expected")
		}

		var album Album
		if err = Db().Where("album_uid = ?", "as6sg6bitoga0004").Find(
			&album,
		).Error; err != nil {
			t.Fatal(err)
		}

		photo_updatedAt := strings.Split(entries[0].UpdatedAt.String(), ".")[0]
		album_updatedAt := strings.Split(album.UpdatedAt.String(), ".")[0]

		assert.Truef(
			t, photo_updatedAt <= album_updatedAt,
			"Expected the UpdatedAt field of an album to be updated when"+
				" new photos are added",
		)
	})
}

func TestAddPhotoToUserAlbums(t *testing.T) {
	t.Run("AddToExistingAlbum", func(t *testing.T) {
		err := AddPhotoToUserAlbums("ps6sg6bexxvl0yh0", []string{"as6sg6bitoga0004"}, "uqxetse3cy5eo9z2")

		if err != nil {
			t.Fatal(err)
		}

		a := Album{AlbumUID: "as6sg6bitoga0004"}

		if found := a.Find(); found == nil {
			t.Fatal("should find album")
		}

		var entries PhotoAlbums

		if err = Db().Where("album_uid = ? AND photo_uid = ?", "as6sg6bitoga0004", "ps6sg6bexxvl0yh0").Find(&entries).Error; err != nil {
			t.Fatal(err)
		}

		if len(entries) < 1 {
			t.Error("at least one album entry expected")
		}

		var album Album
		if err = Db().Where("album_uid = ?", "as6sg6bitoga0004").Find(
			&album,
		).Error; err != nil {
			t.Fatal(err)
		}

		photo_updatedAt := strings.Split(entries[0].UpdatedAt.String(), ".")[0]
		album_updatedAt := strings.Split(album.UpdatedAt.String(), ".")[0]

		assert.Truef(
			t, photo_updatedAt <= album_updatedAt,
			"Expected the UpdatedAt field of an album to be updated when"+
				" new photos are added",
		)
	})
	t.Run("CreateNewAlbum", func(t *testing.T) {
		assert.Nil(t, FindAlbumByAttr([]string{"yyy"}, []string{}, AlbumManual))

		assert.NoError(t, AddPhotoToUserAlbums("ps6sg6bexxvl0yh0", []string{"yyy"}, "uqxetse3cy5eo9z2"))

		assert.NotNil(t, FindAlbumByAttr([]string{"yyy"}, []string{}, AlbumManual))
	})
}

func TestNewAlbum(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := NewAlbum("Christmas 2018", AlbumManual)
		assert.Equal(t, "Christmas 2018", album.AlbumTitle)
		assert.Equal(t, "christmas-2018", album.AlbumSlug)
	})
	t.Run("NameEmpty", func(t *testing.T) {
		album := NewAlbum("", AlbumManual)

		defaultName := time.Now().Format("January 2006")
		defaultSlug := txt.Slug(defaultName)

		assert.Equal(t, defaultName, album.AlbumTitle)
		assert.Equal(t, defaultSlug, album.AlbumSlug)
	})
	t.Run("TypeEmpty", func(t *testing.T) {
		album := NewAlbum("Christmas 2018", "")
		assert.Equal(t, "Christmas 2018", album.AlbumTitle)
		assert.Equal(t, "christmas-2018", album.AlbumSlug)
	})
}

func TestNewFolderAlbum(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := NewFolderAlbum("Dogs", "dogs", "label:dog")
		assert.Equal(t, "Dogs", album.AlbumTitle)
		assert.Equal(t, "dogs", album.AlbumSlug)
		assert.Equal(t, AlbumFolder, album.AlbumType)
		assert.Equal(t, sortby.Added, album.AlbumOrder)
		assert.Equal(t, "label:dog", album.AlbumFilter)
	})
	t.Run("TitleEmpty", func(t *testing.T) {
		album := NewFolderAlbum("", "dogs", "label:dog")
		assert.Nil(t, album)
	})
}

func TestNewMomentsAlbum(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := NewMomentsAlbum("Dogs", "dogs", "label:dog")
		assert.Equal(t, "Dogs", album.AlbumTitle)
		assert.Equal(t, "dogs", album.AlbumSlug)
		assert.Equal(t, AlbumMoment, album.AlbumType)
		assert.Equal(t, sortby.Oldest, album.AlbumOrder)
		assert.Equal(t, "label:dog", album.AlbumFilter)
	})
	t.Run("TitleEmpty", func(t *testing.T) {
		album := NewMomentsAlbum("", "dogs", "label:dog")
		assert.Nil(t, album)
	})
}

func TestNewStateAlbum(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := NewStateAlbum("Dogs", "dogs", "label:dog")
		assert.Equal(t, "Dogs", album.AlbumTitle)
		assert.Equal(t, "dogs", album.AlbumSlug)
		assert.Equal(t, AlbumState, album.AlbumType)
		assert.Equal(t, sortby.Newest, album.AlbumOrder)
		assert.Equal(t, "label:dog", album.AlbumFilter)
	})
	t.Run("TitleEmpty", func(t *testing.T) {
		album := NewStateAlbum("", "dogs", "label:dog")
		assert.Nil(t, album)
	})
}

func TestNewMonthAlbum(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := NewMonthAlbum("Dogs", "dogs", 2020, 7)
		assert.Equal(t, "Dogs", album.AlbumTitle)
		assert.Equal(t, "dogs", album.AlbumSlug)
		assert.Equal(t, AlbumMonth, album.AlbumType)
		assert.Equal(t, sortby.Oldest, album.AlbumOrder)
		assert.Equal(t, "public:true year:2020 month:7", album.AlbumFilter)
		assert.Equal(t, 7, album.AlbumMonth)
		assert.Equal(t, 2020, album.AlbumYear)
	})
	t.Run("TitleEmpty", func(t *testing.T) {
		album := NewMonthAlbum("", "dogs", 2020, 8)
		assert.Nil(t, album)
	})
}

func TestFindMonthAlbum(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result := FindMonthAlbum(2021, 9)

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "September 2021", result.AlbumTitle)
	})
	t.Run("InvalidMonth", func(t *testing.T) {
		result := FindMonthAlbum(2021, 19)

		assert.Nil(t, result)
	})
	t.Run("NoResult", func(t *testing.T) {
		result := FindMonthAlbum(2021, 1)

		assert.Nil(t, result)
	})
}

func TestFindAlbumBySlug(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result := FindAlbumBySlug("holiday-2030", AlbumManual)

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "Holiday 2030", result.AlbumTitle)
		assert.Equal(t, "holiday-2030", result.AlbumSlug)
	})
	t.Run("FindState", func(t *testing.T) {
		result := FindAlbumBySlug("california-usa", AlbumState)

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "California / USA", result.AlbumTitle)
		assert.Equal(t, "california-usa", result.AlbumSlug)
	})
	t.Run("NoResult", func(t *testing.T) {
		result := FindAlbumBySlug("holiday-2030", AlbumMonth)

		if result != nil {
			t.Fatal("album should be nil")
		}
	})
	t.Run("EmptySlug", func(t *testing.T) {
		result := FindAlbumBySlug("", AlbumMonth)

		if result != nil {
			t.Fatal("album should be nil")
		}
	})
}

func TestFindAlbumByAttr(t *testing.T) {
	t.Run("FindByFilter", func(t *testing.T) {
		result := FindAlbumByAttr([]string{}, []string{"path:\"1990/04\" public:true"}, AlbumFolder)

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "April 1990", result.AlbumTitle)
	})
	t.Run("FindBySlug", func(t *testing.T) {
		result := FindAlbumByAttr([]string{"holiday-2030"}, []string{}, AlbumManual)

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "Holiday 2030", result.AlbumTitle)
	})
	t.Run("NoQuery", func(t *testing.T) {
		result := FindAlbumByAttr([]string{}, []string{}, AlbumManual)

		assert.Nil(t, result)
	})
	t.Run("NoResult", func(t *testing.T) {
		result := FindAlbumByAttr([]string{"xxx"}, []string{"xxx"}, AlbumManual)

		assert.Nil(t, result)
	})
}

func TestFindFolderAlbum(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := FindFolderAlbum("1990/04")

		if album == nil {
			t.Fatal("expected to find an album")
		}

		assert.Equal(t, "April 1990", album.AlbumTitle)
		assert.Equal(t, "april-1990", album.AlbumSlug)
		assert.False(t, album.IsDefault())
		assert.False(t, album.IsState())
	})
	t.Run("EmptySlug", func(t *testing.T) {
		album := FindFolderAlbum("")

		if album != nil {
			t.Fatal("album should be nil")
		}
	})
	t.Run("InvalidSlug", func(t *testing.T) {
		album := FindFolderAlbum("3000/04")

		if album != nil {
			t.Fatal("album should be nil")
		}
	})
}

func TestFindAlbum(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := AlbumFixtures.Get("christmas2030")
		result := FindAlbum(album)

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "Christmas 2030", result.AlbumTitle)
		assert.True(t, result.IsDefault())
	})
	t.Run("Success", func(t *testing.T) {
		album := Album{AlbumSlug: "april-1990", AlbumType: AlbumFolder}
		result := FindAlbum(album)

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "April 1990", result.AlbumTitle)
	})
	t.Run("Success", func(t *testing.T) {
		album := Album{AlbumSlug: "april-1990", AlbumType: AlbumFolder, AlbumFilter: "1990/04"}
		result := FindAlbum(album)

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "April 1990", result.AlbumTitle)
	})
	t.Run("Success", func(t *testing.T) {
		album := Album{AlbumSlug: "berlin-2019", AlbumType: AlbumManual}
		result := FindAlbum(album)

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "Berlin 2019", result.AlbumTitle)
	})
	t.Run("NoResult", func(t *testing.T) {
		album := Album{AlbumSlug: "berlin-2019", AlbumType: AlbumManual, CreatedBy: "xxx"}
		result := FindAlbum(album)

		assert.Nil(t, result)
	})
	t.Run("NoResult", func(t *testing.T) {
		album := Album{AlbumSlug: "xxx-xxx", AlbumType: AlbumFolder}
		result := FindAlbum(album)

		assert.Nil(t, result)
	})
}

func TestAlbum_Find(t *testing.T) {
	t.Run("ExistingAlbum", func(t *testing.T) {
		a := Album{AlbumUID: "as6sg6bitoga0004"}

		if found := a.Find(); found == nil {
			t.Fatal("should find album")
		}
	})
	t.Run("InvalidId", func(t *testing.T) {
		a := Album{AlbumUID: "xx"}

		if found := a.Find(); found != nil {
			t.Fatal("should not find album")
		}
	})
	t.Run("AlbumNotExisting", func(t *testing.T) {
		a := Album{AlbumUID: "as6sg6bitogaaxxx"}

		if found := a.Find(); found != nil {
			t.Fatal("should not find album")
		}
	})
}

func TestAlbum_String(t *testing.T) {
	t.Run("ReturnSlug", func(t *testing.T) {
		album := Album{
			AlbumUID:   "abc123",
			AlbumSlug:  "test-slug",
			AlbumType:  AlbumManual,
			AlbumTitle: "Test Title",
		}
		assert.Equal(t, "test-slug", album.String())
	})
	t.Run("ReturnTitle", func(t *testing.T) {
		album := Album{
			AlbumUID:   "abc123",
			AlbumSlug:  "",
			AlbumType:  AlbumManual,
			AlbumTitle: "Test Title",
		}
		assert.Contains(t, album.String(), "Test Title")
	})
	t.Run("ReturnUid", func(t *testing.T) {
		album := Album{
			AlbumUID:   "abc123",
			AlbumSlug:  "",
			AlbumType:  AlbumManual,
			AlbumTitle: "",
		}
		assert.Equal(t, "abc123", album.String())
	})
	t.Run("ReturnUnknown", func(t *testing.T) {
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

func TestAlbum_SetTitle(t *testing.T) {
	t.Run("ValidName", func(t *testing.T) {
		album := NewAlbum("initial name", AlbumManual)
		assert.Equal(t, "initial name", album.AlbumTitle)
		assert.Equal(t, "initial-name", album.AlbumSlug)
		album.SetTitle("New Album \"Name\"")
		assert.Equal(t, "New Album “Name“", album.AlbumTitle)
		assert.Equal(t, "new-album-name", album.AlbumSlug)
	})
	t.Run("EmptyName", func(t *testing.T) {
		album := NewAlbum("initial name", AlbumManual)
		assert.Equal(t, "initial name", album.AlbumTitle)
		assert.Equal(t, "initial-name", album.AlbumSlug)

		album.SetTitle("")
		expected := album.CreatedAt.Format("January 2006")
		assert.Equal(t, expected, album.AlbumTitle)
		assert.Equal(t, txt.Slug(expected), album.AlbumSlug)
	})
	t.Run("LongName", func(t *testing.T) {
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

func TestAlbum_SetLocation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := Album{}
		result := album.SetLocation("world", "Hessen", "de")

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "world", result.AlbumLocation)
		assert.Equal(t, "Hessen", result.AlbumState)
		assert.Equal(t, "de", result.AlbumCountry)
	})
	t.Run("Unknown", func(t *testing.T) {
		album := Album{}
		result := album.SetLocation("", "", "zz")

		if result == nil {
			t.Fatal("album should not be nil")
		}

		assert.Equal(t, "", result.AlbumLocation)
		assert.Equal(t, "", result.AlbumState)
		assert.Equal(t, "", result.AlbumCountry)
	})
}

func TestAlbum_UpdateTitleAndLocation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := Album{ID: 12345, AlbumUID: "as6sg6bxpogaakj6"}
		err := album.UpdateTitleAndLocation("My Picture Title", "world", "Hessen", "de", "test-slug")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "world", album.AlbumLocation)
		assert.Equal(t, "Hessen", album.AlbumState)
		assert.Equal(t, "de", album.AlbumCountry)
		assert.Equal(t, "My Picture Title", album.AlbumTitle)
		assert.Equal(t, "test-slug", album.AlbumSlug)
	})
	t.Run("SuccessMonthAlbum", func(t *testing.T) {
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
	t.Run("TitleMissing", func(t *testing.T) {
		album := Album{ID: 12345, AlbumUID: "as6sg6bxpogaakj6"}
		err := album.UpdateTitleAndLocation("", "world", "Hessen", "de", "test-slug")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", album.AlbumLocation)
		assert.Equal(t, "", album.AlbumState)
		assert.Equal(t, "", album.AlbumCountry)
		assert.Equal(t, "", album.AlbumTitle)
		assert.Equal(t, "", album.AlbumSlug)
	})
	t.Run("NoChange", func(t *testing.T) {
		album := Album{ID: 12345, AlbumUID: "as6sg6bxpogaakj6", AlbumSlug: "test-slug", AlbumState: "Hessen", AlbumCountry: "de"}
		err := album.UpdateTitleAndLocation("test title", "world", "Hessen", "de", "test-slug")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", album.AlbumLocation)
		assert.Equal(t, "Hessen", album.AlbumState)
		assert.Equal(t, "de", album.AlbumCountry)
		assert.Equal(t, "", album.AlbumTitle)
		assert.Equal(t, "test-slug", album.AlbumSlug)
	})
	t.Run("NoUID", func(t *testing.T) {
		album := Album{ID: 12345}
		err := album.UpdateTitleAndLocation("", "world", "Hessen", "de", "test-slug")

		assert.Error(t, err)
	})
}

func TestAlbum_UpdateTitleAndState(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
	t.Run("TitleMissing", func(t *testing.T) {
		album := NewAlbum("Any State", AlbumState)

		assert.Equal(t, "Any State", album.AlbumTitle)
		assert.Equal(t, "any-state", album.AlbumSlug)

		if err := album.Create(); err != nil {
			t.Fatal(err)
		}

		if err := album.UpdateTitleAndState("", "canada-alberta", "Alberta", "ca"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Any State", album.AlbumTitle)
		assert.Equal(t, "", album.AlbumDescription)
		assert.Equal(t, "", album.AlbumLocation)
		assert.Equal(t, "", album.AlbumState)
		assert.Equal(t, "zz", album.AlbumCountry)

		if err := album.DeletePermanently(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("NoChange", func(t *testing.T) {
		album := Album{ID: 12345, AlbumUID: "as6sg6bxpogaakj6", AlbumLocation: "Canada", AlbumTitle: "Alberta", AlbumSlug: "canada-alberta", AlbumState: "Alberta", AlbumCountry: "ca"}

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
	t.Run("NoUID", func(t *testing.T) {
		album := Album{}

		err := album.UpdateTitleAndState("", "canada-alberta", "Alberta", "ca")

		assert.Error(t, err)
	})
}

func TestAlbum_SaveForm(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
	t.Run("NoUID", func(t *testing.T) {
		album := Album{}

		err := album.Update("AlbumSlug", "new-slug")

		assert.Error(t, err)
	})
}

func TestAlbum_Updates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := NewAlbum("Test Title", AlbumManual)
		if err := album.Save(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "test-title", album.AlbumSlug)

		if err := album.Updates(Map{"album_title": "New Title", "album_slug": "new-slug"}); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "new-slug", album.AlbumSlug)

		if err := album.DeletePermanently(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("NoUID", func(t *testing.T) {
		album := Album{}

		err := album.Updates(Map{"album_title": "New Title", "album_slug": "new-slug"})

		assert.Error(t, err)
	})
}

func TestAlbum_UpdateFolder(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		a := Album{ID: 99999, AlbumUID: "as6sg6bitogaaxxx"}
		assert.Empty(t, a.AlbumPath)
		assert.Empty(t, a.AlbumFilter)
		if err := a.UpdateFolder("2222/07", "month:07"); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "2222/07", a.AlbumPath)
		assert.Equal(t, "month:07", a.AlbumFilter)
	})
	t.Run("NoChange", func(t *testing.T) {
		a := Album{ID: 99999, AlbumUID: "as6sg6bitogaaxxx", AlbumSlug: "2222-07", AlbumFilter: "month:07", AlbumPath: "2222/07"}

		if err := a.UpdateFolder("2222/07", "month:07"); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "2222/07", a.AlbumPath)
		assert.Equal(t, "month:07", a.AlbumFilter)
		assert.Equal(t, "2222-07", a.AlbumSlug)
	})
	t.Run("EmptyPath", func(t *testing.T) {
		a := Album{ID: 99999, AlbumUID: "as6sg6bitogaaxxy"}
		assert.Empty(t, a.AlbumPath)
		assert.Empty(t, a.AlbumFilter)
		err := a.UpdateFolder("", "month:07")
		assert.Error(t, err)
	})
	t.Run("EmptyFilter", func(t *testing.T) {
		a := Album{ID: 99999, AlbumUID: "as6sg6bitogaaxxy"}
		assert.Empty(t, a.AlbumPath)
		assert.Empty(t, a.AlbumFilter)
		err := a.UpdateFolder("2222/07", "")
		assert.Error(t, err)
	})
	t.Run("NoUID", func(t *testing.T) {
		a := Album{ID: 99999}
		assert.Empty(t, a.AlbumPath)
		assert.Empty(t, a.AlbumFilter)
		err := a.UpdateFolder("2222/07", "")
		assert.Error(t, err)
	})
}

func TestAlbum_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := AlbumFixtures.Get("christmas2030")

		initialDate := album.UpdatedAt

		err := album.Save()

		if err != nil {
			t.Fatal(err)
		}
		afterDate := album.UpdatedAt

		assert.True(t, afterDate.After(initialDate))
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

func TestAlbum_DeletePermanently(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := NewAlbum("Christmas 2018", AlbumManual)

		if err := album.Save(); err != nil {
			t.Fatal(err)
		}

		if found := album.Find(); found == nil {
			t.Fatal("should find album")
		}

		if err := album.DeletePermanently(); err != nil {
			t.Fatal(err)
		}

		if found := album.Find(); found != nil {
			t.Fatal("should not find album")
		}
	})
	t.Run("NoUID", func(t *testing.T) {
		album := Album{}

		err := album.DeletePermanently()

		assert.Error(t, err)
	})
}

func TestAlbum_DeleteRestore(t *testing.T) {
	t.Run("DeleteAndRestore", func(t *testing.T) {
		album := NewAlbum("Test Title", AlbumManual)

		if err := album.Save(); err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, album.DeletedAt)
		assert.False(t, album.Deleted())

		if err := album.Delete(); err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, album.DeletedAt)
		assert.True(t, album.Deleted())

		if err := album.Restore(); err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, album.DeletedAt)
	})
	t.Run("DeleteAlreadyDeleted", func(t *testing.T) {
		album := NewAlbum("Test Title", AlbumManual)

		if err := album.Save(); err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, album.DeletedAt)

		if err := album.Delete(); err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, album.DeletedAt)

		if err := album.Delete(); err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, album.DeletedAt)
	})
	t.Run("DeleteNoUID", func(t *testing.T) {
		album := Album{}

		err := album.Delete()

		assert.Error(t, err)
	})
	t.Run("RestoreNoUID", func(t *testing.T) {
		album := Album{}

		err := album.Restore()

		assert.Error(t, err)
	})
	t.Run("RestoreNotDeleted", func(t *testing.T) {
		album := NewAlbum("Test Title", AlbumManual)

		if err := album.Save(); err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, album.DeletedAt)
		assert.False(t, album.Deleted())

		if err := album.Restore(); err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, album.DeletedAt)
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

func TestAlbum_ZipName(t *testing.T) {
	t.Run("christmas-2030.zip", func(t *testing.T) {
		album := AlbumFixtures.Get("christmas2030")
		result := album.ZipName()

		assert.Equal(t, "christmas-2030.zip", result)
	})
	t.Run("photoprism-album-1234.zip", func(t *testing.T) {
		album := Album{AlbumSlug: "a", AlbumUID: "1234"}
		result := album.ZipName()

		assert.Equal(t, "photoprism-album-1234.zip", result)
	})
}

func TestAlbum_AddPhotos(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := Album{
			ID:         1000000,
			AlbumUID:   "as6sg6bxpogaaba7",
			AlbumSlug:  "test-slug",
			AlbumType:  AlbumManual,
			AlbumTitle: "Test Title",
		}

		photo1 := PhotoFixtures.Get("19800101_000002_D640C559")
		photo2 := PhotoFixtures.Get("Photo01")
		photo3 := Photo{}
		photos := Photos{photo1, photo2, photo3}

		added := album.AddPhotos(photos)

		var entries PhotoAlbums

		if err := Db().Where(
			"album_uid = ? AND photo_uid in (?)", "as6sg6bxpogaaba7",
			[]string{
				"ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8",
			},
		).Find(&entries).Error; err != nil {
			t.Fatal(err)
		}

		if len(entries) < 2 {
			t.Fatal("at least one album entry expected")
		}

		var a Album
		if err := Db().Where("album_uid = ?", "as6sg6bxpogaaba7").Find(
			&a,
		).Error; err != nil {
			t.Fatal(err)
		}

		firstUpdatedAt := strings.Split(entries[0].UpdatedAt.String(), ".")[0]
		secondUpdatedAt := strings.Split(entries[1].UpdatedAt.String(), ".")[0]
		albumUpdatedAt := strings.Split(a.UpdatedAt.String(), ".")[0]

		assert.Truef(
			t, firstUpdatedAt <= albumUpdatedAt,
			"Expected the UpdatedAt field of an album to be updated when"+
				" new photos are added",
		)
		assert.Truef(
			t, secondUpdatedAt <= albumUpdatedAt,
			"Expected the UpdatedAt field of an album to be updated when"+
				" new photos are added",
		)
		assert.Equal(t, 2, len(added))
	})
	t.Run("NoUID", func(t *testing.T) {
		album := Album{}

		photo1 := PhotoFixtures.Get("19800101_000002_D640C559")
		photo2 := PhotoFixtures.Get("Photo01")
		photo3 := Photo{}
		photos := Photos{photo1, photo2, photo3}

		added := album.AddPhotos(photos)

		assert.Equal(t, 0, len(added))
	})
}

func TestAlbum_RemovePhotos(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		album := Album{
			ID:         1000000,
			AlbumUID:   "as6sg6bxpogaaba7",
			AlbumSlug:  "test-slug",
			AlbumType:  AlbumManual,
			AlbumTitle: "Test Title",
		}
		removed := album.RemovePhotos([]string{"ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8", "xxx"})

		var entries PhotoAlbums

		if err := Db().Where(
			"album_uid = ? AND photo_uid in (?)", "as6sg6bxpogaaba7",
			[]string{
				"ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8",
			},
		).Find(&entries).Error; err != nil {
			t.Fatal(err)
		}

		if len(entries) < 2 {
			t.Error("at least one album entry expected")
		}

		var a Album
		if err := Db().Where("album_uid = ?", "as6sg6bxpogaaba7").Find(
			&a,
		).Error; err != nil {
			t.Fatal(err)
		}

		first_photo_updatedAt := strings.Split(entries[0].UpdatedAt.String(), ".")[0]
		second_photo_updatedAt := strings.Split(entries[1].UpdatedAt.String(), ".")[0]
		album_updatedAt := strings.Split(a.UpdatedAt.String(), ".")[0]

		assert.Truef(
			t, first_photo_updatedAt <= album_updatedAt,
			"Expected the UpdatedAt field of an album to be updated when"+
				" photos are removed",
		)
		assert.Truef(
			t, second_photo_updatedAt <= album_updatedAt,
			"Expected the UpdatedAt field of an album to be updated when"+
				" photos are removed",
		)

		assert.Equal(t, 2, len(removed))
	})
	t.Run("NoUID", func(t *testing.T) {
		album := Album{}

		added := album.RemovePhotos([]string{"ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"})

		assert.Equal(t, 0, len(added))
	})
}

func TestAlbum_Links(t *testing.T) {
	t.Run("OneResult", func(t *testing.T) {
		album := AlbumFixtures.Get("christmas2030")
		links := album.Links()
		assert.Equal(t, "4jxf3jfn2k", links[0].LinkToken)
	})
}
