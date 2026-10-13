package entity

import (
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestCreateUserSettings(t *testing.T) {
	t.Run("EmptyUid", func(t *testing.T) {
		m := &User{UserUID: ""}
		assert.Error(t, CreateUserSettings(m))
		assert.Nil(t, m.UserSettings)
	})
	t.Run("InvalidUid", func(t *testing.T) {
		m := &User{UserUID: "pqxetse3cy5eo9z2"}
		assert.ErrorIs(t, CreateUserSettings(m), ErrInvalidUserUID)
		assert.Equal(t, int64(0), countUserSettings(t, m.UserUID))
	})
	t.Run("Success", func(t *testing.T) {
		m := &User{UserUID: "usg73p55zwgr1gbq"}
		err := CreateUserSettings(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, m.UserSettings)
	})
	t.Run("SuccessNewRecord", func(t *testing.T) {
		userUID := rnd.GenerateUID('u')
		m := &User{UserUID: userUID}
		require.NoError(t, Db().Create(m).Error) // Have to create a user BEFORE adding settings to it.
		require.NoError(t, CreateUserSettings(m))
		assert.NotNil(t, m.UserSettings)
	})
}

func TestUserSettings_HasID(t *testing.T) {
	u := FindUserByName("alice")
	assert.True(t, u.UserSettings.HasID())
}

// invalidUserUIDs lists values that settings and details must not be written for.
var invalidUserUIDs = []string{"", "\x01", "1234", "usg73p55zwgr1gb", "usg73p55zwgr1gb-", "pqxetse3cy5eo9z2"}

// countUserSettings returns the number of settings rows with the given user UID.
func countUserSettings(t *testing.T, uid string) (n int64) {
	t.Helper()
	require.NoError(t, UnscopedDb().Model(&UserSettings{}).Where("user_uid = ?", uid).Count(&n).Error)
	return n
}

func TestUserSettings_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewUserSettings(rnd.GenerateUID(UserUID))
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })
		assert.Equal(t, int64(1), countUserSettings(t, m.UserUID))
	})
	t.Run("InvalidUid", func(t *testing.T) {
		for _, uid := range invalidUserUIDs {
			before := countUserSettings(t, uid)
			assert.ErrorIs(t, NewUserSettings(uid).Create(), ErrInvalidUserUID, "uid %q", uid)
			assert.Equal(t, before, countUserSettings(t, uid), "uid %q", uid)
		}
	})
}

func TestUserSettings_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewUserSettings(rnd.GenerateUID(UserUID))
		m.UILanguage = "de"
		require.NoError(t, m.Save())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		m.UILanguage = "nl"
		require.NoError(t, m.Save())
		assert.Equal(t, int64(1), countUserSettings(t, m.UserUID))

		found := UserSettings{}
		require.NoError(t, UnscopedDb().Where("user_uid = ?", m.UserUID).First(&found).Error)
		assert.Equal(t, "nl", found.UILanguage)
	})
	t.Run("InvalidUid", func(t *testing.T) {
		for _, uid := range invalidUserUIDs {
			before := countUserSettings(t, uid)
			assert.ErrorIs(t, NewUserSettings(uid).Save(), ErrInvalidUserUID, "uid %q", uid)
			assert.Equal(t, before, countUserSettings(t, uid), "uid %q", uid)
		}
	})
}

func TestUserSettings_Updates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewUserSettings(rnd.GenerateUID(UserUID))
		m.UITheme, m.UILanguage = "carbon", "de"
		require.NoError(t, m.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

		assert.NoError(t, m.Updates(UserSettings{UITheme: "vanta", UILanguage: "en"}))
		assert.Equal(t, "vanta", m.UITheme)
		assert.Equal(t, "en", m.UILanguage)

		found := UserSettings{}
		require.NoError(t, UnscopedDb().Where("user_uid = ?", m.UserUID).First(&found).Error)
		assert.Equal(t, "vanta", found.UITheme)
		assert.Equal(t, "en", found.UILanguage)
	})
	t.Run("NoKey", func(t *testing.T) {
		m := &UserSettings{UITheme: "carbon"}

		assert.ErrorIs(t, m.Updates(UserSettings{UITheme: "vanta"}), gorm.ErrMissingWhereClause)
	})
}

