package ytdl

import (
	"context"
)

// New downloads metadata for URL
func New(ctx context.Context, rawURL string, options Options) (result Result, err error) {
	if options.noInfoDownload {
		return Result{
			RawURL:  rawURL,
			Options: options,
		}, nil
	}

	info, rawJSON, err := infoFromURL(ctx, rawURL, options)
	if err != nil {
		return Result{}, err
	}

	rawJSONCopy := make([]byte, len(rawJSON))
	copy(rawJSONCopy, rawJSON)

	return Result{
		Info:    info,
		RawURL:  rawURL,
		RawJSON: rawJSONCopy,
		Options: options,
	}, nil
}
