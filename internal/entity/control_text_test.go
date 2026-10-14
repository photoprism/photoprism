package entity

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// controlText holds NUL, other C0 controls, ESC, DEL, and every kind of line break.
const controlText = "Ab\x00c\x01d\x1be\x7f\r\nf\rg\nh\ti"

// keptText are values that the control filters must leave unchanged.
var keptText = []string{
	"Sommer am See",
	"صورة من العائلة",
	"שלום עולם",
	"日本語のタイトル",
	"👩\u200D👩\u200D👧 am Strand",
	"می\u200Cخواهم",
}

func TestPhoto_SetTitle_ControlCharacters(t *testing.T) {
	t.Run("Removed", func(t *testing.T) {
		m := Photo{}
		m.SetTitle(controlText, SrcManual)
		assert.Equal(t, "Abcde  f g h i", m.PhotoTitle)
	})
	t.Run("Kept", func(t *testing.T) {
		for _, s := range keptText {
			m := Photo{}
			m.SetTitle(s, SrcManual)
			assert.Equal(t, s, m.PhotoTitle)
		}
	})
}

func TestPhoto_SetCaption_ControlCharacters(t *testing.T) {
	t.Run("Removed", func(t *testing.T) {
		m := Photo{}
		m.SetCaption(controlText, SrcManual)
		assert.Equal(t, "Abcde\nf\ng\nh\ti", m.PhotoCaption)
	})
	t.Run("Kept", func(t *testing.T) {
		for _, s := range append(keptText, "First line\nSecond line\n\n\tIndented") {
			m := Photo{}
			m.SetCaption(s, SrcManual)
			assert.Equal(t, s, m.PhotoCaption)
		}
	})
	t.Run("OnlyControls", func(t *testing.T) {
		m := Photo{PhotoCaption: "Existing", CaptionSrc: SrcMeta}
		m.SetCaption("\x00\x01", SrcManual)
		assert.Equal(t, "Existing", m.PhotoCaption)
	})
}

func TestPhoto_NormalizeValues_ControlCharacters(t *testing.T) {
	t.Run("Removed", func(t *testing.T) {
		m := Photo{PhotoTitle: controlText, PhotoCaption: controlText, PhotoExposure: "1/250\x00",
			CameraSerial: "123\x1b456", TimeZone: "Europe/Berlin"}
		assert.True(t, m.NormalizeValues())
		assert.Equal(t, "Abcde  f g h i", m.PhotoTitle)
		assert.Equal(t, "Abcde\nf\ng\nh\ti", m.PhotoCaption)
		assert.Equal(t, "1/250", m.PhotoExposure)
		assert.Equal(t, "123456", m.CameraSerial)
	})
	t.Run("Description", func(t *testing.T) {
		m := Photo{PhotoDescription: "a\x00b\r\nc", DescriptionSrc: SrcMeta}
		assert.True(t, m.NormalizeValues())
		assert.Equal(t, "ab\nc", m.PhotoCaption)
		assert.Equal(t, "", m.PhotoDescription)
	})
	t.Run("Kept", func(t *testing.T) {
		for _, s := range keptText {
			m := Photo{PhotoTitle: s, PhotoCaption: s + "\nnext", PhotoExposure: "1/250", TimeZone: "Europe/Berlin"}
			assert.False(t, m.NormalizeValues())
			assert.Equal(t, s, m.PhotoTitle)
			assert.Equal(t, s+"\nnext", m.PhotoCaption)
		}
	})
}

