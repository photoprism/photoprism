package entity

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/form"
)

func TestSavePhotoForm(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		f := form.Photo{
			TakenAt:          time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC),
			TakenSrc:         "manual",
			TimeZone:         "test",
			PhotoTitle:       "Pink beach",
			TitleSrc:         SrcManual,
			PhotoFavorite:    true,
			PhotoPrivate:     true,
			PhotoType:        "image",
			PhotoLat:         7.9999,
			PhotoLng:         8.8888,
			PhotoAltitude:    2,
			PhotoIso:         5,
			PhotoFocalLength: 10,
			PhotoFNumber:     3.3,
			PhotoExposure:    "exposure",
			CameraID:         uint(3),
			CameraSrc:        SrcMeta,
			LensID:           uint(6),
			CellID:           "1234",
			PlaceSrc:         SrcManual,
			PlaceID:          "765",
			PhotoCountry:     "de",
			Details: form.Details{
				PhotoID:   uint(1000008),
				Keywords:  "test cat dog",
				Subject:   "animals",
				Artist:    "Bender",
				Notes:     "notes",
				Copyright: "copy",
				License:   "",
			},
		}

		m := PhotoFixtures.Get("Photo08")

		if err := SavePhotoForm(m, f); err != nil {
			t.Fatal(err)
		}

		Db().First(&m)

		assert.Equal(t, "manual", m.TakenSrc)
		assert.Equal(t, "test", m.TimeZone)
		assert.Equal(t, "Pink beach", m.PhotoTitle)
		assert.Equal(t, "manual", m.TitleSrc)
		assert.Equal(t, true, m.PhotoFavorite)
		assert.Equal(t, true, m.PhotoPrivate)
		assert.Equal(t, "image", m.PhotoType)
		assert.Equal(t, float32(7.9999), m.PhotoLat)
		assert.NotNil(t, m.EditedAt)

		t.Log(m.GetDetails().Keywords)
	})
}

func TestPhoto_SaveLabels(t *testing.T) {
	t.Run("NewPhoto", func(t *testing.T) {
		photo := Photo{
			ID:               11111,
			TakenAt:          time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC),
			TakenAtLocal:     time.Date(2008, 1, 1, 2, 0, 0, 0, time.UTC),
			TakenSrc:         "meta",
			TimeZone:         "UTC",
			PhotoTitle:       "Black beach",
			TitleSrc:         "manual",
			PhotoFavorite:    false,
			PhotoPrivate:     false,
			PhotoType:        "video",
			PhotoLat:         9.9999,
			PhotoLng:         8.8888,
			PhotoAltitude:    2,
			PhotoIso:         5,
			PhotoFocalLength: 10,
			PhotoFNumber:     3.3,
			PhotoExposure:    "exposure",
			CameraID:         uint(3),
			CameraSrc:        "meta",
			LensID:           uint(6),
			CellID:           "1234",
			PlaceSrc:         "geo",
			PlaceID:          "765",
			PhotoCountry:     "de",
			Keywords:         []Keyword{},
			Details: &Details{
				PhotoID:   11111,
				Keywords:  "test cat dog",
				Subject:   "animals",
				Artist:    "Bender",
				Notes:     "notes",
				Copyright: "copy",
				License:   "",
			},
		}

		err := photo.SaveLabels()

		assert.EqualError(t, err, "photo: cannot save to database, id is empty")
	})

	t.Run("ExistingPhoto", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		err := m.SaveLabels()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPhoto_HasUID(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.True(t, m.HasID())
		assert.True(t, m.HasUID())
	})
	t.Run("False", func(t *testing.T) {
		m := Photo{}
		assert.False(t, m.HasID())
		assert.False(t, m.HasUID())
	})
}

func TestPhoto_GetID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.Equal(t, uint(1000001), m.GetID())
	})
}

