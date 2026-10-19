package photoprism

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// transportShellQuote quotes a test-owned path as one shell argument.
func transportShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// transportFFmpeg wraps the real converter with separate finite remux and encoding barriers.
func transportFFmpeg(t *testing.T, conf *config.Config) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	logName := filepath.Join(dir, "calls")
	remuxRelease := filepath.Join(dir, "remux-ready")
	encodeRelease := filepath.Join(dir, "encode-ready")
	real := conf.FFmpegBin()
	script := filepath.Join(dir, "ffmpeg")
	body := fmt.Sprintf(`#!/bin/sh
case " $* " in
  *" -i "*) ;;
  *) exec %s "$@" ;;
esac
case " $* " in
  *" -codec copy "*) kind=r; gate=%s ;;
  *) kind=t; gate=%s ;;
esac
printf '%%s\n' "$kind" >> %s
tries=0
while [ ! -f "$gate" ]; do
  tries=$((tries + 1))
  if [ "$tries" -gt 1000 ]; then exit 90; fi
  /bin/sleep 0.01
done
case " $* " in *"h264_nvenc"*) exit 17 ;; esac
exec %s "$@"
`, transportShellQuote(real), transportShellQuote(remuxRelease), transportShellQuote(encodeRelease), transportShellQuote(logName), transportShellQuote(real))
	require.NoError(t, os.WriteFile(script, []byte(body), fs.ModeDir))
	old := conf.Options().FFmpegBin
	conf.Options().FFmpegBin = script
	t.Cleanup(func() { conf.Options().FFmpegBin = old })
	return logName, remuxRelease, encodeRelease
}

// transportStarts counts actual conversion starts recorded by the test wrapper.
func transportStarts(name, kind string) int {
	data, _ := os.ReadFile(name) // #nosec G304 -- caller supplies the test wrapper's own log path.
	count := 0
	for _, line := range strings.Fields(string(data)) {
		if line == kind {
			count++
		}
	}
	return count
}

// TestConvert_CoordinatedTransport runs overlapping requests through the real public conversion path.
func TestConvert_CoordinatedTransport(t *testing.T) {
	for _, format := range []struct{ name, codec string }{{"Avc", "libx264"}, {"Mpeg2", "mpeg2video"}} {
		t.Run(format.name, func(t *testing.T) {
			codec := format.codec
			conf := config.TestConfig()
			if !conf.FFmpegEnabled() || !conf.ExifToolJson() {
				t.Skip("FFmpeg and ExifTool are required")
			}
			source := writeFFmpegFixture(t, conf.FFmpegBin(), t.TempDir(), "shared.m2ts", "mpegts", codec)
			logName, remuxRelease, encodeRelease := transportFFmpeg(t, conf)
			type result struct {
				file *MediaFile
				err  error
			}
			results := make(chan result, 8)
			var workers sync.WaitGroup
			defer func() {
				_ = os.WriteFile(remuxRelease, nil, fs.ModeFile)
				_ = os.WriteFile(encodeRelease, nil, fs.ModeFile)
				workers.Wait()
			}()
			launch := func(count int) {
				for range count {
					input, err := NewMediaFile(source)
					require.NoError(t, err)
					_ = input.MetaData()
					convert := NewConvert(conf)
					workers.Add(1)
					go func() {
						defer workers.Done()
						file, err := convert.ToAvc(input, encode.SoftwareAvc, false, false)
						results <- result{file, err}
					}()
				}
			}
			launch(6)
			require.Eventually(t, func() bool { return transportStarts(logName, "r") > 0 }, 5*time.Second, 10*time.Millisecond)
			time.Sleep(100 * time.Millisecond)
			assert.Equal(t, 1, transportStarts(logName, "r"))
			require.NoError(t, os.WriteFile(remuxRelease, nil, fs.ModeFile))
			if codec == "mpeg2video" {
				require.Eventually(t, func() bool { return transportStarts(logName, "t") > 0 }, 5*time.Second, 10*time.Millisecond)
				// Late arrivals must join while fallback encoding has not produced a reusable file.
				launch(2)
				time.Sleep(100 * time.Millisecond)
				assert.Equal(t, 1, transportStarts(logName, "r"))
			} else {
				launch(2)
			}
			require.NoError(t, os.WriteFile(encodeRelease, nil, fs.ModeFile))
			var name string
			var previous *MediaFile
			for range 8 {
				select {
				case result := <-results:
					require.NoError(t, result.err)
					require.NotNil(t, result.file)
					if name == "" {
						name = result.file.FileName()
					} else {
						assert.Equal(t, name, result.file.FileName())
						assert.NotSame(t, previous, result.file)
					}
					previous = result.file
				case <-time.After(10 * time.Second):
					t.Fatal("conversion did not complete")
				}
			}
			assert.Equal(t, 1, transportStarts(logName, "r"))
			expectedEncode := 0
			if codec == "mpeg2video" {
				expectedEncode = 1
			}
			assert.Equal(t, expectedEncode, transportStarts(logName, "t"))
			if codec == "mpeg2video" {
				old := time.Unix(1000000000, 0)
				require.NoError(t, os.Chtimes(name, old, old))
				input, err := NewMediaFile(source)
				require.NoError(t, err)
				forced, err := NewConvert(conf).ToAvc(input, encode.SoftwareAvc, false, true)
				require.NoError(t, err)
				require.NotNil(t, forced)
				info, err := os.Stat(name)
				require.NoError(t, err)
				assert.False(t, info.ModTime().Equal(old))
				assert.Equal(t, 1, transportStarts(logName, "r"))
				assert.Equal(t, 2, transportStarts(logName, "t"))
			}
		})
	}
}