func TestSavePhotoForm_ControlCharacters(t *testing.T) {
	m := NewPhoto(false)
	require.NoError(t, m.Create())
	t.Cleanup(func() { _, _ = m.DeletePermanently() })

	photo := FindPhoto(m)
	require.NotNil(t, photo)

	frm, err := form.NewPhoto(*photo)
	require.NoError(t, err)

	frm.PhotoTitle = "Title\x00 with NUL"
	frm.TitleSrc = SrcManual
	frm.PhotoCaption = "Caption\x00\r\nsecond line"
	frm.CaptionSrc = SrcManual
	frm.PhotoExposure = "1/60\x00"
	frm.Details.PhotoID = photo.ID
	frm.Details.Notes = "Notes\x00\nsecond line"
	frm.Details.NotesSrc = SrcManual
	frm.Details.Artist = "Artist\x00\nName"
	frm.Details.ArtistSrc = SrcManual
	frm.Details.Copyright = "(c)\x1b 2026\nAll rights reserved"
	frm.Details.CopyrightSrc = SrcManual

	require.NoError(t, SavePhotoForm(photo, frm))

	found := FindPhoto(*photo)
	require.NotNil(t, found)
	assert.Equal(t, "Title with NUL", found.PhotoTitle)
	assert.Equal(t, "Caption\nsecond line", found.PhotoCaption)
	assert.Equal(t, "1/60", found.PhotoExposure)
	require.NotNil(t, found.Details)
	assert.Equal(t, "Notes\nsecond line", found.Details.Notes)
	assert.Equal(t, "Artist Name", found.Details.Artist)
	assert.Equal(t, "(c) 2026\nAll rights reserved", found.Details.Copyright)
}

func TestClipDetailsText_ControlCharacters(t *testing.T) {
	t.Run("SingleLine", func(t *testing.T) {
		for _, col := range []string{"keywords", "subject", "artist", "software"} {
			assert.Equal(t, "Abcde  f g h i", clipDetailsText(col, controlText), col)
		}
	})
	t.Run("MultiLine", func(t *testing.T) {
		for _, col := range []string{"notes", "copyright", "license"} {
			assert.Equal(t, "Abcde\nf\ng\nh\ti", clipDetailsText(col, controlText), col)
		}
	})
	t.Run("Kept", func(t *testing.T) {
		for _, col := range []string{"keywords", "notes", "subject", "artist", "copyright", "license", "software"} {
			for _, s := range keptText {
				assert.Equal(t, s, clipDetailsText(col, s), col)
			}
		}
	})
	t.Run("Src", func(t *testing.T) {
		assert.Equal(t, SrcMeta, clipDetailsText("notes_src", SrcMeta))
	})
}

func TestDetails_Save_ControlCharacters(t *testing.T) {
	details := Details{
		PhotoID:   900000011,
		Keywords:  "beach,\x00 sun",
		Notes:     "Line one\x00\r\nLine two",
		Subject:   "Sub\x00ject",
		Artist:    "Art\x1bist",
		Copyright: "Copy\x00right\n2026",
		License:   "CC\x7f BY",
		Software:  "Soft\x00ware",
	}
	deleteDetails(t, details.PhotoID)
	t.Cleanup(func() { deleteDetails(t, details.PhotoID) })

	require.NoError(t, details.Save())

	found := findDetails(t, details.PhotoID)
	assert.Equal(t, "beach, sun", found.Keywords)
	assert.Equal(t, "Line one\nLine two", found.Notes)
	assert.Equal(t, "Subject", found.Subject)
	assert.Equal(t, "Artist", found.Artist)
	assert.Equal(t, "Copyright\n2026", found.Copyright)
	assert.Equal(t, "CC BY", found.License)
	assert.Equal(t, "Software", found.Software)
}

func TestClipType_ControlCharacters(t *testing.T) {
	assert.Equal(t, "Display P3", ClipType("Display\x00 P3"))
	assert.Equal(t, "Adobe Photoshop", ClipType("Adobe\x1b Photoshop"))
	assert.Equal(t, "sRGB IEC61966-2.1", ClipType("sRGB IEC61966-2.1"))
}

func TestNewCamera_ControlCharacters(t *testing.T) {
	cam := NewCamera("Canon\x00", "EOS\x1b 5D\x7f")
	assert.Equal(t, "Canon", cam.CameraMake)
	assert.Equal(t, "EOS 5D", cam.CameraModel)
	assert.Equal(t, NewCamera("Canon", "EOS 5D").CameraSlug, cam.CameraSlug)
	assert.Equal(t, &UnknownCamera, NewCamera("\x00", "\x01"))
}

