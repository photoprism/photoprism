package crop

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestThumbFileName(t *testing.T) {
	t.Run("InvalidHash", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}
		_, err := ThumbFileName("xxx", a, s, "path/b")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "invalid file hash")
	})
	t.Run("PathMissing", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}
		_, err := ThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", a, s, "")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "path missing")
	})
	t.Run("InvalidWidth", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.000, 0.5)
		s := Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}
		_, err := ThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", a, s, "path/b")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "invalid area width")
	})
	t.Run("InvalidCropSize", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile50, Tile320, "Lists", 0, 50, DefaultOptions}
		_, err := ThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", a, s, "path/b")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "invalid crop size")
	})
	t.Run("FileNotFound", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile50, Tile320, "Lists", 50, 50, DefaultOptions}
		_, err := ThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", a, s, "path/b")
		if err == nil {
			t.Fatal(err)
		}
		assert.Contains(t, err.Error(), "not found")
	})
	t.Run("FileExists", func(t *testing.T) {
		a := NewArea("face", 1.000, 0.33333, 0.001, 0.5)
		s := Size{Tile500, "", "FaceNet", 500, 500, DefaultOptions}
		r, err := ThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", a, s, "./testdata")
		if err != nil {
			t.Fatal(err)
		}
		// A sliver of a crop asks for a source no rendition has, so the widest one is used.
		assert.True(t, strings.HasSuffix(r, "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_1280x1024_fit.jpg"), r)
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
	t.Run("ValidFilename", func(t *testing.T) {
		assert.Equal(t, "23b05bc917a5aa61382210cedafc162dd3517dc0", thumbHash("23b05bc917a5aa61382210cedafc162dd3517dc0_2048x2048_fit.jpg"))
	})
	t.Run("EmptyFilename", func(t *testing.T) {
		assert.Equal(t, "", thumbHash(""))
	})
}

func TestFindIdealThumbFileName(t *testing.T) {
	t.Run("HashEmpty", func(t *testing.T) {
		r := findIdealThumbFileName("", 500, "path/b")
		assert.Equal(t, "", r)
	})
	t.Run("PathEmpty", func(t *testing.T) {
		r := findIdealThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", 500, "")
		assert.Equal(t, "", r)
	})
	t.Run("FileDoesNotExist", func(t *testing.T) {
		r := findIdealThumbFileName("2105662d3f8d6e68d9e94280449fbf26ed89xxxx", 500, "path/b")
		assert.Equal(t, "", r)
	})
	// The renditions belong to a portrait picture, so neither reaches the width its name states:
	// the one called 720x720 is 479 px wide and the one called 1280x1024 is 681 px wide.
	const fit720 = "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_720x720_fit.jpg"
	const fit1280 = "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_1280x1024_fit.jpg"
	t.Run("SmallestThatCovers", func(t *testing.T) {
		r := findIdealThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", 60, "./testdata/b/c/c")
		assert.True(t, strings.HasSuffix(r, fit720), r)
	})
	t.Run("SkipsARenditionNarrowerThanItsName", func(t *testing.T) {
		r := findIdealThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", 500, "./testdata/b/c/c")
		assert.True(t, strings.HasSuffix(r, fit1280), r)
	})
	t.Run("WidestAvailableWhenNoneCovers", func(t *testing.T) {
		r := findIdealThumbFileName("bccfeaa526a36e19b555fd4ca5e8f767d5604289", 4000, "./testdata/b/c/c")
		assert.True(t, strings.HasSuffix(r, fit1280), r)
	})
}

func TestImageFromThumb(t *testing.T) {
	t.Run("Layered16BitTiffThumbnail", func(t *testing.T) {
		prevLibrary := thumb.Library
		thumb.Library = thumb.LibVips
		t.Cleanup(func() {
			thumb.Library = prevLibrary
		})

		cachePath := t.TempDir()
		src := fs.Abs("../../../assets/samples/layered-16bit-small.tif")
		fit720 := thumb.Sizes[thumb.Fit720]

		thumbName, err := thumb.FromFile(src, "bccfeaa526a36e19b555fd4ca5e8f767d5604289", cachePath, fit720.Width, fit720.Height, thumb.OrientationNormal, fit720.Options...)
		if err != nil {
			t.Fatal(err)
		}

		img, cropName, srcWidth, err := ImageFromThumb(thumbName, NewArea("crop", 0, 0, 1, 1), Sizes[Tile50], false)
		if err != nil {
			t.Fatal(err)
		}

		assert.NotEmpty(t, cropName)
		assert.Equal(t, filepath.Join(filepath.Dir(thumbName), "bccfeaa526a36e19b555fd4ca5e8f767d5604289_50x50_crop_0000003e83e8.jpg"), cropName)
		assert.Equal(t, 50, img.Bounds().Dx())
		assert.Equal(t, 50, img.Bounds().Dy())
		// The source the crop was taken from, not the crop itself: recording the latter would
		// store the requested size back, which is a constant and says nothing about quality.
		assert.Positive(t, srcWidth)
		assert.NotEqual(t, 50, srcWidth)
	})
}

func TestImageFromIdealThumb(t *testing.T) {
	const hash = "bccfeaa526a36e19b555fd4ca5e8f767d5604289"

	cachePath := t.TempDir()
	src := fs.Abs("../../../assets/samples/6720px_white.jpg")

	renditionWidth := func(t *testing.T, name thumb.Name) int {
		t.Helper()
		size := thumb.Sizes[name]
		fileName, err := thumb.FromFile(src, hash, cachePath, size.Width, size.Height, thumb.OrientationNormal, size.Options...)
		require.NoError(t, err)
		img, _, err := fs.DecodeImageFile(fileName)
		require.NoError(t, err)

		return img.Bounds().Dx()
	}

	width720 := renditionWidth(t, thumb.Fit720)
	width1280 := renditionWidth(t, thumb.Fit1280)
	require.Greater(t, width1280, width720, "the two renditions must differ for this test to mean anything")

	thumbName, err := thumb.Sizes[thumb.Fit720].FileName(hash, cachePath)
	require.NoError(t, err)

	t.Run("LargeAreaKeepsSmallestRendition", func(t *testing.T) {
		// A face filling the frame is already larger than the template, so the detection
		// thumbnail supplies it without upscaling.
		img, err := ImageFromIdealThumb(thumbName, NewArea("face", 0, 0, 1, 1), Sizes[Tile160])
		require.NoError(t, err)
		assert.Equal(t, width720, img.Bounds().Dx())
	})
	t.Run("SmallAreaUpgradesRendition", func(t *testing.T) {
		// A face covering a fifth of the width needs five times the template width, which
		// only the larger rendition can supply.
		img, err := ImageFromIdealThumb(thumbName, NewArea("face", 0.4, 0.4, 0.2, 0.2), Sizes[Tile160])
		require.NoError(t, err)
		assert.Equal(t, width1280, img.Bounds().Dx())
	})
	t.Run("ReturnsWholeImageNotACrop", func(t *testing.T) {
		// Callers warp from the source themselves, so this must not crop to the area.
		img, err := ImageFromIdealThumb(thumbName, NewArea("face", 0, 0, 1, 1), Sizes[Tile160])
		require.NoError(t, err)
		assert.Greater(t, img.Bounds().Dx(), Sizes[Tile160].Width)
	})
	t.Run("NotAThumbName", func(t *testing.T) {
		// Without a hash prefix there is no ladder to search, so the file is decoded as is.
		img, err := ImageFromIdealThumb(src, NewArea("face", 0.4, 0.4, 0.2, 0.2), Sizes[Tile160])
		require.NoError(t, err)
		assert.Equal(t, 6720, img.Bounds().Dx())
	})
	t.Run("MissingFile", func(t *testing.T) {
		img, err := ImageFromIdealThumb(filepath.Join(cachePath, "missing.jpg"), NewArea("face", 0, 0, 1, 1), Sizes[Tile160])
		assert.Error(t, err)
		assert.Nil(t, img)
	})
}

// TestImageFromThumbCachedSource pins that a reused crop reports no source width. The crop's name
// records its area and dimensions but not what it was drawn from, so any answer would be a
// prediction of today's rendition rather than a record of the one the vector came from.
func TestImageFromThumbCachedSource(t *testing.T) {
	thumbName := "testdata/b/c/c/bccfeaa526a36e19b555fd4ca5e8f767d5604289_720x720_fit.jpg"
	area := NewArea("crop", 0, 0, 1, 1)

	if _, err := os.Stat(thumbName); err != nil {
		t.Skip("thumb fixture not available")
	}

	_, cropName, first, err := ImageFromThumb(thumbName, area, Sizes[Tile50], true)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(cropName) })
	assert.Positive(t, first, "the run that creates the crop measures its source")

	_, _, second, err := ImageFromThumb(thumbName, area, Sizes[Tile50], true)
	require.NoError(t, err)
	assert.Zero(t, second, "the run that reuses it cannot know, and must not guess")

	// The crop cache is keyed on hash, area and size alone, so the UI's own face thumbnails
	// satisfy it. A caller that has to record what it sampled therefore opens the source.
	_, _, source, err := ImageFromSource(thumbName, area, Sizes[Tile50], false)
	require.NoError(t, err)
	assert.Equal(t, first, source, "ImageFromSource measures whether or not a crop is cached")
}