// TestConvert_CoordinatedTransportFailure checks a retry through the same public entry point.
func TestConvert_CoordinatedTransportFailure(t *testing.T) {
	conf := config.TestConfig()
	if !conf.FFmpegEnabled() || !conf.ExifToolJson() {
		t.Skip("FFmpeg and ExifTool are required")
	}
	dir := t.TempDir()
	source := writeFFmpegFixture(t, conf.FFmpegBin(), dir, "retry.m2ts", "mpegts", "libx264")
	valid, err := os.ReadFile(source) // #nosec G304 -- fixture-generated input.
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(source, []byte("unreadable transport stream"), fs.ModeFile))
	logName, remuxRelease, encodeRelease := transportFFmpeg(t, conf)
	var workers sync.WaitGroup
	results := make(chan error, 4)
	defer func() {
		_ = os.WriteFile(remuxRelease, nil, fs.ModeFile)
		_ = os.WriteFile(encodeRelease, nil, fs.ModeFile)
		workers.Wait()
	}()
	for range 4 {
		input, err := NewMediaFile(source)
		require.NoError(t, err)
		_ = input.MetaData()
		convert := NewConvert(conf)
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, err := convert.ToAvc(input, encode.SoftwareAvc, false, false)
			results <- err
		}()
	}
	require.Eventually(t, func() bool { return transportStarts(logName, "r") > 0 }, 5*time.Second, 10*time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	require.NoError(t, os.WriteFile(remuxRelease, nil, fs.ModeFile))
	require.NoError(t, os.WriteFile(encodeRelease, nil, fs.ModeFile))
	var failure error
	for range 4 {
		select {
		case err := <-results:
			require.Error(t, err)
			if failure == nil {
				failure = err
			} else {
				assert.Equal(t, failure, err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("failed conversion did not complete")
		}
	}
	assert.Equal(t, 1, transportStarts(logName, "r"))
	require.NoError(t, os.WriteFile(source, valid, fs.ModeFile)) // #nosec G703 -- restores this test's own generated fixture.
	input, err := NewMediaFile(source)
	require.NoError(t, err)
	output, err := NewConvert(conf).ToAvc(input, encode.SoftwareAvc, false, false)
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 2, transportStarts(logName, "r"))
}

// TestConvert_CoordinatedTransportIndependent starts distinct destinations before either completes.
func TestConvert_CoordinatedTransportIndependent(t *testing.T) {
	conf := config.TestConfig()
	if !conf.FFmpegEnabled() || !conf.ExifToolJson() {
		t.Skip("FFmpeg and ExifTool are required")
	}
	dir := t.TempDir()
	sources := []string{
		writeFFmpegFixture(t, conf.FFmpegBin(), dir, "one.m2ts", "mpegts", "libx264"),
		writeFFmpegFixture(t, conf.FFmpegBin(), dir, "two.m2ts", "mpegts", "libx264"),
	}
	logName, remuxRelease, encodeRelease := transportFFmpeg(t, conf)
	results := make(chan error, 2)
	var workers sync.WaitGroup
	defer func() {
		_ = os.WriteFile(remuxRelease, nil, fs.ModeFile)
		_ = os.WriteFile(encodeRelease, nil, fs.ModeFile)
		workers.Wait()
	}()
	for _, source := range sources {
		input, err := NewMediaFile(source)
		require.NoError(t, err)
		_ = input.MetaData()
		convert := NewConvert(conf)
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, err := convert.ToAvc(input, encode.SoftwareAvc, false, false)
			results <- err
		}()
	}
	require.Eventually(t, func() bool { return transportStarts(logName, "r") == 2 }, 5*time.Second, 10*time.Millisecond)
	require.NoError(t, os.WriteFile(remuxRelease, nil, fs.ModeFile))
	require.NoError(t, os.WriteFile(encodeRelease, nil, fs.ModeFile))
	for range 2 {
		select {
		case err := <-results:
			require.NoError(t, err)
		case <-time.After(10 * time.Second):
			t.Fatal("independent conversion did not complete")
		}
	}
}