func TestUserSettings_Apply(t *testing.T) {
	m := &UserSettings{
		UITheme:    "carbon",
		UILanguage: "de",
	}

	s := &customize.Settings{
		UI: customize.UISettings{
			Theme:     "onyx",
			Language:  "nl",
			TimeZone:  "Europe/Berlin",
			StartPage: "Places",
		},
		Download: customize.DownloadSettings{
			Name:         "file",
			Disabled:     false,
			Originals:    true,
			MediaRaw:     false,
			MediaSidecar: true,
		},
		Maps: customize.MapsSettings{
			Animate: 1,
			Style:   "outdoor",
		},
		Index: customize.IndexSettings{
			Path:         "index-path",
			Convert:      true,
			Rescan:       true,
			SkipArchived: false,
		},
		Import: customize.ImportSettings{
			Path: "imports/2023",
			Move: false,
			Dest: customize.DefaultImportDest,
		},
		Search: customize.SearchSettings{
			BatchSize:    -1,
			ListView:     true,
			ShowTitles:   false,
			ShowCaptions: true,
		},
	}
	r := m.Apply(s)

	assert.Equal(t, "nl", r.UILanguage)
	assert.Equal(t, "onyx", r.UITheme)
	assert.Equal(t, "Europe/Berlin", r.UITimeZone)
	assert.Equal(t, "Places", r.UIStartPage)
	assert.Equal(t, -1, r.DownloadMediaRaw)
	assert.Equal(t, 1, r.DownloadOriginals)
	assert.Equal(t, 1, r.DownloadMediaSidecar)
	assert.Equal(t, "outdoor", r.MapsStyle)
	assert.Equal(t, 1, r.MapsAnimate)
	assert.Equal(t, 1, r.IndexRescan)
	assert.Equal(t, "index-path", r.IndexPath)
	assert.Equal(t, -1, r.ImportMove)
	assert.Equal(t, "imports/2023", r.ImportPath)
	assert.Equal(t, 1, r.SearchListView)
	assert.Equal(t, -1, r.SearchShowTitles)
	assert.Equal(t, 1, r.SearchShowCaptions)

	s2 := &customize.Settings{
		Download: customize.DownloadSettings{
			Name:         "file",
			Disabled:     false,
			Originals:    false,
			MediaRaw:     true,
			MediaSidecar: false,
		},
		Maps: customize.MapsSettings{
			Animate: 0,
			Style:   "outdoor",
		},
		Index: customize.IndexSettings{
			Path:         "index-path",
			Convert:      true,
			Rescan:       false,
			SkipArchived: false,
		},
		Import: customize.ImportSettings{
			Path: "imports/2023",
			Move: true,
		},
		Search: customize.SearchSettings{
			BatchSize:    -1,
			ListView:     false,
			ShowTitles:   true,
			ShowCaptions: false,
		},
	}
	r2 := m.Apply(s2)

	assert.Equal(t, "nl", r2.UILanguage)
	assert.Equal(t, "onyx", r2.UITheme)
	assert.Equal(t, "Europe/Berlin", r2.UITimeZone)
	assert.Equal(t, 1, r2.DownloadMediaRaw)
	assert.Equal(t, -1, r2.DownloadOriginals)
	assert.Equal(t, -1, r2.DownloadMediaSidecar)
	assert.Equal(t, "outdoor", r2.MapsStyle)
	assert.Equal(t, -1, r2.MapsAnimate)
	assert.Equal(t, -1, r2.IndexRescan)
	assert.Equal(t, "index-path", r2.IndexPath)
	assert.Equal(t, 1, r2.ImportMove)
	assert.Equal(t, "imports/2023", r2.ImportPath)
	assert.Equal(t, -1, r.SearchListView)
	assert.Equal(t, 1, r.SearchShowTitles)
	assert.Equal(t, -1, r.SearchShowCaptions)
}

