package dl

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
	"testing"
)

const (
	testVideoRawURL          = "https://www.youtube.com/watch?v=fD6VYfy3B2s"
	playlistRawURL           = "https://soundcloud.com/mattheis/sets/kindred-phenomena"
	channelRawURL            = "https://www.youtube.com/channel/UCHDm-DKoMyJxKVgwGmuTaQA"
	subtitlesTestVideoRawURL = "https://www.youtube.com/watch?v=QRS8MkLhQmM"
)

func TestPlaylist(t *testing.T) {
	t.Skip("skipping test because playlist URL is unreliable.")

	ydlResult, ydlResultErr := NewMetadata(context.Background(), playlistRawURL, Options{
		Type:              TypePlaylist,
		DownloadThumbnail: false,
	})

	if ydlResultErr != nil {
		t.Errorf("failed to download: %s", ydlResultErr)
	}

	expectedTitle := "Kindred Phenomena"
	if ydlResult.Info.Title != expectedTitle {
		t.Errorf("expected title %q got %q", expectedTitle, ydlResult.Info.Title)
	}

	expectedEntries := 8
	if len(ydlResult.Info.Entries) != expectedEntries {
		t.Errorf("expected %d entries got %d", expectedEntries, len(ydlResult.Info.Entries))
	}

	expectedTitleOne := "A1 Mattheis - Herds"
	if ydlResult.Info.Entries[0].Title != expectedTitleOne {
		t.Errorf("expected title %q got %q", expectedTitleOne, ydlResult.Info.Entries[0].Title)
	}
}

func TestChannel(t *testing.T) {
	t.Skip("skip youtube for now")

	ydlResult, ydlResultErr := NewMetadata(
		context.Background(),
		channelRawURL,
		Options{
			Type:              TypeChannel,
			DownloadThumbnail: false,
		},
	)

	if ydlResultErr != nil {
		t.Errorf("failed to download: %s", ydlResultErr)
	}

	expectedTitle := "Simon Yapp"
	if ydlResult.Info.Title != expectedTitle {
		t.Errorf("expected title %q got %q", expectedTitle, ydlResult.Info.Title)
	}

	expectedEntries := 5
	if len(ydlResult.Info.Entries) != expectedEntries {
		t.Errorf("expected %d entries got %d", expectedEntries, len(ydlResult.Info.Entries))
	}

	expectedTitleOne := "#RNLI Shoreham #LifeBoat demo of launch."
	if ydlResult.Info.Entries[0].Title != expectedTitleOne {
		t.Errorf("expected title %q got %q", expectedTitleOne, ydlResult.Info.Entries[0].Title)
	}
}

func TestUnsupportedURL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped download test in short mode")
	}

	_, ydlResultErr := NewMetadata(context.Background(), "https://www.google.com", Options{})
	if ydlResultErr == nil {
		t.Errorf("expected unsupported url")
	}
	expectedErrPrefix := "Unsupported URL:"
	if ydlResultErr != nil && !strings.HasPrefix(ydlResultErr.Error(), expectedErrPrefix) {
		t.Errorf("expected error prefix %q got %q", expectedErrPrefix, ydlResultErr.Error())

	}
}

func TestPlaylistWithPrivateVideo(t *testing.T) {
	t.Skip("skip youtube for now")

	plRawURL := "https://www.youtube.com/playlist?list=PLX0g748fkegS54oiDN4AXKl7BR7mLIydP"
	ydlResult, ydlResultErr := NewMetadata(context.Background(), plRawURL, Options{
		Type:              TypePlaylist,
		DownloadThumbnail: false,
	})

	if ydlResultErr != nil {
		t.Errorf("failed to download: %s", ydlResultErr)
	}

	expectedLen := 2
	actualLen := len(ydlResult.Info.Entries)
	if expectedLen != actualLen {
		t.Errorf("expected len %d got %d", expectedLen, actualLen)
	}
}

func TestSubtitles(t *testing.T) {
	t.Skip("skip youtube for now")

	ydlResult, ydlResultErr := NewMetadata(
		context.Background(),
		subtitlesTestVideoRawURL,
		Options{
			DownloadSubtitles: true,
		})

	if ydlResultErr != nil {
		t.Errorf("failed to download: %s", ydlResultErr)
	}

	for _, subtitles := range ydlResult.Info.Subtitles {
		for _, subtitle := range subtitles {
			if subtitle.Ext == "" {
				t.Errorf("%s: %s: expected extension", ydlResult.Info.URL, subtitle.Language)
			}
			if subtitle.Language == "" {
				t.Errorf("%s: %s: expected language", ydlResult.Info.URL, subtitle.Language)
			}
			if subtitle.URL == "" {
				t.Errorf("%s: %s: expected url", ydlResult.Info.URL, subtitle.Language)
			}
			if len(subtitle.Bytes) == 0 {
				t.Errorf("%s: %s: expected bytes", ydlResult.Info.URL, subtitle.Language)
			}
		}
	}
}

func TestErrorNotASingleEntry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped download test in short mode")
	}

	_, ydlResultErr := NewMetadata(context.Background(), playlistRawURL, Options{
		Type:              TypeSingle,
		DownloadThumbnail: false,
	})

	if ydlResultErr != ErrNotASingleEntry {
		t.Fatalf("expected is single entry error, got %s", ydlResultErr)
	}
}