func TestNewLens_ControlCharacters(t *testing.T) {
	lens := NewLens("Canon\x00", "EF 50mm\x1b f/1.8")
	assert.Equal(t, NewLens("Canon", "EF 50mm f/1.8").LensSlug, lens.LensSlug)
	assert.NotContains(t, lens.LensModel, "\x1b")
	assert.Equal(t, &UnknownLens, NewLens("\x00", "\x01"))
}

func TestAlbum_SaveForm_ControlCharacters(t *testing.T) {
	album := NewAlbum("Control Characters Album", AlbumManual)
	require.NoError(t, album.Save())
	t.Cleanup(func() { _ = album.DeletePermanently() })

	frm, err := form.NewAlbum(*album)
	require.NoError(t, err)

	frm.AlbumTitle = "Control\x00 Characters\x1b Album"
	frm.AlbumLocation = "Berlin\x00"
	frm.AlbumCategory = "fam\x01ily"
	frm.AlbumCaption = "Cap\x00tion\r\nnext"
	frm.AlbumDescription = "Desc\x00ription\nnext"
	frm.AlbumNotes = "No\x7ftes\nnext"
	frm.AlbumCountry = "d\x00e"

	require.NoError(t, album.SaveForm(frm))

	found := FindAlbum(Album{AlbumUID: album.AlbumUID})
	require.NotNil(t, found)
	assert.Equal(t, "Control Characters Album", found.AlbumTitle)
	assert.Equal(t, "control-characters-album", found.AlbumSlug)
	assert.Equal(t, "Berlin", found.AlbumLocation)
	assert.Equal(t, "Family", found.AlbumCategory)
	assert.Equal(t, "Caption\nnext", found.AlbumCaption)
	assert.Equal(t, "Description\nnext", found.AlbumDescription)
	assert.Equal(t, "Notes\nnext", found.AlbumNotes)
	assert.Equal(t, "de", found.AlbumCountry)
}

func TestAlbum_SetTitle_ControlCharacters(t *testing.T) {
	t.Run("Removed", func(t *testing.T) {
		album := NewAlbum("Ab\x00c\x1bd", AlbumManual)
		assert.Equal(t, "Abcd", album.AlbumTitle)
	})
	t.Run("Kept", func(t *testing.T) {
		for _, s := range keptText {
			assert.Equal(t, s, NewAlbum(s, AlbumManual).AlbumTitle)
		}
	})
	t.Run("FixtureResolves", func(t *testing.T) {
		fixture := AlbumFixtures.Get("christmas2030")
		album := NewUserAlbum(fixture.AlbumTitle, AlbumManual, fixture.AlbumOrder, fixture.CreatedBy)
		found := album.Find()
		require.NotNil(t, found)
		assert.Equal(t, fixture.AlbumUID, found.AlbumUID)
	})
}

func TestLabel_SaveForm_ControlCharacters(t *testing.T) {
	label := NewLabel("Control Characters Label", 0)
	require.NoError(t, label.Save())
	t.Cleanup(func() { _ = label.Delete() })

	frm, err := form.NewLabel(*label)
	require.NoError(t, err)

	frm.LabelDescription = "Desc\x00ription\r\nnext"
	frm.LabelNotes = "No\x1btes\nnext"
	frm.Thumb = "a1b2\x00c3"
	frm.ThumbSrc = "manual\x00"

	require.NoError(t, label.SaveForm(frm))

	found, err := FindLabel(label.LabelSlug, false)
	require.NoError(t, err)
	assert.Equal(t, "Control Characters Label", found.LabelName)
	assert.Equal(t, "Description\nnext", found.LabelDescription)
	assert.Equal(t, "Notes\nnext", found.LabelNotes)
	assert.Equal(t, "a1b2c3", found.Thumb)
	assert.Equal(t, "manual", found.ThumbSrc)
}

