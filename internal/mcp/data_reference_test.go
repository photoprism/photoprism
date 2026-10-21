package mcp

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// TestBuildFileFormats compares the file formats with the rows of the report behind
// `photoprism show file-formats`.
func TestBuildFileFormats(t *testing.T) {
	rows, _ := media.Report(fs.ExtensionList.Types(true), true, true, true)
	items := buildFileFormats()

	require.Len(t, items, len(rows))

	for i, item := range items {
		require.Equal(t, rows[i][0], item.Format)
		require.Equal(t, rows[i][1], item.Description)
		require.Equal(t, strings.ToLower(rows[i][2]), item.Type)
		require.Equal(t, rows[i][3], strings.Join(item.Extensions, ", "))
		require.NotEmpty(t, item.Type, "item %d (%s) must have a type", i, item.Format)
		require.NotEmpty(t, item.Extensions, "item %d (%s) must have extensions", i, item.Format)
	}

	jpeg := findItem(t, items, func(f FileFormat) bool { return f.Format == "JPEG" || f.Format == "JPG" })
	require.Greater(t, len(jpeg.Extensions), 1)

	dng := findItem(t, items, func(f FileFormat) bool { return f.Format == "DNG" })
	require.Equal(t, "raw", dng.Type)
	require.Equal(t, []string{".dng"}, dng.Extensions)
}

// TestFileFormatTypes covers sorting, deduplication, and skipping empty types.
func TestFileFormatTypes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		formats := []FileFormat{{Type: "video"}, {Type: "image"}, {Type: "video"}, {Type: ""}, {Type: "raw"}}
		require.Equal(t, []string{"image", "raw", "video"}, fileFormatTypes(formats))
	})
	t.Run("Empty", func(t *testing.T) {
		require.Empty(t, fileFormatTypes(nil))
	})
	t.Run("Clipped", func(t *testing.T) {
		types := fileFormatTypes([]FileFormat{{Type: "image"}, {Type: "raw"}})
		require.Equal(t, len(types), cap(types))
	})
	t.Run("Report", func(t *testing.T) {
		require.Equal(t, []string{"archive", "document", "image", "raw", "sidecar", "vector", "video"}, fileFormatTypes(buildFileFormats()))
	})
}

// TestBuildMetadataFields compares the metadata fields with the rows of the report behind
// `photoprism show metadata` and checks that they are sorted by type and field.
func TestBuildMetadataFields(t *testing.T) {
	rows, _ := meta.Report(&meta.Data{})
	items := buildMetadataFields()

	require.Len(t, items, len(rows))
	require.True(t, sort.SliceIsSorted(items, func(i, j int) bool {
		if items[i].Type == items[j].Type {
			return items[i].Field < items[j].Field
		}

		return items[i].Type < items[j].Type
	}))

	sort.Slice(rows, func(i, j int) bool {
		if rows[i][1] == rows[j][1] {
			return rows[i][0] < rows[j][0]
		}

		return rows[i][1] < rows[j][1]
	})

	for i, item := range items {
		require.Equal(t, rows[i][0], item.Field)
		require.Equal(t, rows[i][1], item.Type)
		require.Equal(t, rows[i][2], strings.Join(item.Exiftool, ", "))
		require.Equal(t, rows[i][3], strings.Join(item.XMP, ", "))
		require.Equal(t, rows[i][4], strings.Join(item.DCMI, ", "))
	}

	title := findItem(t, items, func(f MetadataField) bool { return f.Field == "Title" })
	require.NotEmpty(t, title.DCMI)

	takenAt := findItem(t, items, func(f MetadataField) bool { return f.Field == "TakenAt" })
	require.Equal(t, "timestamp", takenAt.Type)
	require.Contains(t, takenAt.Exiftool, "DateTimeOriginal")
	require.Contains(t, takenAt.XMP, "exif:DateTimeOriginal")
}

// TestBuildMetadataDocs checks that every documentation link is returned.
func TestBuildMetadataDocs(t *testing.T) {
	items := buildMetadataDocs()

	require.Len(t, items, len(meta.Docs))

	for i, item := range items {
		require.Equal(t, meta.Docs[i][0], item.Namespace)
		require.Equal(t, meta.Docs[i][1], item.URL)
	}
}

// TestBuildThumbnailSizes compares the thumbnail sizes with the rows of the report behind
// `photoprism show thumb-sizes` and checks the video size marker.
func TestBuildThumbnailSizes(t *testing.T) {
	rows, _ := thumb.Report(thumb.Sizes.All(), false)
	items := buildThumbnailSizes()

	require.Len(t, items, len(rows))
	require.Len(t, items, len(thumb.Sizes))

	videos := 0

	for i, item := range items {
		require.Equal(t, rows[i][0], item.Name)
		require.Positive(t, item.Width, "item %d (%s) must have a width", i, item.Name)
		require.Positive(t, item.Height, "item %d (%s) must have a height", i, item.Name)

		if item.Video {
			videos++
		}
	}

	require.Equal(t, len(thumb.VideoSizes), videos)

	fit1920 := findItem(t, items, func(s ThumbnailSize) bool { return s.Name == thumb.Fit1920.String() })
	require.Equal(t, ThumbnailSize{Name: "fit_1920", Width: 1920, Height: 1200, AspectRatio: "preserved", Available: "on-demand", Usage: "Full HD", Video: true}, fit1920)

	tile500 := findItem(t, items, func(s ThumbnailSize) bool { return s.Name == thumb.Tile500.String() })
	require.Equal(t, "1:1", tile500.AspectRatio)
	require.Equal(t, "always", tile500.Available)
	require.False(t, tile500.Video)
}

// TestSplitList covers trimming and dropping empty parts.
func TestSplitList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		require.Equal(t, []string{"a", "b", "c"}, splitList("a, b,c"))
	})
	t.Run("Empty", func(t *testing.T) {
		require.Empty(t, splitList(""))
		require.Empty(t, splitList(" , "))
	})
}

// findItem returns the first item that matches the specified function, failing the test if there is none.
func findItem[T any](t *testing.T, items []T, match func(T) bool) T {
	t.Helper()

	for _, item := range items {
		if match(item) {
			return item
		}
	}

	require.FailNow(t, "item not found")

	var zero T

	return zero
}
