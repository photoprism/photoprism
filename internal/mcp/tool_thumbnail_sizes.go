package mcp

import (
	"context"
	"fmt"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListThumbnailSizesInput defines the inputs for the thumbnail size lookup tool.
type ListThumbnailSizesInput struct {
	Type string `json:"type,omitempty" jsonschema:"optional size type: image, video, or all"`
}

// ListThumbnailSizesOutput defines the structured output for the thumbnail size lookup tool.
type ListThumbnailSizesOutput struct {
	Items []ThumbnailSize `json:"items"`
	Total int             `json:"total"`
}

// listThumbnailSizes validates the caller's input and returns every thumbnail size,
// or only the video sizes if the type is "video".
func listThumbnailSizes(_ context.Context, _ *sdkmcp.CallToolRequest, input ListThumbnailSizesInput, data *Dataset) (*sdkmcp.CallToolResult, ListThumbnailSizesOutput, error) {
	sizeType, err := validateThumbnailSizeType(input.Type)
	if err != nil {
		return nil, ListThumbnailSizesOutput{}, err
	}

	items := make([]ThumbnailSize, 0, len(data.ThumbnailSizes))

	for _, size := range data.ThumbnailSizes {
		if sizeType != "video" || size.Video {
			items = append(items, size)
		}
	}

	return nil, ListThumbnailSizesOutput{Items: items, Total: len(items)}, nil
}

// validateThumbnailSizeType normalizes the requested size type. An empty input
// defaults to "all", which returns the same sizes as "image".
func validateThumbnailSizeType(sizeType string) (string, error) {
	switch normalized := strings.TrimSpace(strings.ToLower(sizeType)); normalized {
	case "", "all":
		return "all", nil
	case "image", "video":
		return normalized, nil
	default:
		return "", fmt.Errorf("type must be one of image, video, or all")
	}
}