func TestPhoto_ClassifyLabels(t *testing.T) {
	t.Run("NewPhoto", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo19")
		Db().Set("gorm:auto_preload", true).Model(&m).Related(&m.Labels)
		labels := m.ClassifyLabels()
		assert.Empty(t, labels)
	})
	t.Run("ExistingPhoto", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		Db().Set("gorm:auto_preload", true).Model(&m).Related(&m.Labels)
		labels := m.ClassifyLabels()
		assert.LessOrEqual(t, 2, labels.Len())
	})
	t.Run("EmptyLabel", func(t *testing.T) {
		p := Photo{}
		labels := p.ClassifyLabels()
		assert.Empty(t, labels)
	})
}

func TestPhoto_PreloadFiles(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.Empty(t, m.Files)
		m.PreloadFiles()
		assert.NotEmpty(t, m.Files)
	})
}

func TestPhoto_PreloadKeywords(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.Empty(t, m.Keywords)
		m.PreloadKeywords()
		assert.NotEmpty(t, m.Keywords)
	})
}

func TestPhoto_PreloadAlbums(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.Empty(t, m.Albums)
		m.PreloadAlbums()
		assert.NotEmpty(t, m.Albums)
	})
}

func TestPhoto_PreloadMany(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		assert.Empty(t, m.Albums)
		assert.Empty(t, m.Files)
		assert.Empty(t, m.Keywords)

		m.PreloadMany()

		assert.NotEmpty(t, m.Files)
		assert.NotEmpty(t, m.Albums)
		assert.NotEmpty(t, m.Keywords)
	})
}

func TestPhoto_NoCameraSerial(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo04")
		assert.True(t, m.NoCameraSerial())
	})
	t.Run("False", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo05")
		assert.False(t, m.NoCameraSerial())
	})
}

func TestPhoto_GetDetails(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		result := m.GetDetails()

		if result == nil {
			t.Fatal("result should never be nil")
		}

		if result.PhotoID != 1000000 {
			t.Fatal("PhotoID should not be 1000000")
		}
	})
	t.Run("False", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo12")
		result := m.GetDetails()

		if result == nil {
			t.Fatal("result should never be nil")
		}

		if result.PhotoID != 1000012 {
			t.Fatal("PhotoID should not be 1000012")
		}
	})
	t.Run("NoID", func(t *testing.T) {
		m := Photo{}
		result := m.GetDetails()
		assert.Equal(t, uint(0x0), result.PhotoID)
	})
	t.Run("NewPhotoWithID", func(t *testing.T) {
		m := Photo{ID: 79550, PhotoUID: "prjwufg1z97rcxff"}
		result := m.GetDetails()
		assert.Equal(t, uint(0x136be), result.PhotoID)
	})
}

func TestPhoto_AddLabels(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")
		classifyLabels := classify.Labels{{Name: "cactus", Uncertainty: 30, Source: SrcManual, Priority: 5, Categories: []string{"plant"}}}
		len1 := len(m.Labels)
		m.AddLabels(classifyLabels)
		assert.Greater(t, len(m.Labels), len1)
	})
	t.Run("Update", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		classifyLabels := classify.Labels{{Name: "landscape", Uncertainty: 10, Source: SrcManual, Priority: 5, Categories: []string{"plant"}}}
		assert.Equal(t, 20, m.Labels[0].Uncertainty)
		assert.Equal(t, SrcImage, m.Labels[0].LabelSrc)
		len1 := len(m.Labels)
		m.AddLabels(classifyLabels)
		assert.Equal(t, len(m.Labels), len1)
		assert.Equal(t, 10, m.Labels[0].Uncertainty)
		assert.Equal(t, SrcManual, m.Labels[0].LabelSrc)
	})
}

func TestPhoto_SetDescription(t *testing.T) {
	t.Run("EmptyDescription", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo description non-photographic", m.PhotoDescription)
		m.SetDescription("", SrcManual)
		assert.Equal(t, "photo description non-photographic", m.PhotoDescription)
	})
	t.Run("DescriptionNotFromTheSameSource", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo description non-photographic", m.PhotoDescription)
		m.SetDescription("new photo description", SrcName)
		assert.Equal(t, "photo description non-photographic", m.PhotoDescription)
	})
	t.Run("Ok", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo15")
		assert.Equal(t, "photo description non-photographic", m.PhotoDescription)
		m.SetDescription("new photo description", SrcMeta)
		assert.Equal(t, "new photo description", m.PhotoDescription)
	})
}

