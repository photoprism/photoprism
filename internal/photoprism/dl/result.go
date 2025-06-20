package dl

import (
	"context"
	"io"
)

// DownloadResult download result
type DownloadResult struct {
	reader io.ReadCloser
	waitCh chan struct{}
}

// Download format matched by filter (usually a format id or quality designator).
// If filter is empty, then youtube-dl will use its default format selector.
// It's a shortcut of DownloadWithOptions where the options use the default value
func (result Metadata) Download(ctx context.Context, filter string) (*DownloadResult, error) {
	return result.DownloadWithOptions(ctx, DownloadOptions{
		Filter: filter,
	})
}

func (dr *DownloadResult) Read(p []byte) (n int, err error) {
	return dr.reader.Read(p)
}

// Close downloader and wait for resource cleanup
func (dr *DownloadResult) Close() error {
	err := dr.reader.Close()
	<-dr.waitCh
	return err
}

// Formats return all formats
// helper to take care of mixed info and format
func (result Metadata) Formats() []Format {
	if len(result.Info.Formats) > 0 {
		return result.Info.Formats
	}
	return []Format{result.Info.Format}
}
