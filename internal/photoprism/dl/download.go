package dl

import (
	"context"
)

// Download downloads media from a URL using the specified options and filter (usually a format ID or quality flag).
// If filter is empty, then youtube-dl will use its default format selector.
func Download(
	ctx context.Context,
	rawURL string,
	options Options,
	filter string,
) (*DownloadResult, error) {
	options.noInfoDownload = true
	d, err := NewMetadata(ctx, rawURL, options)
	if err != nil {
		return nil, err
	}
	return d.Download(ctx, filter)
}