func TestService_SaveForm_ControlCharacters(t *testing.T) {
	frm, err := form.NewService(Service{AccName: "Control", AccURL: "https://dav.example.com/", AccType: "webdav", AccPass: "pass"})
	require.NoError(t, err)

	model, err := AddService(frm)
	require.NoError(t, err)
	t.Cleanup(func() { _ = model.Delete() })

	frm.AccName = "Con\x00trol"
	frm.AccOwner = "Own\x1ber"
	frm.AccUser = "us\x00er"
	frm.AccPass = "pa\x01ss"
	frm.SyncPath = "/sync\x00"
	frm.SharePath = "/share\x7f"

	require.NoError(t, model.SaveForm(frm))

	assert.Equal(t, "Control", model.AccName)
	assert.Equal(t, "Owner", model.AccOwner)
	assert.Equal(t, "user", model.AccUser)
	assert.Equal(t, "pa\x01ss", model.AccPass)
	assert.Equal(t, "/sync", model.SyncPath)
	assert.Equal(t, "/share", model.SharePath)
}

func TestUserDetails_Save_ControlCharacters(t *testing.T) {
	m := NewUserDetails(rnd.GenerateUID(UserUID))
	m.GivenName = "Jo\x00hn"
	m.FamilyName = "عبد\x1bالله"
	m.UserAbout = "About\x00\r\nme"
	m.UserBio = "Bio\x7f\nline"
	m.UserLocation = "Berlin\x00"
	m.SiteURL = "https://example.com/\x00"
	m.NickName = "👩\u200D👩\u200D👧"
	require.NoError(t, m.Save())
	t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

	found := UserDetails{}
	require.NoError(t, UnscopedDb().Where("user_uid = ?", m.UserUID).First(&found).Error)
	assert.Equal(t, "John", found.GivenName)
	assert.Equal(t, "عبدالله", found.FamilyName)
	assert.Equal(t, "About\nme", found.UserAbout)
	assert.Equal(t, "Bio\nline", found.UserBio)
	assert.Equal(t, "Berlin", found.UserLocation)
	assert.Equal(t, "https://example.com/", found.SiteURL)
	assert.Equal(t, "👩\u200D👩\u200D👧", found.NickName)
}

func TestPhoto_normalizeText(t *testing.T) {
	t.Run("Changed", func(t *testing.T) {
		m := Photo{UUID: "a\x00b", TakenSrc: "meta\x00", PhotoType: "image\x01", TypeSrc: "\x00", PhotoTitle: "T\x1bitle",
			TitleSrc: "manual\x00", PhotoCaption: "Line\x00\r\nnext", CaptionSrc: "x\x00", PhotoDescription: "D\x00",
			DescriptionSrc: "x\x00", OriginalName: "orig\x00.jpg", TimeZone: "Europe/Berlin\x00", PlaceID: "zz\x00",
			PlaceSrc: "x\x00", CellID: "zz\x00", PhotoCountry: "d\x00e", PhotoExposure: "1/60\x00",
			CameraSerial: "12\x0034", CameraSrc: "meta\x00", CreatedBy: "u\x00"}
		assert.True(t, m.normalizeText())
		assert.Equal(t, Photo{UUID: "ab", TakenSrc: "meta", PhotoType: "image", TypeSrc: "", PhotoTitle: "Title",
			TitleSrc: "manual", PhotoCaption: "Line\nnext", CaptionSrc: "x", PhotoDescription: "D", DescriptionSrc: "x",
			OriginalName: "orig.jpg", TimeZone: "Europe/Berlin", PlaceID: "zz", PlaceSrc: "x", CellID: "zz",
			PhotoCountry: "de", PhotoExposure: "1/60", CameraSerial: "1234", CameraSrc: "meta", CreatedBy: "u"}, m)
	})
	t.Run("Unchanged", func(t *testing.T) {
		m := Photo{PhotoTitle: "صورة", PhotoCaption: "日本\nnext", OriginalName: "IMG 1.jpg", TimeZone: "Asia/Tokyo",
			PhotoPath: "a\x01b", PhotoName: "c\x01d"}
		assert.False(t, m.normalizeText())
		assert.Equal(t, "a\x01b", m.PhotoPath)
		assert.Equal(t, "c\x01d", m.PhotoName)
	})
}

