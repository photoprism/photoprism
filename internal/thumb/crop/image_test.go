package crop

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThumbFileName(t *testing.T) {
	t.Run("Invalid hash", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}
		_, err := ThumbFileName("xxx", a, s, "path/b")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "invalid file hash")
	})
	t.Run("Path missing", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}
		_, err := ThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", a, s, "")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "path missing")
	})
	t.Run("invalid width", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.000, 0.5)
		s := Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}
		_, err := ThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", a, s, "path/b")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "invalid area width")
	})
	t.Run("invalid crop size", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile50, Tile320, "Lists", 0, 50, DefaultOptions}
		_, err := ThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", a, s, "path/b")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "invalid crop size")
	})
	t.Run("File not found", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}
		_, err := ThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", a, s, "path/b")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "not found")
	})
	t.Run("File exists", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile500, "", "FaceNet", 500, 500, DefaultOptions}
		r, err := ThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", a, s, "./testdata")
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, strings.HasSuffix(r, "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_720x720_fit.jpg"), r)
	})
}

func TestFileWidth(t *testing.T) {
	t.Run("Tile50", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		assert.Equal(t, 49999, FileWidth(a, Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}))
	})
	t.Run("Tile500", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		assert.Equal(t, 499999, FileWidth(a, Size{Tile500, "", "FaceNet", 500, 500, DefaultOptions}))
	})
}

func TestThumbHash(t *testing.T) {
	t.Run("valid filename", func(t *testing.T) {
		assert.Equal(t, "23b05bc917a5aa61382210cedafc162dd3517dc0", thumbHash("23b05bc917a5aa61382210cedafc162dd3517dc0_2048x2048_fit.jpg"))
	})
	t.Run("empty filename", func(t *testing.T) {
		assert.Equal(t, "", thumbHash(""))
	})
}

func TestFindIdealThumbFileName(t *testing.T) {
	t.Run("hash empty", func(t *testing.T) {
		r := findIdealThumbFileName("", 500, "path/b")
		assert.Equal(t, "", r)
	})
	t.Run("path empty", func(t *testing.T) {
		r := findIdealThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", 500, "")
		assert.Equal(t, "", r)
	})
	t.Run("file does not exist", func(t *testing.T) {
		r := findIdealThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", 500, "path/b")
		assert.Equal(t, "", r)
	})
	t.Run("width: 500", func(t *testing.T) {
		r := findIdealThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", 500, "./testdata/b/c/c")
		assert.True(t, strings.HasSuffix(r, "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_720x720_fit.jpg"), r)
	})
	t.Run("width: 720", func(t *testing.T) {
		r := findIdealThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", 720, "./testdata/b/c/c")
		assert.True(t, strings.HasSuffix(r, "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_720x720_fit.jpg"), r)
	})
	t.Run("width: 800", func(t *testing.T) {
		r := findIdealThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", 800, "./testdata/b/c/c")
		assert.True(t, strings.HasSuffix(r, "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_720x720_fit.jpg"), r)
	})
	t.Run("width: 60", func(t *testing.T) {
		r := findIdealThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", 60, "./testdata/b/c/c")
		assert.True(t, strings.HasSuffix(r, "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_720x720_fit.jpg"), r)
	})
}
