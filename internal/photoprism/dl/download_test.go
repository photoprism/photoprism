//go:build yt

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

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestDownloadSections(t *testing.T) {
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
		t.Skip("skipped download test in short mode")
	}

	_, ydlResultErr := NewMetadata(context.Background(), testVideoRawURL, Options{
		Type:              TypePlaylist,
		DownloadThumbnail: false,
	})
	if ydlResultErr != ErrNotAPlaylist {
		t.Errorf("expected is playlist error, got %s", ydlResultErr)
	}
}

func TestOptionDownloader(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped download test in short mode")
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

func TestParseInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped download test in short mode")
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

func TestDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped download test in short mode")
	}

	// Fetch metadata.
	stderrBuf := &bytes.Buffer{}
	r, err := NewMetadata(context.Background(), testVideoRawURL, Options{
		StderrFn: func(cmd *exec.Cmd) io.Writer {
			return stderrBuf
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Write metadata to "./testdata/download-test.json".
	jsonName := fs.Abs("./testdata/download-test.json")
	if json := r.Info.JSON(); json != nil {
		if err = fs.WriteFile(jsonName, json, fs.ModeFile); err != nil {
			t.Errorf("%s could not be created: %s", jsonName, err)
		}
	} else {
		t.Errorf("%s could not be created because json data is nil", jsonName)
	}

	// Remove metadata file.
	assert.FileExists(t, jsonName)
	_ = os.Remove(jsonName)

	// Download video based on metadata.
	dr, err := r.Download(context.Background(), r.Info.Formats[0].FormatID)
	if err != nil {
		t.Fatal(err)
	}
	downloadBuf := &bytes.Buffer{}
	n, err := io.Copy(downloadBuf, dr)
	if err != nil {
		t.Fatal(err)
	}
	dr.Close()

	if n != int64(downloadBuf.Len()) {
		t.Errorf("copy n not equal to download buffer: %d!=%d", n, downloadBuf.Len())
	}

	t.Logf("error: %s", stderrBuf.String())

	if n < 10000 {
		t.Errorf("should have copied at least 10000 bytes: %d", n)
	}

	if !strings.Contains(stderrBuf.String(), "Destination") {
		t.Errorf("did not find expected log message on stderr: %q", stderrBuf.String())
	}
}

func TestDownloadWithoutInfo(t *testing.T) {
	stderrBuf := &bytes.Buffer{}
	dr, err := Download(context.Background(), testVideoRawURL, Options{
		StderrFn: func(cmd *exec.Cmd) io.Writer {
			return stderrBuf
		},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	downloadBuf := &bytes.Buffer{}
	n, err := io.Copy(downloadBuf, dr)
	if err != nil {
		t.Fatal(err)
	}
	dr.Close()

	if n != int64(downloadBuf.Len()) {
		t.Errorf("copy n not equal to download buffer: %d!=%d", n, downloadBuf.Len())
	}

	if n < 10000 {
		t.Errorf("should have copied at least 10000 bytes: %d", n)
	}

	if !strings.Contains(stderrBuf.String(), "Destination") {
		t.Errorf("did not find expected log message on stderr: %q", stderrBuf.String())
	}
}