func TestPhoto_SetCameraSerial_ControlCharacters(t *testing.T) {
	m := Photo{}
	m.SetCameraSerial("AB\n12\t34\x00")
	assert.Equal(t, "AB 12 34", m.CameraSerial)

	m.SetExposure(0, 0, 0, "1\n/60", SrcMeta)
	assert.Equal(t, "1 /60", m.PhotoExposure)

	m.SetDocumentID("xmp.did:\x00abc\n")
	assert.Equal(t, "xmp.did:abc", m.UUID)
}

func TestFile_SetInstanceID_ControlCharacters(t *testing.T) {
	m := File{}
	m.SetInstanceID("xmp.iid:\x00abc\tdef")
	assert.Equal(t, "xmp.iid:abc def", m.InstanceID)
}

func TestAlbum_normalizeText(t *testing.T) {
	m := Album{ParentUID: "a\x00", AlbumSlug: "s\x00", AlbumType: "album\x00", AlbumTitle: "T\x00", AlbumLocation: "L\x00",
		AlbumCategory: "C\x00", AlbumCaption: "c\x00\nn", AlbumDescription: "d\x00\r\nn", AlbumNotes: "n\x7f\nn",
		AlbumFilter: "label:cat\x00", AlbumOrder: "name\x00", AlbumTemplate: "t\x00", AlbumState: "Berlin\x00",
		AlbumCountry: "d\x00e", Thumb: "h\x00", ThumbSrc: "m\x00", CreatedBy: "u\x00", AlbumPath: "a\x01b"}
	m.normalizeText()
	assert.Equal(t, Album{ParentUID: "a", AlbumSlug: "s", AlbumType: "album", AlbumTitle: "T", AlbumLocation: "L",
		AlbumCategory: "C", AlbumCaption: "c\nn", AlbumDescription: "d\nn", AlbumNotes: "n\nn", AlbumFilter: "label:cat",
		AlbumOrder: "name", AlbumTemplate: "t", AlbumState: "Berlin", AlbumCountry: "de", Thumb: "h", ThumbSrc: "m",
		CreatedBy: "u", AlbumPath: "a\x01b"}, m)
}

func TestAlbum_SetLocation_ControlCharacters(t *testing.T) {
	m := Album{}
	m.SetLocation("Ber\x00lin\n", "", "de")
	assert.Equal(t, "Berlin ", m.AlbumLocation)
	m.SetLocation("\x00", "", "de")
	assert.Equal(t, "Berlin ", m.AlbumLocation)
}

func TestUserDetails_normalizeText_AllFields(t *testing.T) {
	m := UserDetails{}
	v := reflect.ValueOf(&m).Elem()

	for i := range v.NumField() {
		if f := v.Field(i); f.Kind() == reflect.String && v.Type().Field(i).Name != "UserUID" {
			f.SetString("x\x00y")
		}
	}

	m.normalizeText()

	for i := range v.NumField() {
		if f := v.Field(i); f.Kind() == reflect.String && v.Type().Field(i).Name != "UserUID" {
			assert.Equal(t, "xy", f.String(), v.Type().Field(i).Name)
		}
	}
}

func TestUserDetails_normalizeText(t *testing.T) {
	m := UserDetails{GivenName: "Jo\x00hn", FamilyName: "عبد\x1bالله", UserAbout: "About\x00\r\nme", UserBio: "Bio\x7f\nline",
		OrgEmail: "a\x00@example.com", NickName: "👩\u200D👩\u200D👧"}
	m.normalizeText()
	assert.Equal(t, UserDetails{GivenName: "John", FamilyName: "عبدالله", UserAbout: "About\nme", UserBio: "Bio\nline",
		OrgEmail: "a@example.com", NickName: "👩\u200D👩\u200D👧"}, m)
}

func TestDetails_SrcControlCharacters(t *testing.T) {
	assert.Equal(t, SrcManual, clipDetailsText("notes_src", SrcManual+"\x00"))
	assert.Equal(t, "", clipDetailsText("notes_src", "manual\x00manualx"))
}