func TestPhoto_Delete(t *testing.T) {
	t.Run("NotPermanent", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo16")
		files, err := m.Delete(false)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, files, 1)
	})
	t.Run("Permanent", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo16")
		files, err := m.Delete(true)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, files, 1)
	})
	t.Run("NoID", func(t *testing.T) {
		m := Photo{}
		_, err := m.Delete(true)

		assert.Error(t, err)
	})
}

func TestPhotos_UIDs(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		photo1 := Photo{PhotoUID: "abc123"}
		photo2 := Photo{PhotoUID: "abc456"}
		photos := Photos{photo1, photo2}
		assert.Equal(t, []string{"abc123", "abc456"}, photos.UIDs())
	})
}

func TestPhoto_String(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var m *Photo
		assert.Equal(t, "Photo<nil>", m.String())
		assert.Equal(t, "Photo<nil>", fmt.Sprintf("%s", m))
	})
	t.Run("New", func(t *testing.T) {
		m := &Photo{PhotoUID: "", PhotoName: "", OriginalName: ""}
		assert.Equal(t, "*Photo", m.String())
		assert.Equal(t, "*Photo", fmt.Sprintf("%s", m))
	})
	t.Run("Original", func(t *testing.T) {
		m := Photo{PhotoUID: "", PhotoName: "", OriginalName: "holidayOriginal"}
		assert.Equal(t, "holidayOriginal", m.String())
	})
	t.Run("UID", func(t *testing.T) {
		m := Photo{PhotoUID: "ps6sg6be2lvl0k53", PhotoName: "", OriginalName: ""}
		assert.Equal(t, "uid ps6sg6be2lvl0k53", m.String())
	})
}

func TestPhoto_Create(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		photo := Photo{PhotoUID: "567", PhotoName: "Holiday", OriginalName: "holidayOriginal2"}
		err := photo.Create()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPhoto_Save(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		photo := Photo{PhotoUID: "567", PhotoName: "Holiday", OriginalName: "holidayOriginal2"}
		err := photo.Save()
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		photo := Photo{PhotoUID: "ps6sg6be2lvl0yh0"}
		assert.Error(t, photo.Save())
	})
}

func TestFindPhoto(t *testing.T) {
	t.Run("Save", func(t *testing.T) {
		photo := Photo{PhotoUID: "pt9atdre2lvl0yhx", PhotoName: "Holiday", OriginalName: "holidayOriginal2"}

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, FindPhoto(photo))
	})
	t.Run("Found", func(t *testing.T) {
		photo := Photo{PhotoUID: "ps6sg6be2lvl0yh0"}
		assert.NotNil(t, photo.Find())
		assert.NotNil(t, FindPhoto(photo))
	})
	t.Run("EmptyStruct", func(t *testing.T) {
		photo := Photo{}
		assert.Nil(t, FindPhoto(photo))
		assert.Nil(t, photo.Find())
	})
	t.Run("InvalidID", func(t *testing.T) {
		photo := Photo{ID: 647487}
		assert.Nil(t, FindPhoto(photo))
		assert.Nil(t, photo.Find())
	})
	t.Run("InvalidUID", func(t *testing.T) {
		photo := Photo{PhotoUID: "ps6sg6be2lvl0iuj"}
		assert.Nil(t, FindPhoto(photo))
		assert.Nil(t, photo.Find())
	})
	t.Run("FindByID", func(t *testing.T) {
		photo := Photo{ID: 1000001}
		assert.NotNil(t, FindPhoto(photo))
	})
}

