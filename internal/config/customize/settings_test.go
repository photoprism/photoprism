package customize

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSettings(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		s := NewDefaultSettings()
		assert.IsType(t, new(Settings), s)
		assert.Equal(t, DefaultTheme, s.UI.Theme)
		assert.Equal(t, DefaultLanguage, s.UI.Language)
		assert.Equal(t, DefaultTimeZone, s.UI.TimeZone)
		assert.Equal(t, DefaultStartPage, s.UI.StartPage)
		assert.Equal(t, DefaultMapsStyle, s.Maps.Style)
	})
	t.Run("Custom", func(t *testing.T) {
		s := NewSettings("test", "fr", "Europe/Paris")
		assert.IsType(t, new(Settings), s)
		assert.Equal(t, "test", s.UI.Theme)
		assert.Equal(t, "fr", s.UI.Language)
		assert.Equal(t, "Europe/Paris", s.UI.TimeZone)
		assert.Equal(t, true, s.Search.ListView)
		assert.Equal(t, true, s.Search.ShowTitles)
		assert.Equal(t, true, s.Search.ShowCaptions)
		assert.Equal(t, DefaultStartPage, s.UI.StartPage)
		assert.Equal(t, DefaultMapsStyle, s.Maps.Style)
		s.UI.Language = ""
		s.UI.TimeZone = ""
		s.UI.StartPage = ""
		s.Maps.Style = ""
		s.Propagate()
		assert.Equal(t, DefaultLanguage, s.UI.Language)
		assert.Equal(t, DefaultTimeZone, s.UI.TimeZone)
		assert.Equal(t, DefaultStartPage, s.UI.StartPage)
		assert.Equal(t, DefaultMapsStyle, s.Maps.Style)
	})
}

func TestSettings_Load(t *testing.T) {
	t.Run("ExistingFilename", func(t *testing.T) {
		s := NewDefaultSettings()

		if err := s.Load("testdata/settings.yml"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "onyx", s.UI.Theme)
		assert.Equal(t, "de", s.UI.Language)
	})
	t.Run("NotExistingFilename", func(t *testing.T) {
		s := NewDefaultSettings()

		err := s.Load("testdata/settings_123.yml")

		assert.Error(t, err)

		assert.Equal(t, "default", s.UI.Theme)
		assert.Equal(t, "en", s.UI.Language)
	})
}
func TestSettings_Save(t *testing.T) {
	t.Run("ExistingFilename", func(t *testing.T) {
		s := NewDefaultSettings()

		assert.Equal(t, "default", s.UI.Theme)
		assert.Equal(t, "en", s.UI.Language)

		s.UI.Theme = "onyx"
		s.UI.Language = "de"

		assert.Equal(t, "onyx", s.UI.Theme)
		assert.Equal(t, "de", s.UI.Language)

		if err := s.Save("testdata/settings.yml"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("NotExistingFilename", func(t *testing.T) {
		s := NewDefaultSettings()
		s.UI.Theme = "onyx"
		s.UI.Language = "de"

		assert.Equal(t, "onyx", s.UI.Theme)
		assert.Equal(t, "de", s.UI.Language)

		if err := s.Save("testdata/settings_tmp.yml"); err != nil {
			t.Fatal(err)
		}

		if err := os.Remove("testdata/settings_tmp.yml"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSettings_Stacks(t *testing.T) {
	s := NewDefaultSettings()

	assert.False(t, s.StackSequences())
	assert.True(t, s.StackUUID())
	assert.True(t, s.StackMeta())
}