func TestInvalidOptionTypeField(t *testing.T) {
	_, err := NewMetadata(context.Background(), playlistRawURL, Options{
		Type: 42,
	})
	if err == nil {
		t.Error("should have failed")
	}
}

func TestDownloadPlaylistEntry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped download test in short mode")
	}

	// Download file by specifying the playlist index
	stderrBuf := &bytes.Buffer{}
	r, err := NewMetadata(context.Background(), playlistRawURL, Options{
		StderrFn: func(cmd *exec.Cmd) io.Writer {
			return stderrBuf
		},
	})

	if err != nil {
		if errors.Is(err, ErrPlaylistEmpty) {
			t.Skipf("playlist unavailable: %v", err)
		}
		t.Fatal(err)
	}

	expectedTitle := "Kindred Phenomena"
	if r.Info.Title != expectedTitle {
		t.Errorf("expected title %q got %q", expectedTitle, r.Info.Title)
	}

	expectedEntries := 8

	if len(r.Info.Entries) != expectedEntries {
		t.Fatalf("expected %d entries got %d", expectedEntries, len(r.Info.Entries))
	}

	expectedTitleOne := "B1 Mattheis - Ben M"
	playlistIndex := 2

	if r.Info.Entries[playlistIndex].Title != expectedTitleOne {
		t.Errorf("expected title %q got %q", expectedTitleOne, r.Info.Entries[playlistIndex].Title)
	}
	if len(r.Info.Entries[playlistIndex].Formats) == 0 {
		t.Fatalf("entry %d has no downloadable formats", playlistIndex)
	}

	dr, err := r.DownloadWithOptions(context.Background(), DownloadOptions{
		PlaylistIndex: int(r.Info.Entries[playlistIndex].PlaylistIndex),
		Filter:        r.Info.Entries[playlistIndex].Formats[0].FormatID,
	})

	if err != nil {
		t.Fatal(err)
	}

	playlistBuf := &bytes.Buffer{}
	n, err := io.Copy(playlistBuf, dr)

	if err != nil {
		t.Fatal(err)
	}

	_ = dr.Close()

	if n != int64(playlistBuf.Len()) {
		t.Errorf("copy n not equal to download buffer: %d!=%d", n, playlistBuf.Len())
	}

	if n < 10000 {
		t.Errorf("should have copied at least 10000 bytes: %d", n)
	}

	if !strings.Contains(stderrBuf.String(), "Destination") {
		t.Errorf("did not find expected log message on stderr: %q", stderrBuf.String())
	}

	// Download the same file but with the direct link
	dlUrl := "https://soundcloud.com/mattheis/b1-mattheis-ben-m"

	stderrBuf = &bytes.Buffer{}

	r, err = NewMetadata(context.Background(), dlUrl, Options{
		StderrFn: func(cmd *exec.Cmd) io.Writer {
			return stderrBuf
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	if r.Info.Title != expectedTitleOne {
		t.Errorf("expected title %q got %q", expectedTitleOne, r.Info.Title)
	}

	expectedEntries = 0

	if len(r.Info.Entries) != expectedEntries {
		t.Errorf("expected %d entries got %d", expectedEntries, len(r.Info.Entries))
	}

	if len(r.Info.Formats) == 0 {
		t.Fatalf("direct link has no downloadable formats")
	}

	dr, err = r.Download(context.Background(), r.Info.Formats[0].FormatID)

	if err != nil {
		t.Fatal(err)
	}

	directLinkBuf := &bytes.Buffer{}
	n, err = io.Copy(directLinkBuf, dr)

	if err != nil {
		t.Fatal(err)
	}

	_ = dr.Close()

	if n != int64(directLinkBuf.Len()) {
		t.Errorf("copy n not equal to download buffer: %d!=%d", n, directLinkBuf.Len())
	}

	if n < 10000 {
		t.Errorf("should have copied at least 10000 bytes: %d", n)
	}

	if !strings.Contains(stderrBuf.String(), "Destination") {
		t.Errorf("did not find expected log message on stderr: %q", stderrBuf.String())
	}

	if directLinkBuf.Len() != playlistBuf.Len() {
		t.Errorf("not the same content size between the playlist index entry and the direct link entry: %d != %d", playlistBuf.Len(), directLinkBuf.Len())
	}

	if !bytes.Equal(directLinkBuf.Bytes(), playlistBuf.Bytes()) {
		t.Error("not the same content between the playlist index entry and the direct link entry")
	}
}

func TestFormatDownloadError(t *testing.T) {
	t.Skip("test URL broken")

	ydl, ydlErr := NewMetadata(
		context.Background(),
		"https://www.reddit.com/r/newsbabes/s/92rflI0EB0",
		Options{},
	)

	if ydlErr != nil {
		// reddit seems to not like github action hosts
		if strings.Contains(ydlErr.Error(), "HTTPError 403: Blocked") {
			t.Skip()
		}
		t.Error(ydlErr)
	}

	// no pre-muxed audio/video format available
	_, ytDlErr := ydl.Download(context.Background(), "best")
	expectedErr := "Requested format is not available"

	if ydlErr != nil && !strings.Contains(ytDlErr.Error(), expectedErr) {
		t.Errorf("expected error prefix %q got %q", expectedErr, ytDlErr.Error())
	}
}