func TestPhoto_RemoveKeyword(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		keyword := Keyword{Keyword: "snake"}
		keyword2 := Keyword{Keyword: "otter"}
		keywords := []Keyword{keyword, keyword2}
		photo := &Photo{Keywords: keywords}
		err := photo.Save()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 2, len(photo.Keywords))
		err = photo.RemoveKeyword("otter")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 2, len(photo.Keywords))
	})
}

func TestPhoto_SyncKeywordLabels(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		labelotter := Label{LabelName: "otter", LabelSlug: "otter"}
		var deletedTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		labelsnake := Label{LabelName: "snake", LabelSlug: "snake", DeletedAt: &deletedTime}

		err := labelsnake.Save()
		if err != nil {
			t.Fatal(err)
		}

		err = labelotter.Save()
		if err != nil {
			t.Fatal(err)
		}

		details := &Details{Keywords: "cow, flower, snake, otter"}
		photo := Photo{ID: 34567, Details: details}

		err = photo.Save()
		if err != nil {
			t.Fatal(err)
		}

		p := FindPhoto(photo)

		assert.Equal(t, 0, len(p.Labels))

		err = p.SyncKeywordLabels()
		if err != nil {
			t.Fatal(err)
		}

		p = FindPhoto(*p)

		assert.Equal(t, 25, len(p.Details.Keywords))
		assert.Equal(t, 3, len(p.Labels))
	})
}

func TestPhoto_LocationLoaded(t *testing.T) {
	t.Run("Photo", func(t *testing.T) {
		photo := Photo{PhotoUID: "56798", PhotoName: "Holiday", OriginalName: "holidayOriginal2"}
		assert.False(t, photo.LocationLoaded())
	})
	t.Run("PhotoWithCell", func(t *testing.T) {
		location := &Cell{Place: nil}
		photo := Photo{PhotoName: "Holiday", Cell: location}
		assert.False(t, photo.LocationLoaded())
	})
}

func TestPhoto_LoadLocation(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo03")
		if err := photo.LoadLocation(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("UnknownLocation", func(t *testing.T) {
		location := &Cell{Place: nil}
		photo := Photo{PhotoName: "Holiday", Cell: location}
		assert.Error(t, photo.LoadLocation())
	})
	t.Run("KnownLocation", func(t *testing.T) {
		location := CellFixtures.Pointer("mexico")
		photo := Photo{PhotoName: "Holiday", Cell: location}
		assert.Error(t, photo.LoadLocation())
	})
}

func TestPhoto_PlaceLoaded(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		photo := Photo{PhotoUID: "56798", PhotoName: "Holiday", OriginalName: "holidayOriginal2"}
		assert.False(t, photo.PlaceLoaded())
	})
}

func TestPhoto_LoadPlace(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo03")
		err := photo.LoadPlace()
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("UnknownLocation", func(t *testing.T) {
		location := &Cell{Place: nil}
		photo := Photo{PhotoName: "Holiday", Cell: location}
		assert.Error(t, photo.LoadPlace())
	})
}

func TestPhoto_HasDescription(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		photo := Photo{PhotoDescription: ""}
		assert.False(t, photo.HasDescription())
	})
	t.Run("True", func(t *testing.T) {
		photo := Photo{PhotoDescription: "bcss"}
		assert.True(t, photo.HasDescription())
	})
}

func TestPhoto_NoDescription(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		photo := Photo{PhotoDescription: ""}
		assert.True(t, photo.NoDescription())
	})
	t.Run("False", func(t *testing.T) {
		photo := Photo{PhotoDescription: "bcss"}
		assert.False(t, photo.NoDescription())
	})
}

func TestPhoto_AllFilesMissing(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		photo := Photo{ID: 6969866}
		assert.True(t, photo.AllFilesMissing())
	})
}

