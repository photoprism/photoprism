package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

// callTool calls the named tool, requires a successful result, and decodes its structured content.
func callTool[T any](t *testing.T, session *sdkmcp.ClientSession, name string, args map[string]any) T {
	t.Helper()

	res, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: name, Arguments: args})
	require.NoError(t, err)
	require.False(t, res.IsError, "%s returned an error result: %v", name, res.Content)
	require.NotNil(t, res.StructuredContent)

	raw, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)

	var out T
	require.NoError(t, json.Unmarshal(raw, &out))

	return out
}

// callToolError calls the named tool, requires an error result, and returns its message.
func callToolError(t *testing.T, session *sdkmcp.ClientSession, name string, args map[string]any) string {
	t.Helper()

	res, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: name, Arguments: args})
	require.NoError(t, err)
	require.True(t, res.IsError, "%s should return an error result", name)
	require.Len(t, res.Content, 1)

	text, ok := res.Content[0].(*sdkmcp.TextContent)
	require.True(t, ok)

	return text.Text
}

// TestListFileFormatsTool exercises file format lookups, filters, and validation.
func TestListFileFormatsTool(t *testing.T) {
	session := connectTestClient(t)
	all := buildFileFormats()

	t.Run("Success", func(t *testing.T) {
		out := callTool[ListFileFormatsOutput](t, session, "list_file_formats", nil)
		require.Equal(t, len(all), out.Total)
		require.Equal(t, all[:defaultResultLimit], out.Items)
		require.Equal(t, defaultResultLimit, out.Limit)
		require.Zero(t, out.Offset)
		require.True(t, out.Truncated)
		require.Len(t, out.Warnings, 1)
		require.Contains(t, out.Warnings[0], "call again with offset 20")
	})
	t.Run("FilterByType", func(t *testing.T) {
		out := callTool[ListFileFormatsOutput](t, session, "list_file_formats", map[string]any{"type": " RAW ", "limit": maxResultLimit})
		require.NotEmpty(t, out.Items)
		require.False(t, out.Truncated)
		require.Empty(t, out.Warnings)

		raw := 0
		for _, item := range all {
			if item.Type == "raw" {
				raw++
			}
		}
		require.Equal(t, raw, out.Total)

		for _, item := range out.Items {
			require.Equal(t, "raw", item.Type)
		}
	})
	t.Run("QueryExtension", func(t *testing.T) {
		out := callTool[ListFileFormatsOutput](t, session, "list_file_formats", map[string]any{"query": ".DNG"})
		require.Equal(t, 1, out.Total)
		require.Equal(t, "DNG", out.Items[0].Format)
	})
	t.Run("QueryDescription", func(t *testing.T) {
		out := callTool[ListFileFormatsOutput](t, session, "list_file_formats", map[string]any{"query": "digital negative"})
		require.NotZero(t, out.Total)
	})
	t.Run("NoResults", func(t *testing.T) {
		out := callTool[ListFileFormatsOutput](t, session, "list_file_formats", map[string]any{"query": "definitely-no-match", "type": "video"})
		require.NotNil(t, out.Items)
		require.Empty(t, out.Items)
		require.Zero(t, out.Total)
		require.False(t, out.Truncated)
	})
	t.Run("InvalidType", func(t *testing.T) {
		msg := callToolError(t, session, "list_file_formats", map[string]any{"type": "photo"})
		require.Equal(t, "type must be one of archive, document, image, raw, sidecar, vector, video, or all", msg)
	})
	t.Run("QueryTooLong", func(t *testing.T) {
		msg := callToolError(t, session, "list_file_formats", map[string]any{"query": strings.Repeat("a", maxQueryLength+1)})
		require.Contains(t, msg, "query must not exceed 200 characters")
	})
}

// TestListMetadataFieldsTool exercises metadata field lookups and validation.
func TestListMetadataFieldsTool(t *testing.T) {
	session := connectTestClient(t)
	all := buildMetadataFields()

	t.Run("Success", func(t *testing.T) {
		out := callTool[ListMetadataFieldsOutput](t, session, "list_metadata_fields", map[string]any{"limit": maxResultLimit})
		require.Equal(t, all, out.Items)
		require.Equal(t, len(all), out.Total)
		require.False(t, out.Truncated)
	})
	t.Run("QueryField", func(t *testing.T) {
		out := callTool[ListMetadataFieldsOutput](t, session, "list_metadata_fields", map[string]any{"query": "takenat"})
		require.NotZero(t, out.Total)
		require.Equal(t, "TakenAt", out.Items[0].Field)
	})
	t.Run("QueryTag", func(t *testing.T) {
		out := callTool[ListMetadataFieldsOutput](t, session, "list_metadata_fields", map[string]any{"query": "exif:GPSAltitude"})
		require.Equal(t, 1, out.Total)
		require.Equal(t, "Altitude", out.Items[0].Field)
	})
	t.Run("NoResults", func(t *testing.T) {
		out := callTool[ListMetadataFieldsOutput](t, session, "list_metadata_fields", map[string]any{"query": "definitely-no-match"})
		require.Empty(t, out.Items)
		require.Zero(t, out.Total)
	})
	t.Run("QueryTooLong", func(t *testing.T) {
		msg := callToolError(t, session, "list_metadata_fields", map[string]any{"query": strings.Repeat("a", maxQueryLength+1)})
		require.Contains(t, msg, "query must not exceed 200 characters")
	})
}

