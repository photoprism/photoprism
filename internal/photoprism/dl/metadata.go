package dl

import (
	"context"
)

// Metadata represents information and options related to a video download URL.
type Metadata struct {
	Info    Info
	RawURL  string
	RawJSON []byte  // saved raw JSON. Used later when downloading
	Options Options // options passed to NewMetadata
}

// NewMetadata downloads metadata for URL
func NewMetadata(ctx context.Context, rawURL string, options Options) (result Metadata, err error) {
	if options.noInfoDownload {
		return Metadata{
			RawURL:  rawURL,
			Options: options,
		}, nil
	}

	info, rawJSON, err := infoFromURL(ctx, rawURL, options)
	if err != nil {
		return Metadata{}, err
	}

	rawJSONCopy := make([]byte, len(rawJSON))
	copy(rawJSONCopy, rawJSON)

	return Metadata{
		Info:    info,
		RawURL:  rawURL,
		RawJSON: rawJSONCopy,
		Options: options,
	}, nil
}