func TestUserSettings_ApplyBounds(t *testing.T) {
	long := strings.Repeat("a", 200)
	path := "/" + strings.Repeat("ä", 1000)

	m := NewUserSettings(rnd.GenerateUID(UserUID)).Apply(&customize.Settings{
		UI:     customize.UISettings{Theme: long, Language: long, TimeZone: long, StartPage: long},
		Maps:   customize.MapsSettings{Style: long, Animate: math.MaxInt},
		Index:  customize.IndexSettings{Path: path},
		Import: customize.ImportSettings{Path: path},
	})

	assert.Len(t, m.UITheme, userSettingsTypeBytes)
	assert.Len(t, m.UILanguage, userSettingsTypeBytes)
	assert.Len(t, m.MapsStyle, userSettingsTypeBytes)
	assert.Equal(t, "", m.UITimeZone)
	assert.LessOrEqual(t, len(m.UIStartPage), 64)
	assert.Equal(t, math.MaxInt32, m.MapsAnimate)
	assert.LessOrEqual(t, len(m.IndexPath), PathBytes)
	assert.LessOrEqual(t, len(m.ImportPath), PathBytes)
	assert.True(t, utf8.ValidString(m.IndexPath))
	assert.True(t, utf8.ValidString(m.ImportPath))

	require.NoError(t, m.Save())
	t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })

	found := UserSettings{}
	require.NoError(t, UnscopedDb().Where("user_uid = ?", m.UserUID).First(&found).Error)
	assert.Equal(t, m.UITheme, found.UITheme)
	assert.Equal(t, m.MapsAnimate, found.MapsAnimate)
	assert.Equal(t, m.IndexPath, found.IndexPath)

	t.Run("ControlChars", func(t *testing.T) {
		nul := "a\x00b\x1f"
		c := NewUserSettings(rnd.GenerateUID(UserUID)).Apply(&customize.Settings{
			UI:     customize.UISettings{Theme: nul, Language: nul, TimeZone: nul, StartPage: nul},
			Maps:   customize.MapsSettings{Style: nul},
			Index:  customize.IndexSettings{Path: "/" + nul},
			Import: customize.ImportSettings{Path: "/" + nul},
		})

		for _, v := range []string{c.UITheme, c.UILanguage, c.UIStartPage, c.MapsStyle} {
			assert.Equal(t, "ab", v)
		}

		// A time zone with control characters is not a valid name, so none is stored.
		assert.Equal(t, "", c.UITimeZone)

		assert.Equal(t, "/ab", c.IndexPath)
		assert.Equal(t, "/ab", c.ImportPath)

		require.NoError(t, c.Save())
		t.Cleanup(func() { _ = UnscopedDb().Delete(c).Error })
	})
}

