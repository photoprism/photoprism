package customize

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultSettings(t *testing.T) {
	s := NewDefaultSettings()

	assert.IsType(t, new(Settings), s)
	assert.Equal(t, DefaultTheme, s.UI.Theme)
	assert.Equal(t, DefaultLocale, s.UI.Language)
}

func TestNewSettings(t *testing.T) {
	s := NewSettings("test", "fr")

	assert.IsType(t, new(Settings), s)
	assert.Equal(t, "test", s.UI.Theme)
	assert.Equal(t, "fr", s.UI.Language)
	assert.Equal(t, "fr", s.UI.Language)
	assert.Equal(t, true, s.Search.ListView)
}

func TestSettings_Load(t *testing.T) {
	t.Run("existing filename", func(t *testing.T) {
		s := NewDefaultSettings()

		if err := s.Load("testdata/settings.yml"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "onyx", s.UI.Theme)
		assert.Equal(t, "de", s.UI.Language)
	})
	t.Run("not existing filename", func(t *testing.T) {
		s := NewDefaultSettings()

		err := s.Load("testdata/settings_123.yml")

		assert.Error(t, err)

		assert.Equal(t, "default", s.UI.Theme)
		assert.Equal(t, "en", s.UI.Language)
	})
}
func TestSettings_Save(t *testing.T) {
	t.Run("existing filename", func(t *testing.T) {
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
	t.Run("not existing filename", func(t *testing.T) {
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