// TestListThumbnailSizesTool exercises thumbnail size lookups and validation.
func TestListThumbnailSizesTool(t *testing.T) {
	session := connectTestClient(t)
	all := buildThumbnailSizes()

	t.Run("Default", func(t *testing.T) {
		out := callTool[ListThumbnailSizesOutput](t, session, "list_thumbnail_sizes", nil)
		require.Equal(t, all, out.Items)
		require.Equal(t, len(all), out.Total)
	})
	t.Run("Image", func(t *testing.T) {
		out := callTool[ListThumbnailSizesOutput](t, session, "list_thumbnail_sizes", map[string]any{"type": "image"})
		require.Equal(t, all, out.Items)
	})
	t.Run("Video", func(t *testing.T) {
		out := callTool[ListThumbnailSizesOutput](t, session, "list_thumbnail_sizes", map[string]any{"type": "Video"})
		require.Equal(t, 10, out.Total)
		require.Len(t, out.Items, 10)

		for _, item := range out.Items {
			require.True(t, item.Video)
			require.Positive(t, item.Width)
			require.Positive(t, item.Height)
		}
	})
	t.Run("InvalidType", func(t *testing.T) {
		msg := callToolError(t, session, "list_thumbnail_sizes", map[string]any{"type": "thumb"})
		require.Equal(t, "type must be one of image, video, or all", msg)
	})
}

// TestReferenceToolPagination checks the limit and offset handling of the paginated reference tools.
func TestReferenceToolPagination(t *testing.T) {
	session := connectTestClient(t)
	total := len(buildFileFormats())

	t.Run("LimitBounds", func(t *testing.T) {
		for limit, expected := range map[int]int{0: 20, 1: 1, 50: 50, 51: 50} {
			out := callTool[ListFileFormatsOutput](t, session, "list_file_formats", map[string]any{"limit": limit})
			require.Equal(t, expected, out.Limit, "limit %d", limit)
			require.Len(t, out.Items, expected, "limit %d", limit)
			require.Equal(t, total, out.Total)
			require.True(t, out.Truncated)
		}
	})
	t.Run("NegativeLimit", func(t *testing.T) {
		msg := callToolError(t, session, "list_metadata_fields", map[string]any{"limit": -1})
		require.Contains(t, msg, "limit must be greater than 0")
	})
	t.Run("NegativeOffset", func(t *testing.T) {
		msg := callToolError(t, session, "list_file_formats", map[string]any{"offset": -1})
		require.Contains(t, msg, "offset must be 0 or greater")

		msg = callToolError(t, session, "list_metadata_fields", map[string]any{"offset": -5})
		require.Contains(t, msg, "offset must be 0 or greater")
	})
	t.Run("LastPage", func(t *testing.T) {
		out := callTool[ListFileFormatsOutput](t, session, "list_file_formats", map[string]any{"limit": 50, "offset": 50})
		require.Len(t, out.Items, total-50)
		require.Equal(t, 50, out.Offset)
		require.False(t, out.Truncated)
		require.Empty(t, out.Warnings)
	})
	t.Run("MaxOffset", func(t *testing.T) {
		out := callTool[ListFileFormatsOutput](t, session, "list_file_formats", map[string]any{"offset": int64(9223372036854774784)})
		require.NotNil(t, out.Items)
		require.Empty(t, out.Items)
		require.Equal(t, total, out.Total)
		require.False(t, out.Truncated)
	})
	t.Run("OffsetPastEnd", func(t *testing.T) {
		out := callTool[ListMetadataFieldsOutput](t, session, "list_metadata_fields", map[string]any{"offset": 1000})
		require.NotNil(t, out.Items)
		require.Empty(t, out.Items)
		require.Equal(t, len(buildMetadataFields()), out.Total)
		require.Equal(t, 1000, out.Offset)
		require.False(t, out.Truncated)
	})
	t.Run("Pages", func(t *testing.T) {
		var items []FileFormat

		for offset := 0; offset < total; offset += 7 {
			out := callTool[ListFileFormatsOutput](t, session, "list_file_formats", map[string]any{"limit": 7, "offset": offset})
			items = append(items, out.Items...)
		}

		require.Equal(t, buildFileFormats(), items)
	})
}

// TestValidateOffset covers accepted and rejected offsets.
func TestValidateOffset(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		offset, err := validateOffset(0)
		require.NoError(t, err)
		require.Zero(t, offset)

		offset, err = validateOffset(30)
		require.NoError(t, err)
		require.Equal(t, 30, offset)
	})
	t.Run("Negative", func(t *testing.T) {
		_, err := validateOffset(-1)
		require.EqualError(t, err, "offset must be 0 or greater")
	})
}