func TestUserSettings_ApplyTo(t *testing.T) {
	m := &UserSettings{
		UITheme:              "lavender",
		UILanguage:           "ch",
		UITimeZone:           "Europe",
		UIStartPage:          "",
		MapsStyle:            "satellite",
		MapsAnimate:          1,
		IndexPath:            "flowers",
		IndexRescan:          -1,
		ImportPath:           "import",
		ImportMove:           1,
		DownloadOriginals:    -1,
		DownloadMediaRaw:     1,
		DownloadMediaSidecar: -1,
		SearchListView:       -1,
		SearchShowTitles:     1,
		SearchShowCaptions:   -1,
	}

	s := &customize.Settings{
		UI: customize.UISettings{
			Theme:     "onyx",
			Language:  "nl",
			TimeZone:  "Europe/Amsterdam",
			StartPage: "default",
		},
		Download: customize.DownloadSettings{
			Name:         "file",
			Disabled:     false,
			Originals:    true,
			MediaRaw:     false,
			MediaSidecar: true,
		},
		Maps: customize.MapsSettings{
			Animate: 0,
			Style:   "outdoor",
		},
		Index: customize.IndexSettings{
			Path:         "index-path",
			Convert:      true,
			Rescan:       true,
			SkipArchived: false,
		},
		Import: customize.ImportSettings{
			Path: "imports/2023",
			Move: false,
		},
	}
	r := m.ApplyTo(s)

	assert.IsType(t, &customize.Settings{}, r)
	assert.Equal(t, "ch", r.UI.Language)
	assert.Equal(t, "lavender", r.UI.Theme)
	assert.Equal(t, "Europe", r.UI.TimeZone)
	assert.Equal(t, "default", r.UI.StartPage)
	assert.Equal(t, true, r.Download.MediaRaw)
	assert.Equal(t, false, r.Download.Originals)
	assert.Equal(t, false, r.Download.MediaSidecar)
	assert.Equal(t, "satellite", r.Maps.Style)
	assert.Equal(t, 1, r.Maps.Animate)
	assert.Equal(t, false, r.Index.Rescan)
	assert.Equal(t, "flowers", r.Index.Path)
	assert.Equal(t, true, r.Import.Move)
	assert.Equal(t, "import", r.Import.Path)
	assert.Equal(t, false, r.Search.ListView)
	assert.Equal(t, true, r.Search.ShowTitles)
	assert.Equal(t, false, r.Search.ShowCaptions)

	m2 := &UserSettings{
		UITheme:              "lavender",
		UILanguage:           "ch",
		UITimeZone:           "Europe",
		UIStartPage:          "Places",
		MapsStyle:            "satellite",
		MapsAnimate:          -1,
		IndexPath:            "flowers",
		IndexRescan:          1,
		ImportPath:           "import",
		ImportMove:           -1,
		DownloadOriginals:    1,
		DownloadMediaRaw:     -1,
		DownloadMediaSidecar: 1,
		SearchListView:       1,
		SearchShowTitles:     -1,
		SearchShowCaptions:   1,
	}

	r2 := m2.ApplyTo(s)

	assert.IsType(t, &customize.Settings{}, r2)
	assert.Equal(t, "ch", s.UI.Language)
	assert.Equal(t, "lavender", s.UI.Theme)
	assert.Equal(t, "Europe", s.UI.TimeZone)
	assert.Equal(t, "Places", s.UI.StartPage)
	assert.Equal(t, false, s.Download.MediaRaw)
	assert.Equal(t, true, s.Download.Originals)
	assert.Equal(t, true, s.Download.MediaSidecar)
	assert.Equal(t, "satellite", s.Maps.Style)
	assert.Equal(t, 0, s.Maps.Animate)
	assert.Equal(t, true, s.Index.Rescan)
	assert.Equal(t, "flowers", s.Index.Path)
	assert.Equal(t, false, s.Import.Move)
	assert.Equal(t, "import", s.Import.Path)
	assert.Equal(t, true, r.Search.ListView)
	assert.Equal(t, false, r.Search.ShowTitles)
	assert.Equal(t, true, r.Search.ShowCaptions)
}

func TestUserSettings_ApplyTimeZone(t *testing.T) {
	apply := func(stored, zone string) string {
		m := &UserSettings{UITimeZone: stored}
		m.Apply(&customize.Settings{UI: customize.UISettings{TimeZone: zone}})
		return m.UITimeZone
	}

	t.Run("Stored", func(t *testing.T) {
		for _, zone := range []string{"Local", "UTC", "GMT", "UTC+1", "UTC-12", "Europe/Berlin", "America/Port-au-Prince"} {
			assert.Equal(t, zone, apply("", zone), zone)
		}
		assert.Equal(t, "UTC+2", apply("", "+02:00"))
	})
	t.Run("Dropped", func(t *testing.T) {
		for _, zone := range []string{"Europe/Berlin\x00", "Europe/\x1bBerlin", "UTC+99", "Europe//Berlin", strings.Repeat("A", 65)} {
			assert.Equal(t, "Asia/Tokyo", apply("Asia/Tokyo", zone), "%q", zone)
		}
	})
}