func TestPhoto_Updates(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		photo := Photo{PhotoDescription: "bcss", PhotoName: "InitialName"}

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "InitialName", photo.PhotoName)
		assert.Equal(t, "bcss", photo.PhotoDescription)

		if err := photo.Updates(Photo{PhotoName: "UpdatedName", PhotoDescription: "UpdatedDesc"}); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "UpdatedName", photo.PhotoName)
		assert.Equal(t, "UpdatedDesc", photo.PhotoDescription)

	})
}

func TestPhoto_SetFavorite(t *testing.T) {
	t.Run("SetTrue", func(t *testing.T) {
		photo := Photo{PhotoFavorite: true}

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		if err := photo.SetFavorite(false); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, false, photo.PhotoFavorite)
	})
	t.Run("SetFalse", func(t *testing.T) {
		photo := Photo{PhotoFavorite: false}

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		if err := photo.SetFavorite(true); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, true, photo.PhotoFavorite)
	})
}

func TestPhoto_SetStack(t *testing.T) {
	t.Run("Ignore", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo27")
		assert.Equal(t, IsStackable, m.PhotoStack)
		m.SetStack(IsStackable)
		assert.Equal(t, IsStackable, m.PhotoStack)
	})
	t.Run("Update", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo27")
		assert.Equal(t, IsStackable, m.PhotoStack)
		m.SetStack(IsUnstacked)
		assert.Equal(t, IsUnstacked, m.PhotoStack)
		m.SetStack(IsStackable)
		assert.Equal(t, IsStackable, m.PhotoStack)
	})
}

func TestPhoto_Approve(t *testing.T) {
	t.Run("Quality4", func(t *testing.T) {
		photo := Photo{PhotoQuality: 4}

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		if err := photo.Approve(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 4, photo.PhotoQuality)
	})
	t.Run("Quality1", func(t *testing.T) {
		photo := Photo{PhotoQuality: 1}

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		assert.False(t, photo.Approved())

		if err := photo.Approve(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, photo.PhotoQuality)
		assert.True(t, photo.Approved())
	})
	t.Run("NoID", func(t *testing.T) {
		photo := Photo{PhotoUID: ""}

		assert.False(t, photo.Approved())

		assert.Error(t, photo.Approve())
	})
}

func TestPhoto_Links(t *testing.T) {
	t.Run("OneResult", func(t *testing.T) {
		photo := Photo{PhotoUID: "ps6sg6b1wowuy3c3"}
		links := photo.Links()
		assert.Equal(t, "7jxf3jfn2k", links[0].LinkToken)
	})
}

func TestPhoto_SetPrimary(t *testing.T) {
	t.Run("NoChange", func(t *testing.T) {
		m := PhotoFixtures.Get("19800101_000002_D640C559")

		f1, err := m.PrimaryFile()

		if err != nil {
			t.Fatal(err)
		}

		if err := m.SetPrimary(""); err != nil {
			t.Fatal(err)
		}

		f2, err := m.PrimaryFile()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, f1, f2)
	})
	t.Run("ChangePrimary", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo06")

		f1, err := m.PrimaryFile()

		if err != nil {
			t.Fatal(err)
		}

		assert.NotEqual(t, f1.FileUID, "fs6sg6bqhhinlplo")

		if err := m.SetPrimary("fs6sg6bqhhinlplo"); err != nil {
			t.Fatal(err)
		}

		f2, err := m.PrimaryFile()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, f2.FileUID, "fs6sg6bqhhinlplo")

		if err2 := m.SetPrimary("fs6sg6bqhhinlplp"); err2 != nil {
			t.Fatal(err2)
		}

		f3, err3 := m.PrimaryFile()

		if err3 != nil {
			t.Fatal(err3)
		}

		assert.Equal(t, f3.FileUID, "fs6sg6bqhhinlplp")
	})
	t.Run("PhotoUIDEmpty", func(t *testing.T) {
		m := Photo{}

		err := m.SetPrimary("")
		assert.Error(t, err)
	})
	t.Run("NoPreviewImage", func(t *testing.T) {
		m := Photo{PhotoUID: "1245678"}

		err := m.SetPrimary("")
		assert.Error(t, err)
	})
}