// TestPageOf covers full, partial, and empty pages.
func TestPageOf(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	t.Run("FirstPage", func(t *testing.T) {
		page, more := pageOf(items, 2, 0)
		require.Equal(t, []int{1, 2}, page)
		require.True(t, more)
	})
	t.Run("LastPage", func(t *testing.T) {
		page, more := pageOf(items, 2, 4)
		require.Equal(t, []int{5}, page)
		require.False(t, more)
	})
	t.Run("ExactEnd", func(t *testing.T) {
		page, more := pageOf(items, 5, 0)
		require.Equal(t, items, page)
		require.False(t, more)
	})
	t.Run("PastEnd", func(t *testing.T) {
		page, more := pageOf(items, 2, 5)
		require.NotNil(t, page)
		require.Empty(t, page)
		require.False(t, more)
	})
	t.Run("Copy", func(t *testing.T) {
		page, _ := pageOf(items, 2, 0)
		page[0] = 99
		require.Equal(t, 1, items[0])
	})
}

// TestPageWarnings checks that a warning names the next offset only when more items follow.
func TestPageWarnings(t *testing.T) {
	t.Run("Truncated", func(t *testing.T) {
		require.Equal(t, []string{"returned 20 of 61 matches starting at offset 20; call again with offset 40 to retrieve more"}, pageWarnings(20, 61, 20, true))
	})
	t.Run("Complete", func(t *testing.T) {
		require.Nil(t, pageWarnings(21, 61, 40, false))
	})
}

// TestValidateFileFormatType covers the default, accepted, and rejected types.
func TestValidateFileFormatType(t *testing.T) {
	types := []string{"image", "raw"}

	t.Run("Default", func(t *testing.T) {
		v, err := validateFileFormatType(" ", types)
		require.NoError(t, err)
		require.Equal(t, "all", v)
	})
	t.Run("Success", func(t *testing.T) {
		v, err := validateFileFormatType("Image", types)
		require.NoError(t, err)
		require.Equal(t, "image", v)
	})
	t.Run("Invalid", func(t *testing.T) {
		_, err := validateFileFormatType("video", types)
		require.EqualError(t, err, "type must be one of image, raw, or all")
	})
	t.Run("SharedTypes", func(t *testing.T) {
		shared := make([]string, 2, 3)
		copy(shared, types)

		_, err := validateFileFormatType("video", shared)
		require.Error(t, err)
		require.Equal(t, []string{"image", "raw", ""}, shared[:cap(shared)], "the accepted types must not be appended in place")
	})
}

// TestValidateThumbnailSizeType covers the default, accepted, and rejected types.
func TestValidateThumbnailSizeType(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		v, err := validateThumbnailSizeType("")
		require.NoError(t, err)
		require.Equal(t, "all", v)
	})
	t.Run("Success", func(t *testing.T) {
		v, err := validateThumbnailSizeType(" VIDEO ")
		require.NoError(t, err)
		require.Equal(t, "video", v)
	})
	t.Run("Invalid", func(t *testing.T) {
		_, err := validateThumbnailSizeType("fit")
		require.EqualError(t, err, "type must be one of image, video, or all")
	})
}

// TestMatchesFileFormatQuery covers matches on each searched field.
func TestMatchesFileFormatQuery(t *testing.T) {
	format := FileFormat{Format: "DNG", Description: "Adobe Digital Negative", Type: "raw", Extensions: []string{".dng"}}

	t.Run("Match", func(t *testing.T) {
		require.True(t, matchesFileFormatQuery(format, ""))
		require.True(t, matchesFileFormatQuery(format, "dng"))
		require.True(t, matchesFileFormatQuery(format, "negative"))
		require.True(t, matchesFileFormatQuery(FileFormat{Format: "X", Extensions: []string{".foo"}}, ".foo"))
	})
	t.Run("NoMatch", func(t *testing.T) {
		require.False(t, matchesFileFormatQuery(format, "raw"))
	})
}

// TestMatchesMetadataFieldQuery covers matches on the field name and each tag list.
func TestMatchesMetadataFieldQuery(t *testing.T) {
	field := MetadataField{Field: "Title", Type: "text", Exiftool: []string{"Headline"}, XMP: []string{"dc:title"}, DCMI: []string{"title"}}

	t.Run("Match", func(t *testing.T) {
		require.True(t, matchesMetadataFieldQuery(field, ""))
		require.True(t, matchesMetadataFieldQuery(field, "titl"))
		require.True(t, matchesMetadataFieldQuery(field, "headline"))
		require.True(t, matchesMetadataFieldQuery(MetadataField{Field: "A", XMP: []string{"dc:foo"}}, "dc:foo"))
		require.True(t, matchesMetadataFieldQuery(MetadataField{Field: "A", DCMI: []string{"bar"}}, "bar"))
	})
	t.Run("NoMatch", func(t *testing.T) {
		require.False(t, matchesMetadataFieldQuery(field, "text"))
	})
}

// TestReadOnly checks that each call returns separate read-only annotations.
func TestReadOnly(t *testing.T) {
	a, b := readOnly(), readOnly()
	require.True(t, a.ReadOnlyHint)
	require.NotSame(t, a, b)
}
