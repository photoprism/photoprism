package dl

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// DownloadResult download result
type DownloadResult struct {
	reader     io.ReadCloser
	waitCh     chan struct{}
	waitErr    error         // exit status of yt-dlp, set before waitCh is closed
	stderrDone chan struct{} // closed once stderr has been read to the end
	stderrErr  string        // last "ERROR: " message on stderr, set before stderrDone is closed
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

// Close stops reading, waits for yt-dlp to exit, and returns its error, so a stream that ended early is not taken as complete.
func (dr *DownloadResult) Close() error {
	err := dr.reader.Close()
	<-dr.waitCh
	<-dr.stderrDone
	switch {
	case err != nil:
		return err
	case dr.waitErr == nil:
		return nil
	case dr.stderrErr != "":
		return errors.New(dr.stderrErr)
	default:
		return fmt.Errorf("yt-dlp failed: %w", dr.waitErr)
	}
}

// Formats return all formats
// helper to take care of mixed info and format
func (result Metadata) Formats() []Format {
	if len(result.Info.Formats) > 0 {
		return result.Info.Formats
	}
	return []Format{result.Info.Format}
}
