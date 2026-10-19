package mcp

import (
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// FileFormat represents a single file format report row.
type FileFormat struct {
	Format      string   `json:"format"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Extensions  []string `json:"extensions"`
}

// MetadataField represents a single metadata field report row with the tags it is read from.
type MetadataField struct {
	Field    string   `json:"field"`
	Type     string   `json:"type"`
	Exiftool []string `json:"exiftool"`
	XMP      []string `json:"xmp"`
	DCMI     []string `json:"dcmi"`
}

// MetadataDoc links a metadata namespace to its documentation.
type MetadataDoc struct {
	Namespace string `json:"namespace"`
	URL       string `json:"url"`
}

// ThumbnailSize represents a single thumbnail size report row.
type ThumbnailSize struct {
	Name        string `json:"name"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
	Available   string `json:"available"`
	Usage       string `json:"usage"`
	Video       bool   `json:"video"`
}

// FileFormatsResource contains the JSON payload for the file-formats resource.
type FileFormatsResource struct {
	Edition string       `json:"edition"`
	Items   []FileFormat `json:"items"`
}

// MetadataFieldsResource contains the JSON payload for the metadata-fields resource.
type MetadataFieldsResource struct {
	Edition string          `json:"edition"`
	Items   []MetadataField `json:"items"`
	Docs    []MetadataDoc   `json:"docs"`
}

// ThumbnailSizesResource contains the JSON payload for the thumbnail-sizes resource.
type ThumbnailSizesResource struct {
	Edition string          `json:"edition"`
	Items   []ThumbnailSize `json:"items"`
}

// buildFileFormats returns the rows of the `photoprism show file-formats` report,
// sorted by type and format, with lowercase types and the extensions as a list.
func buildFileFormats() []FileFormat {
	rows, _ := media.Report(fs.ExtensionList.Types(true), true, true, true)

	items := make([]FileFormat, 0, len(rows))

	for _, row := range rows {
		if len(row) < 4 {
			continue
		}

		items = append(items, FileFormat{
			Format:      row[0],
			Description: row[1],
			Type:        strings.ToLower(row[2]),
			Extensions:  splitList(row[3]),
		})
	}

	return items
}

// fileFormatTypes returns the sorted, distinct types of the specified file formats.
func fileFormatTypes(formats []FileFormat) []string {
	types := make([]string, 0, 8)

	for _, f := range formats {
		if f.Type != "" && !slices.Contains(types, f.Type) {
			types = append(types, f.Type)
		}
	}

	slices.Sort(types)

	return slices.Clip(types)
}

// buildMetadataFields returns the rows of the `photoprism show metadata` report in
// the same order, by type and then field.
func buildMetadataFields() []MetadataField {
	rows, _ := meta.Report(&meta.Data{})

	sort.Slice(rows, func(i, j int) bool {
		if rows[i][1] == rows[j][1] {
			return rows[i][0] < rows[j][0]
		}

		return rows[i][1] < rows[j][1]
	})

	items := make([]MetadataField, 0, len(rows))

	for _, row := range rows {
		if len(row) < 5 {
			continue
		}

		items = append(items, MetadataField{
			Field:    row[0],
			Type:     row[1],
			Exiftool: splitList(row[2]),
			XMP:      splitList(row[3]),
			DCMI:     splitList(row[4]),
		})
	}

	return items
}

// buildMetadataDocs returns the documentation links that `photoprism show metadata` prints.
func buildMetadataDocs() []MetadataDoc {
	items := make([]MetadataDoc, 0, len(meta.Docs))

	for _, row := range meta.Docs {
		if len(row) < 2 {
			continue
		}

		items = append(items, MetadataDoc{Namespace: row[0], URL: row[1]})
	}

	return items
}

// buildThumbnailSizes returns the rows of the `photoprism show thumb-sizes` report,
// sorted by width and name, and marks the sizes that `show video-sizes` lists.
func buildThumbnailSizes() []ThumbnailSize {
	rows, _ := thumb.Report(thumb.Sizes.All(), false)

	videoSizes := make(map[string]bool, len(thumb.VideoSizes))

	for _, s := range thumb.VideoSizes {
		videoSizes[s.Name.String()] = true
	}

	items := make([]ThumbnailSize, 0, len(rows))

	for _, row := range rows {
		if len(row) < 6 {
			continue
		}

		width, _ := strconv.Atoi(row[1])
		height, _ := strconv.Atoi(row[2])

		items = append(items, ThumbnailSize{
			Name:        row[0],
			Width:       width,
			Height:      height,
			AspectRatio: strings.ToLower(row[3]),
			Available:   strings.ToLower(row[4]),
			Usage:       row[5],
			Video:       videoSizes[row[0]],
		})
	}

	return items
}

// splitList splits a comma-separated report value into its trimmed, non-empty parts.
func splitList(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}

	return result
}