func TestMapKey(t *testing.T) {
	assert.Equal(t, "ogh006/abc236", MapKey(time.Date(2016, 11, 11, 9, 7, 18, 0, time.UTC), "abc236"))
}

func TestNewPhoto(t *testing.T) {
	t.Run("Stackable", func(t *testing.T) {
		m := NewPhoto(true)
		assert.Equal(t, IsStackable, m.PhotoStack)
	})
	t.Run("NotStackable", func(t *testing.T) {
		m := NewPhoto(false)
		assert.Equal(t, IsUnstacked, m.PhotoStack)
	})
}

func TestPhoto_FirstOrCreate(t *testing.T) {
	t.Run("ExistingPhoto", func(t *testing.T) {
		initialUID := "567454"
		photo := Photo{PhotoUID: initialUID, PhotoName: "Light", OriginalName: "lightBlub.jpg"}

		if err := photo.Save(); err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, FindPhoto(photo))
		assert.Nil(t, FindPhoto(Photo{PhotoUID: initialUID}))

		created := photo.FirstOrCreate()

		assert.NotNil(t, created)
		assert.Equal(t, photo.ID, created.ID)
		assert.Equal(t, photo.PhotoUID, created.PhotoUID)
	})
	t.Run("NewPhoto", func(t *testing.T) {
		initialUID := "567459"
		photo := Photo{PhotoUID: initialUID, PhotoName: "Light2", OriginalName: "lightBlub2.jpg"}

		assert.Nil(t, FindPhoto(photo))

		if created := photo.FirstOrCreate(); created == nil {
			t.Fatal("created must not be nil")
		} else {
			assert.Truef(t, created.ID > 0, "%d should be > 0", created.ID)
			assert.Equal(t, photo.PhotoUID, created.PhotoUID)
			assert.Nil(t, FindPhoto(Photo{PhotoUID: initialUID}))
			assert.NotNil(t, FindPhoto(photo))
			assert.NotNil(t, FindPhoto(*created))
		}
	})
}

func TestPhoto_UnknownCamera(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		photo := Photo{}
		assert.True(t, photo.UnknownCamera())
	})
	t.Run("False", func(t *testing.T) {
		photo := Photo{CameraID: 100000}
		assert.False(t, photo.UnknownCamera())
	})
}

func TestPhoto_UnknownLens(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		photo := Photo{}
		assert.True(t, photo.UnknownLens())
	})
	t.Run("False", func(t *testing.T) {
		photo := Photo{LensID: 100000}
		assert.False(t, photo.UnknownLens())
	})
}

func TestPhoto_UpdateDateFields(t *testing.T) {
	t.Run("YearTooSmall", func(t *testing.T) {
		photo := &Photo{TakenAt: time.Date(900, 11, 11, 9, 7, 18, 0, time.UTC)}
		photo.UpdateDateFields()
		assert.Equal(t, time.Date(900, 11, 11, 9, 7, 18, 0, time.UTC), photo.TakenAt)
		assert.Empty(t, photo.TakenAtLocal)
	})
	t.Run("SetToUnknown", func(t *testing.T) {
		photo := &Photo{TakenAt: time.Date(1900, 11, 11, 9, 7, 18, 0, time.UTC), TakenSrc: SrcAuto, CreatedAt: time.Date(1900, 11, 11, 5, 7, 18, 0, time.UTC)}
		photo.UpdateDateFields()
		assert.Equal(t, UnknownYear, photo.PhotoYear)
	})
}