// TestConvert_CoordinatedTransportFallback keeps software fallback in the current destination job.
func TestConvert_CoordinatedTransportFallback(t *testing.T) {
	conf := config.TestConfig()
	if !conf.FFmpegEnabled() || !conf.ExifToolJson() {
		t.Skip("FFmpeg and ExifTool are required")
	}
	source := writeFFmpegFixture(t, conf.FFmpegBin(), t.TempDir(), "fallback.m2ts", "mpegts", "mpeg2video")
	logName, remuxRelease, encodeRelease := transportFFmpeg(t, conf)
	require.NoError(t, os.WriteFile(remuxRelease, nil, fs.ModeFile))
	require.NoError(t, os.WriteFile(encodeRelease, nil, fs.ModeFile))
	input, err := NewMediaFile(source)
	require.NoError(t, err)
	result := make(chan error, 1)
	go func() { _, err := NewConvert(conf).ToAvc(input, encode.NvidiaAvc, false, false); result <- err }()
	select {
	case err := <-result:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("fallback did not complete")
	}
	assert.Equal(t, 2, transportStarts(logName, "t"), "one rejected hardware attempt and one software attempt")
}

// TestConvert_CoordinatedTransportInvalidPath refuses paths before registering an operation.
func TestConvert_CoordinatedTransportInvalidPath(t *testing.T) {
	convert := NewConvert(config.TestConfig())
	_, err := convert.coordinatedTransport(&MediaFile{}, encode.SoftwareAvc, false, false)
	require.Error(t, err)
	_, err = convert.coordinatedTransport(&MediaFile{fileName: filepath.Join(t.TempDir(), "missing.m2ts")}, encode.SoftwareAvc, false, false)
	require.Error(t, err)
}

// TestConvert_CoordinatedTransportExcludes queues distinct exclusion snapshots on one config.
func TestConvert_CoordinatedTransportExcludes(t *testing.T) {
	conf := config.TestConfig()
	if !conf.FFmpegEnabled() || !conf.ExifToolJson() {
		t.Skip("FFmpeg and ExifTool are required")
	}
	source := filepath.Join(t.TempDir(), "retry.m2ts")
	require.NoError(t, os.WriteFile(source, []byte("unreadable transport stream"), fs.ModeFile))
	logName, remuxRelease, encodeRelease := transportFFmpeg(t, conf)
	saved := ffmpeg.Exclude()
	t.Cleanup(func() { ffmpeg.SetExclude(saved) })
	ffmpeg.SetExclude(video.NewFormats("avi"))
	first := NewConvert(conf)
	ffmpeg.SetExclude(video.NewFormats("mov"))
	second := NewConvert(conf)
	results := make(chan error, 2)
	var workers sync.WaitGroup
	defer func() {
		_ = os.WriteFile(remuxRelease, nil, fs.ModeFile)
		_ = os.WriteFile(encodeRelease, nil, fs.ModeFile)
		workers.Wait()
	}()
	launch := func(convert *Convert) {
		input, err := NewMediaFile(source)
		require.NoError(t, err)
		_ = input.MetaData()
		require.True(t, convert.FFmpegAllowed(input))
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, err := convert.ToAvc(input, encode.SoftwareAvc, false, false)
			results <- err
		}()
	}
	launch(first)
	require.Eventually(t, func() bool { return transportStarts(logName, "r") == 1 }, 5*time.Second, 10*time.Millisecond)
	var key string
	var firstCall *transportCall
	transportConversions.mutex.Lock()
	for destination, call := range transportConversions.calls {
		if call.request.source == source {
			key, firstCall = destination, call
			break
		}
	}
	transportConversions.mutex.Unlock()
	require.NotNil(t, firstCall)
	launch(second)
	require.Eventually(t, func() bool {
		transportConversions.mutex.Lock()
		defer transportConversions.mutex.Unlock()
		call := transportConversions.calls[key]
		return call != nil && call != firstCall
	}, 3*time.Second, 10*time.Millisecond, "distinct exclusion snapshots must queue separate operations")
	assert.Equal(t, 1, transportStarts(logName, "r"))
	require.NoError(t, os.WriteFile(remuxRelease, nil, fs.ModeFile))
	require.NoError(t, os.WriteFile(encodeRelease, nil, fs.ModeFile))
	for range 2 {
		select {
		case err := <-results:
			require.Error(t, err)
		case <-time.After(10 * time.Second):
			t.Fatal("conversion did not complete")
		}
	}
	assert.Equal(t, 2, transportStarts(logName, "r"))
}
