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

	skipIfRefused(t, ydlResultErr)

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

	// Load the playlist and the direct link of the same track first, since their format lists may differ.
	playlistStderr := &bytes.Buffer{}
	r, err := NewMetadata(context.Background(), playlistRawURL, Options{
		StderrFn: func(cmd *exec.Cmd) io.Writer {
			return playlistStderr
		},
	})

	if err != nil {
		if errors.Is(err, ErrPlaylistEmpty) {
			t.Skipf("playlist unavailable: %v", err)
		}
		skipIfRefused(t, err)
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
	entry := r.Info.Entries[playlistIndex]

	if entry.Title != expectedTitleOne {
		t.Errorf("expected title %q got %q", expectedTitleOne, entry.Title)
	}
	if len(entry.Formats) == 0 {
		t.Fatalf("entry %d has no downloadable formats", playlistIndex)
	}

	dlUrl := "https://soundcloud.com/mattheis/b1-mattheis-ben-m"
	directStderr := &bytes.Buffer{}
	direct, err := NewMetadata(context.Background(), dlUrl, Options{
		StderrFn: func(cmd *exec.Cmd) io.Writer {
			return directStderr
		},
	})

	if err != nil {
		skipIfRefused(t, err)
		t.Fatal(err)
	}

	if direct.Info.Title != expectedTitleOne {
		t.Errorf("expected title %q got %q", expectedTitleOne, direct.Info.Title)
	}

	if len(direct.Info.Entries) != 0 {
		t.Errorf("expected %d entries got %d", 0, len(direct.Info.Entries))
	}

	if len(direct.Info.Formats) == 0 {
		t.Fatalf("direct link has no downloadable formats")
	}

	formatID := commonFormatID(entry.Formats, direct.Info.Formats)

	if formatID == "" {
		t.Skipf("playlist entry and direct link share no format: %v, %v", entry.Formats, direct.Info.Formats)
	}

	// Download the file by specifying the playlist index.
	dr, err := r.DownloadWithOptions(context.Background(), DownloadOptions{
		PlaylistIndex: int(entry.PlaylistIndex),
		Filter:        formatID,
	})

	if err != nil {
		skipIfRefused(t, err)
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

	if !strings.Contains(playlistStderr.String(), "Destination") {
		t.Errorf("did not find expected log message on stderr: %q", playlistStderr.String())
	}

	// Download the same file in the same format with the direct link.
	dr, err = direct.Download(context.Background(), formatID)

	if err != nil {
		skipIfRefused(t, err)
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

	if !strings.Contains(directStderr.String(), "Destination") {
		t.Errorf("did not find expected log message on stderr: %q", directStderr.String())
	}

	if directLinkBuf.Len() != playlistBuf.Len() {
		t.Errorf("not the same content size between the playlist index entry and the direct link entry: %d != %d", playlistBuf.Len(), directLinkBuf.Len())
	}

	if !bytes.Equal(directLinkBuf.Bytes(), playlistBuf.Bytes()) {
		t.Error("not the same content between the playlist index entry and the direct link entry")
	}
}

// skipIfRefused skips the test if the download service refused the request, e.g. because it rate limits the host.
func skipIfRefused(t *testing.T, err error) {
	t.Helper()

	if err != nil && (strings.Contains(err.Error(), "HTTP Error 403") || strings.Contains(err.Error(), "HTTP Error 429")) {
		t.Skipf("service refused the request: %v", err)
	}
}

func TestSkipIfRefused(t *testing.T) {
	t.Run("Refused", func(t *testing.T) {
		var skipped bool
		t.Run("Inner", func(t *testing.T) {
			defer func() { skipped = t.Skipped() }()
			skipIfRefused(t, errors.New("Unable to download JSON metadata: HTTP Error 403: Forbidden"))
		})
		if !skipped {
			t.Error("expected the test to be skipped")
		}
	})
	t.Run("OtherError", func(t *testing.T) {
		skipIfRefused(t, errors.New("Unable to download JSON metadata: HTTP Error 404: Not Found"))
		skipIfRefused(t, nil)
	})
}

// commonFormatID returns the ID of the first format in a that b also offers, or an empty string if there is none.
func commonFormatID(a, b []Format) string {
	for _, fa := range a {
		for _, fb := range b {
			if fa.FormatID != "" && fa.FormatID == fb.FormatID {
				return fa.FormatID
			}
		}
	}

	return ""
}

func TestCommonFormatID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		a := []Format{{FormatID: "hls_mp3_0_0"}, {FormatID: "hls_aac_96k"}, {FormatID: "hls_aac_160k"}}
		b := []Format{{FormatID: "hls_aac_160k"}, {FormatID: "hls_aac_96k"}}
		if id := commonFormatID(a, b); id != "hls_aac_96k" {
			t.Errorf("expected %q got %q", "hls_aac_96k", id)
		}
	})
	t.Run("NoCommonFormat", func(t *testing.T) {
		if id := commonFormatID([]Format{{FormatID: "hls_mp3_0_0"}}, []Format{{FormatID: "hls_aac_96k"}}); id != "" {
			t.Errorf("expected no format got %q", id)
		}
	})
	t.Run("EmptyID", func(t *testing.T) {
		if id := commonFormatID([]Format{{}}, []Format{{}}); id != "" {
			t.Errorf("expected no format got %q", id)
		}
	})
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