func TestPhoto_SetCamera(t *testing.T) {
	t.Run("CameraNil", func(t *testing.T) {
		photo := &Photo{}
		photo.SetCamera(nil, SrcAuto)
		assert.Empty(t, photo.Camera)
	})
	t.Run("CameraUnknown", func(t *testing.T) {
		photo := &Photo{}
		camera := &Camera{CameraSlug: ""}
		photo.SetCamera(camera, SrcAuto)
		assert.Empty(t, photo.Camera)
	})
	t.Run("DoNotOverwriteManualChanges", func(t *testing.T) {
		cameraOld := &Camera{CameraSlug: "OldCamera", ID: 10000000111}
		photo := &Photo{CameraSrc: SrcManual, Camera: cameraOld, CameraID: 10000000111}
		assert.Equal(t, "OldCamera", photo.Camera.CameraSlug)
		assert.Equal(t, SrcManual, photo.CameraSrc)
		assert.False(t, photo.UnknownCamera())
		camera := &Camera{CameraSlug: "NewCamera"}
		photo.SetCamera(camera, SrcAuto)
		assert.Equal(t, "OldCamera", photo.Camera.CameraSlug)
	})
	t.Run("SetNewCamera", func(t *testing.T) {
		cameraOld := &Camera{CameraSlug: "OldCamera", ID: 10000000111}
		photo := &Photo{CameraSrc: SrcAuto, Camera: cameraOld, CameraID: 10000000111}
		assert.Equal(t, "OldCamera", photo.Camera.CameraSlug)
		camera := &Camera{CameraSlug: "NewCamera"}
		photo.SetCamera(camera, SrcMeta)
		assert.Equal(t, "NewCamera", photo.Camera.CameraSlug)
	})
	t.Run("Scanner", func(t *testing.T) {
		cameraOld := &Camera{CameraSlug: "OldCamera", ID: 10000000111}
		photo := &Photo{CameraSrc: SrcAuto, Camera: cameraOld, CameraID: 10000000111}
		assert.Equal(t, "OldCamera", photo.Camera.CameraSlug)
		assert.False(t, photo.PhotoScan)
		camera := &Camera{CameraSlug: "MSscanner"}
		photo.SetCamera(camera, SrcMeta)
		assert.Equal(t, "MSscanner", photo.Camera.CameraSlug)
		assert.True(t, photo.PhotoScan)
	})
}

func TestPhoto_SetLens(t *testing.T) {
	t.Run("LensNil", func(t *testing.T) {
		photo := &Photo{}
		photo.SetLens(nil, SrcAuto)
		assert.Empty(t, photo.Lens)
	})
	t.Run("LensUnknown", func(t *testing.T) {
		photo := &Photo{}
		lens := &Lens{LensSlug: ""}
		photo.SetLens(lens, SrcAuto)
		assert.Empty(t, photo.Lens)
	})
	t.Run("DoNotOverwriteManualChanges", func(t *testing.T) {
		lensOld := &Lens{LensSlug: "OldLens", ID: 10000000111}
		photo := &Photo{CameraSrc: SrcManual, Lens: lensOld, LensID: 10000000111}
		assert.Equal(t, "OldLens", photo.Lens.LensSlug)
		lens := &Lens{LensSlug: "NewLens"}
		photo.SetLens(lens, SrcAuto)
		assert.Equal(t, "OldLens", photo.Lens.LensSlug)
	})
	t.Run("SetNewLens", func(t *testing.T) {
		lensOld := &Lens{LensSlug: "OldLens", ID: 10000000111}
		photo := &Photo{CameraSrc: SrcAuto, Lens: lensOld, LensID: 10000000111}
		assert.Equal(t, "OldLens", photo.Lens.LensSlug)
		lens := &Lens{LensSlug: "NewLens"}
		photo.SetLens(lens, SrcMeta)
		assert.Equal(t, "NewLens", photo.Lens.LensSlug)
	})
}