func TestWidestCachedSize(t *testing.T) {
	restore := thumb.SizeCached
	t.Cleanup(func() { thumb.SizeCached = restore })

	t.Run("Default", func(t *testing.T) {
		thumb.SizeCached = 2560
		assert.Equal(t, thumb.Fit2560, WidestCachedSize().Name)
	})
	t.Run("BelowTheNextRendition", func(t *testing.T) {
		// A box is pre-generated only when the limit covers its longer edge, so a limit between
		// two renditions selects the smaller of them.
		thumb.SizeCached = 4095
		assert.Equal(t, thumb.Fit2560, WidestCachedSize().Name)
	})
	t.Run("Raised", func(t *testing.T) {
		thumb.SizeCached = 4096
		assert.Equal(t, thumb.Fit4096, WidestCachedSize().Name)
	})
	t.Run("BelowTheSmallest", func(t *testing.T) {
		// Nothing a crop could be taken from, which the caller has to be able to tell.
		thumb.SizeCached = 320
		assert.Zero(t, WidestCachedSize().Width)
	})
}

func TestUsableSizes(t *testing.T) {
	sizes := UsableSizes()

	require.Len(t, sizes, len(thumbFileNames))
	assert.Equal(t, thumb.Fit720, sizes[0].Name)

	// Ascending, which is what lets a caller take the first that clears its requirement.
	for i := 1; i < len(sizes); i++ {
		assert.GreaterOrEqual(t, sizes[i].Width, sizes[i-1].Width)
	}

	// A copy, so a caller cannot recalibrate the selection the crop path itself walks.
	sizes[0] = thumb.Size{}
	assert.Equal(t, thumb.Fit720, UsableSizes()[0].Name)
}

