package dl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/photoprism/photoprism/pkg/fs"
)

const (
	testVideoRawURL          = "https://www.youtube.com/watch?v=fD6VYfy3B2s"
	playlistRawURL           = "https://soundcloud.com/mattheis/sets/kindred-phenomena"
	channelRawURL            = "https://www.youtube.com/channel/UCHDm-DKoMyJxKVgwGmuTaQA"
	subtitlesTestVideoRawURL = "https://www.youtube.com/watch?v=QRS8MkLhQmM"
)

func TestParseInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	for _, c := range []struct {
		url           string
		expectedTitle string
	}{
		{"https://soundcloud.com/avalonemerson/avalon-emerson-live-at-printworks-london-march-2017",
			"Avalon Emerson Live at Printworks London 2017"},
		{"https://www.infoq.com/presentations/Simple-Made-Easy",
			"Simple Made Easy - InfoQ"},
		{testVideoRawURL,
			"Cinematic Epic Deep Trailer - Background Music for Trailers and Film"},
	} {
		t.Run(c.url, func(t *testing.T) {
			ctx, cancelFn := context.WithCancel(context.Background())
			ydlResult, err := NewMetadata(ctx, c.url, Options{
				DownloadThumbnail: true,
			})
			if err != nil {
				cancelFn()
				t.Errorf("failed to parse: %v", err)
				return
			}
			cancelFn()

			yi := ydlResult.Info
			results := ydlResult.Formats()

			if yi.Title != c.expectedTitle {
				t.Errorf("expected title %q got %q", c.expectedTitle, yi.Title)
			}

			if yi.Thumbnail != "" && len(yi.ThumbnailBytes) == 0 {
				t.Errorf("expected thumbnail bytes")
			}

			var dummy map[string]interface{}
			if err := json.Unmarshal(ydlResult.RawJSON, &dummy); err != nil {
				t.Errorf("failed to parse RawJSON")
			}

			if len(results) == 0 {
				t.Errorf("expected formats")
			}

			for _, f := range results {
				if f.FormatID == "" {
					t.Errorf("expected to have FormatID")
				}
				if f.Ext == "" {
					t.Errorf("expected to have Ext")
				}
				if (f.ACodec == "" || f.ACodec == "none") &&
					(f.VCodec == "" || f.VCodec == "none") &&
					f.Ext == "" {
					t.Errorf("expected to have some media: audio %q video %q ext %q", f.ACodec, f.VCodec, f.Ext)
				}
			}
		})
	}
}

func TestPlaylist(t *testing.T) {
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
		t.Skip("skipping test in short mode.")
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

func TestDownloadSections(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	fileName := fs.Abs("./testdata/duration_test_file")
	duration := 5

	defer func() {
		_ = os.Remove(fileName)
	}()

	cmd := exec.Command(FindFFmpegBin(), "-version")
	_, err := cmd.Output()

	if err != nil {
		t.Errorf("failed to check ffmpeg installed: %s", err)
	}

	ydlResult, ydlResultErr := NewMetadata(
		context.Background(),
		testVideoRawURL,
		Options{
			DownloadSections: fmt.Sprintf("*0:0-0:%d", duration),
		})

	if ydlResult.Options.DownloadSections != "*0:0-0:5" {
		t.Errorf("failed to setup --download-sections")
	}

	if ydlResultErr != nil {
		t.Errorf("failed to download: %s", ydlResultErr)
	}

	dr, err := ydlResult.Download(context.Background(), "")

	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(fileName)

	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	_, err = io.Copy(f, dr)
	if err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command(FindFFprobeBin(), "-v", "quiet", "-show_entries", "format=duration", fileName)
	stdout, err := cmd.Output()

	if err != nil {
		t.Fatal(err)
	}

	var gotDurationString string
	output := string(stdout)
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "duration") {
			if d, found := strings.CutPrefix(line, "duration="); found {
				gotDurationString = d
			}
		}
	}

	gotDuration, err := strconv.ParseFloat(gotDurationString, 32)

	if err != nil {
		t.Fatal(err)
	}

	seconds := int(gotDuration)

	if seconds != duration {
		t.Fatalf("did not get expected duration of %d, but got %d", duration, seconds)
	}

	_ = dr.Close()
}

func TestErrorNotAPlaylist(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	_, ydlResultErr := NewMetadata(context.Background(), testVideoRawURL, Options{
		Type:              TypePlaylist,
		DownloadThumbnail: false,
	})
	if ydlResultErr != ErrNotAPlaylist {
		t.Errorf("expected is playlist error, got %s", ydlResultErr)
	}
}

func TestErrorNotASingleEntry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	_, ydlResultErr := NewMetadata(context.Background(), playlistRawURL, Options{
		Type:              TypeSingle,
		DownloadThumbnail: false,
	})

	if ydlResultErr != ErrNotASingleEntry {
		t.Fatalf("expected is single entry error, got %s", ydlResultErr)
	}
}

func TestOptionDownloader(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ydlResult, ydlResultErr := NewMetadata(
		context.Background(),
		testVideoRawURL,
		Options{
			Downloader: "ffmpeg",
		})

	if ydlResultErr != nil {
		t.Fatalf("failed to download: %s", ydlResultErr)
	}

	dr, err := ydlResult.Download(context.Background(), ydlResult.Info.Formats[0].FormatID)

	if err != nil {
		t.Fatal(err)
	}

	downloadBuf := &bytes.Buffer{}
	_, err = io.Copy(downloadBuf, dr)

	if err != nil {
		t.Fatal(err)
	}
	dr.Close()
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
		t.Skip("skipping test in short mode.")
	}

	// Download file by specifying the playlist index
	stderrBuf := &bytes.Buffer{}
	r, err := NewMetadata(context.Background(), playlistRawURL, Options{
		StderrFn: func(cmd *exec.Cmd) io.Writer {
			return stderrBuf
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedTitle := "Kindred Phenomena"
	if r.Info.Title != expectedTitle {
		t.Errorf("expected title %q got %q", expectedTitle, r.Info.Title)
	}

	expectedEntries := 8
	if len(r.Info.Entries) != expectedEntries {
		t.Errorf("expected %d entries got %d", expectedEntries, len(r.Info.Entries))
	}

	expectedTitleOne := "B1 Mattheis - Ben M"
	playlistIndex := 2
	if r.Info.Entries[playlistIndex].Title != expectedTitleOne {
		t.Errorf("expected title %q got %q", expectedTitleOne, r.Info.Entries[playlistIndex].Title)
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
	dr.Close()

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
	url := "https://soundcloud.com/mattheis/b1-mattheis-ben-m"
	stderrBuf = &bytes.Buffer{}
	r, err = NewMetadata(context.Background(), url, Options{
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

	dr, err = r.Download(context.Background(), r.Info.Formats[0].FormatID)
	if err != nil {
		t.Fatal(err)
	}
	directLinkBuf := &bytes.Buffer{}
	n, err = io.Copy(directLinkBuf, dr)
	if err != nil {
		t.Fatal(err)
	}
	dr.Close()

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