func TestPhoto_SetExposure(t *testing.T) {
	t.Run("Priority", func(t *testing.T) {
		photo := &Photo{PhotoFocalLength: 5, PhotoFNumber: 3, PhotoIso: 300, PhotoExposure: "45", CameraSrc: SrcMeta}
		photo.SetExposure(8, 9, 500, "66", SrcManual)
		assert.Equal(t, 8, photo.PhotoFocalLength)
		assert.Equal(t, float32(9), photo.PhotoFNumber)
		assert.Equal(t, 500, photo.PhotoIso)
		assert.Equal(t, "66", photo.PhotoExposure)
	})
	t.Run("NoPriority", func(t *testing.T) {
		photo := &Photo{PhotoFocalLength: 5, PhotoFNumber: 3, PhotoIso: 300, PhotoExposure: "45", CameraSrc: SrcManual}
		photo.SetExposure(8, 9, 500, "66", SrcMeta)
		assert.Equal(t, 5, photo.PhotoFocalLength)
		assert.Equal(t, float32(3), photo.PhotoFNumber)
		assert.Equal(t, 300, photo.PhotoIso)
		assert.Equal(t, "45", photo.PhotoExposure)
	})
	t.Run("ValidRange", func(t *testing.T) {
		photo := &Photo{}
		photo.SetExposure(256000, 256000, 256000, "256000", SrcManual)
		assert.Equal(t, 0, photo.PhotoFocalLength)
		assert.Equal(t, float32(0), photo.PhotoFNumber)
		assert.Equal(t, 0, photo.PhotoIso)
		assert.Equal(t, "256000", photo.PhotoExposure)
		photo.SetExposure(1, 1, 1, "1", SrcManual)
		assert.Equal(t, 1, photo.PhotoFocalLength)
		assert.Equal(t, float32(1), photo.PhotoFNumber)
		assert.Equal(t, 1, photo.PhotoIso)
		assert.Equal(t, "1", photo.PhotoExposure)
	})
}

func TestPhoto_AllFiles(t *testing.T) {
	t.Run("PhotoWithFiles", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo01")
		files := m.AllFiles()
		assert.Equal(t, 2, len(files))
	})
	t.Run("PhotoWithoutFiles", func(t *testing.T) {
		m := &Photo{ID: 100000023456}
		files := m.AllFiles()
		assert.Equal(t, 0, len(files))
	})
}

func TestPhoto_ArchiveRestore(t *testing.T) {
	t.Run("NotYetArchived", func(t *testing.T) {
		m := &Photo{ID: 10000, PhotoUID: "prjwufg1z97rcxff", PhotoTitle: "HappyLilly"}
		assert.Empty(t, m.DeletedAt)
		err := m.Archive()
		if err != nil {
			t.Fatal(err)
		}
		assert.NotEmpty(t, m.DeletedAt)
		err = m.Restore()
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, m.DeletedAt)
	})
	t.Run("AlreadyArchived", func(t *testing.T) {
		var deletedTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		m := &Photo{ID: 10000, PhotoUID: "prjwufg1z97rcxff", PhotoTitle: "HappyLilly", DeletedAt: &deletedTime}
		assert.NotEmpty(t, m.DeletedAt)
		err := m.Archive()
		if err != nil {
			t.Fatal(err)
		}
		assert.NotEmpty(t, m.DeletedAt)
		err = m.Restore()
		if err != nil {
			t.Fatal(err)
		}
		assert.Empty(t, m.DeletedAt)
	})
	t.Run("NoID", func(t *testing.T) {
		m := &Photo{PhotoTitle: "HappyLilly"}
		err := m.Archive()
		assert.Error(t, err)
		err = m.Restore()
		assert.Error(t, err)
	})
}

func TestPhoto_SetCameraSerial(t *testing.T) {
	m := &Photo{}
	assert.Empty(t, m.CameraSerial)
	m.SetCameraSerial("abcCamera")
	assert.Equal(t, "abcCamera", m.CameraSerial)
}

func TestPhoto_MapKey(t *testing.T) {
	m := &Photo{TakenAt: time.Date(2016, 11, 11, 9, 7, 18, 0, time.UTC), CellID: "abc236"}
	assert.Equal(t, "ogh006/abc236", m.MapKey())
}

func TestPhoto_FaceCount(t *testing.T) {
	t.Run("Photo04", func(t *testing.T) {
		m := PhotoFixtures.Get("Photo04")
		assert.Equal(t, 3, m.FaceCount())
	})
}
