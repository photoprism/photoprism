package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/stretchr/testify/assert"
)

func TestCreateUserSettings(t *testing.T) {
	t.Run("Empty UID", func(t *testing.T) {
		m := &User{UserUID: ""}
		assert.Error(t, CreateUserSettings(m))
		assert.Nil(t, m.UserSettings)
	})
	t.Run("Success", func(t *testing.T) {
		m := &User{UserUID: "1234"}
		err := CreateUserSettings(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, m.UserSettings)
	})
}

func TestUserSettings_HasID(t *testing.T) {
	u := FindUserByName("alice")
	assert.True(t, u.UserSettings.HasID())
}

func TestUserSettings_Updates(t *testing.T) {
	m := &User{
		UserUID: "1234",
		UserSettings: &UserSettings{
			UITheme:    "carbon",
			UILanguage: "de",
		}}

	m.UserSettings.Updates(UserSettings{UITheme: "vanta", UILanguage: "en"})
	assert.Equal(t, "vanta", m.UserSettings.UITheme)
	assert.Equal(t, "en", m.UserSettings.UILanguage)
}

func TestUserSettings_Apply(t *testing.T) {
	m := &UserSettings{
		UITheme:    "carbon",
		UILanguage: "de",
	}

	s := &customize.Settings{
		UI: customize.UISettings{
			Theme:    "onyx",
			Language: "nl",
			TimeZone: "Europe/Berlin",
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
		},
	}
	r := m.Apply(s)

	assert.Equal(t, "nl", r.UILanguage)
	assert.Equal(t, "onyx", r.UITheme)
	assert.Equal(t, "Europe/Berlin", r.UITimeZone)
	assert.Equal(t, -1, r.DownloadMediaRaw)
	assert.Equal(t, 1, r.DownloadOriginals)
	assert.Equal(t, 1, r.DownloadMediaSidecar)
	assert.Equal(t, "outdoor", r.MapsStyle)
	assert.Equal(t, 1, r.MapsAnimate)
	assert.Equal(t, 1, r.IndexRescan)
	assert.Equal(t, "index-path", r.IndexPath)
	assert.Equal(t, -1, r.ImportMove)
	assert.Equal(t, "imports/2023", r.ImportPath)

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
}

func TestUserSettings_ApplyTo(t *testing.T) {
	m := &UserSettings{
		UITheme:              "lavender",
		UILanguage:           "ch",
		UITimeZone:           "Europe",
		MapsStyle:            "satellite",
		MapsAnimate:          1,
		IndexPath:            "flowers",
		IndexRescan:          -1,
		ImportPath:           "import",
		ImportMove:           1,
		DownloadOriginals:    -1,
		DownloadMediaRaw:     1,
		DownloadMediaSidecar: -1,
	}

	s := &customize.Settings{
		UI: customize.UISettings{
			Theme:    "onyx",
			Language: "nl",
			TimeZone: "Europe/Berlin",
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
	assert.Equal(t, true, r.Download.MediaRaw)
	assert.Equal(t, false, r.Download.Originals)
	assert.Equal(t, false, r.Download.MediaSidecar)
	assert.Equal(t, "satellite", r.Maps.Style)
	assert.Equal(t, 1, r.Maps.Animate)
	assert.Equal(t, false, r.Index.Rescan)
	assert.Equal(t, "flowers", r.Index.Path)
	assert.Equal(t, true, r.Import.Move)
	assert.Equal(t, "import", r.Import.Path)

	m2 := &UserSettings{
		UITheme:              "lavender",
		UILanguage:           "ch",
		UITimeZone:           "Europe",
		MapsStyle:            "satellite",
		MapsAnimate:          -1,
		IndexPath:            "flowers",
		IndexRescan:          1,
		ImportPath:           "import",
		ImportMove:           -1,
		DownloadOriginals:    1,
		DownloadMediaRaw:     -1,
		DownloadMediaSidecar: 1,
	}

	r2 := m2.ApplyTo(s)

	assert.IsType(t, &customize.Settings{}, r2)
	assert.Equal(t, "ch", s.UI.Language)
	assert.Equal(t, "lavender", s.UI.Theme)
	assert.Equal(t, "Europe", s.UI.TimeZone)
	assert.Equal(t, false, s.Download.MediaRaw)
	assert.Equal(t, true, s.Download.Originals)
	assert.Equal(t, true, s.Download.MediaSidecar)
	assert.Equal(t, "satellite", s.Maps.Style)
	assert.Equal(t, 0, s.Maps.Animate)
	assert.Equal(t, true, s.Index.Rescan)
	assert.Equal(t, "flowers", s.Index.Path)
	assert.Equal(t, false, s.Import.Move)
	assert.Equal(t, "import", s.Import.Path)
}