// TestOpenIdealThumbFile pins that the name comes back with the image. The selection swaps in a
// wider rendition, so a caller reporting the name it passed in describes a file it never read -
// which sent a diagnosis of upscaled face crops to the source photos twice.
func TestOpenIdealThumbFile(t *testing.T) {
	const hash = "bccfeaa526a36e19b555fd4ca5e8f767d5604289"

	requested := filepath.Join("testdata", "b", "c", "c", hash+"_720x720_fit.jpg")
	size := Size{Tile160, Tile160, "Faces", 160, 160, DefaultOptions}

	t.Run("ReportsTheRenditionItRead", func(t *testing.T) {
		// 160/0.32 asks for 500px, which the 720x720 rendition of this portrait picture cannot
		// supply: it is 479px wide, so the selection moves up to the next one.
		img, opened, err := openIdealThumbFile(requested, hash, NewArea("face", 0.5, 0.5, 0.32, 0.32), size)

		require.NoError(t, err)
		require.NotNil(t, img)
		assert.True(t, strings.HasSuffix(opened, "_1280x1024_fit.jpg"), opened)
	})
	t.Run("ReportsTheRequestedName", func(t *testing.T) {
		// A small area is covered by the rendition that was asked for, which is then also the
		// one that was read.
		_, opened, err := openIdealThumbFile(requested, hash, NewArea("face", 0.5, 0.5, 0.9, 0.9), size)

		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(opened, "_720x720_fit.jpg"), opened)
	})
	t.Run("NotAThumbName", func(t *testing.T) {
		// Decoded as passed, so the name reported is the one that was opened here too.
		_, opened, err := openIdealThumbFile(requested, "", NewArea("face", 0.5, 0.5, 0.32, 0.32), size)

		require.NoError(t, err)
		assert.True(t, strings.HasSuffix(opened, "_720x720_fit.jpg"), opened)
	})
}
