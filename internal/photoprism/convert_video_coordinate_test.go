package photoprism

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/projection"
)

// blockingAvcCommand records converter calls and holds them before executing real FFmpeg.
func blockingAvcCommand(t *testing.T, w *Convert, fail bool) (string, string, func()) {
	t.Helper()
	dir := t.TempDir()
	started, release, calls := filepath.Join(dir, "started"), filepath.Join(dir, "release"), filepath.Join(dir, "calls")
	real := w.conf.FFmpegBin()
	command := filepath.Join(dir, "ffmpeg")
	script := "#!/bin/sh\nprintf 'run\\n' >> " + transportShellQuote(calls) + "\ntouch " + transportShellQuote(started) + "\nwhile [ ! -f " + transportShellQuote(release) + " ]; do sleep 0.01; done\n"
	if fail {
		script += "for last; do :; done\nprintf partial > \"$last\"\nexit 1\n"
	} else {
		script += "exec " + transportShellQuote(real) + " \"$@\"\n"
	}
	require.NoError(t, os.WriteFile(command, []byte(script), fs.ModeSecretFile|0100))
	old := w.conf.Options().FFmpegBin
	w.conf.Options().FFmpegBin = command
	unblock := func() { _ = os.WriteFile(release, nil, fs.ModeFile) }
	t.Cleanup(func() { unblock(); w.conf.Options().FFmpegBin = old })
	return started, calls, unblock
}

// TestAvcDestinationKey checks stable destination keys before directory creation and through links.
func TestAvcDestinationKey(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "new", "nested", "clip.mov.avc")
	before := avcDestinationKey(name)
	require.NoError(t, fs.MkdirAll(filepath.Dir(name)))
	assert.Equal(t, before, avcDestinationKey(name))
	alias := filepath.Join(t.TempDir(), "alias")
	require.NoError(t, os.Symlink(dir, alias))
	assert.Equal(t, before, avcDestinationKey(filepath.Join(alias, "new", "nested", "clip.mov.avc")))
	assert.Equal(t, filepath.Join(dir, "future", "clip.avc"), avcDestinationKey(filepath.Join(alias, "future", "clip.avc")))
}

// TestConvert_CoordinatedAvc ensures Export and playback reuse one atomically published sidecar.
func TestConvert_CoordinatedAvc(t *testing.T) {
	for _, firstExport := range []bool{false, true} {
		name := "PlaybackFirst"
		if firstExport {
			name = "ExportFirst"
		}
		t.Run(name, func(t *testing.T) {
			w := exportTestConfig(t)
			f := exportTestVideo(t, w, "shared.mp4", "libx265")
			media, err := NewMediaFile(filepath.Join(w.conf.OriginalsPath(), f.FileName))
			require.NoError(t, err)
			destination, err := w.avcName(media)
			require.NoError(t, err)
			var wg sync.WaitGroup
			t.Cleanup(wg.Wait)
			started, calls, unblock := blockingAvcCommand(t, w, false)
			results := make(chan error, 2)
			prepare := func(export bool) {
				defer wg.Done()
				if export {
					_, callErr := w.PrepareExport(f, "avc")
					results <- callErr
				} else {
					_, callErr := w.ToAvc(media, encode.SoftwareAvc, false, false)
					results <- callErr
				}
			}
			wg.Add(1)
			go prepare(firstExport)
			require.Eventually(t, func() bool { return fs.FileExists(started) }, 5*time.Second, time.Millisecond)
			assert.NoFileExists(t, destination)
			probe, err := w.ProbeExport(f)
			require.NoError(t, err)
			require.NotNil(t, probe.Avc)
			assert.False(t, probe.Avc.Ready)
			_, _, _, err = w.ExportFile(f, "avc")
			assert.ErrorIs(t, err, ErrExportNotReady)
			avcConversions.mutex.Lock()
			firstCall := avcConversions.calls[avcDestinationKey(destination)]
			avcConversions.mutex.Unlock()
			wg.Add(1)
			go prepare(!firstExport)
			require.Eventually(t, func() bool {
				avcConversions.mutex.Lock()
				defer avcConversions.mutex.Unlock()
				call := avcConversions.calls[avcDestinationKey(destination)]
				return call != nil && call != firstCall
			}, 5*time.Second, time.Millisecond)
			unblock()
			for range 2 {
				require.NoError(t, <-results)
			}
			data, err := os.ReadFile(calls) //nolint:gosec // calls is a disposable test marker
			require.NoError(t, err)
			assert.Equal(t, 1, strings.Count(string(data), "run\n"))
			actual, _ := exportTestSidecar(t, w, f)
			assert.Equal(t, destination, actual)
			entries, err := os.ReadDir(w.conf.SidecarPath())
			require.NoError(t, err)
			require.Len(t, entries, 1)
			assert.NoDirExists(t, filepath.Join(w.conf.CachePath(), "export"))
		})
	}
}

// TestConvert_ToAvcAtomicFailure preserves an existing output when forced replacement fails.
func TestConvert_ToAvcAtomicFailure(t *testing.T) {
	w := exportTestConfig(t)
	f := exportTestVideo(t, w, "replace.mp4", "libx265")
	_, err := w.PrepareExport(f, "avc")
	require.NoError(t, err)
	name, _ := exportTestSidecar(t, w, f)
	before, err := os.ReadFile(name) //nolint:gosec // name is in the disposable sidecar directory
	require.NoError(t, err)
	media, err := NewMediaFile(filepath.Join(w.conf.OriginalsPath(), f.FileName))
	require.NoError(t, err)
	started, _, unblock := blockingAvcCommand(t, w, true)
	var wg sync.WaitGroup
	t.Cleanup(func() { unblock(); wg.Wait() })
	result := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, runErr := w.ToAvc(media, encode.SoftwareAvc, false, true)
		result <- runErr
	}()
	require.Eventually(t, func() bool { return fs.FileExists(started) }, 5*time.Second, time.Millisecond)
	during, err := os.ReadFile(name) //nolint:gosec // name is in the disposable sidecar directory
	require.NoError(t, err)
	assert.Equal(t, before, during)
	unblock()
	assert.Error(t, <-result)
	after, err := os.ReadFile(name) //nolint:gosec // name is in the disposable sidecar directory
	require.NoError(t, err)
	assert.Equal(t, before, after)
	entries, err := os.ReadDir(w.conf.SidecarPath())
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

// TestConvert_SourceChanges rejects source replacement during encoding, fallback, and remuxing.
func TestConvert_SourceChanges(t *testing.T) {
	for _, mode := range []string{"Encoding", "HardwareFallback", "Remux"} {
		t.Run(mode, func(t *testing.T) {
			w := exportTestConfig(t)
			name := filepath.Join(w.conf.OriginalsPath(), "changing.mp4")
			if mode == "Remux" {
				name = writeFFmpegFixture(t, w.conf.FFmpegBin(), w.conf.OriginalsPath(), "changing.m2ts", "mpegts", "libx264")
			} else {
				exportTestVideo(t, w, "changing.mp4", "libx265")
			}
			media, err := NewMediaFile(name)
			require.NoError(t, err)
			before, err := os.ReadFile(name) //nolint:gosec // disposable source fixture
			require.NoError(t, err)
			started, calls, unblock := blockingAvcCommand(t, w, mode == "HardwareFallback")
			var wg sync.WaitGroup
			t.Cleanup(func() { unblock(); wg.Wait() })
			encoder := encode.SoftwareAvc
			if mode == "HardwareFallback" {
				encoder = encode.NvidiaAvc
			}
			result := make(chan error, 1)
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, runErr := w.ToAvc(media, encoder, false, false)
				result <- runErr
			}()
			require.Eventually(t, func() bool { return fs.FileExists(started) }, 5*time.Second, time.Millisecond)
			replacement := name + ".replacement"
			require.NoError(t, os.WriteFile(replacement, before, fs.ModeFile)) //nolint:gosec // fixture path is confined to the disposable originals directory
			require.NoError(t, os.Rename(replacement, name))
			unblock()
			assert.ErrorIs(t, <-result, ErrExportSourceChanged)
			data, err := os.ReadFile(calls) //nolint:gosec // disposable command marker
			require.NoError(t, err)
			assert.Equal(t, 1, strings.Count(string(data), "run\n"))
			entries, err := os.ReadDir(w.conf.SidecarPath())
			require.NoError(t, err)
			assert.Empty(t, entries)
		})
	}
}

// TestExportQueuedSymlink rejects retargeting while an exact request waits for its destination.
func TestExportQueuedSymlink(t *testing.T) {
	w := exportTestConfig(t)
	original := exportTestVideo(t, w, "old.mp4", "libx265")
	exportTestVideo(t, w, "new.mp4", "libx264")
	alias := filepath.Join(w.conf.OriginalsPath(), "alias.mp4")
	require.NoError(t, os.Symlink(filepath.Join(w.conf.OriginalsPath(), original.FileName), alias))
	selected := *original
	selected.FileName = "alias.mp4"
	blocked, key, unblock := blockAvcDestination(t, w, &selected)
	var wg sync.WaitGroup
	t.Cleanup(func() { unblock(); wg.Wait() })
	result := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := w.PrepareExport(&selected, "avc")
		result <- err
	}()
	require.Eventually(t, func() bool {
		avcConversions.mutex.Lock()
		defer avcConversions.mutex.Unlock()
		return avcConversions.calls[key] != blocked
	}, 5*time.Second, time.Millisecond)
	replacement := alias + ".replacement"
	require.NoError(t, os.Symlink(filepath.Join(w.conf.OriginalsPath(), "new.mp4"), replacement))
	require.NoError(t, os.Rename(replacement, alias))
	unblock()
	assert.ErrorIs(t, <-result, ErrExportSourceChanged)
	entries, err := os.ReadDir(w.conf.SidecarPath())
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// TestConvert_CoordinatedAvcProjection preserves dewarp metadata for sidecars inside Originals.
func TestConvert_CoordinatedAvcProjection(t *testing.T) {
	w := exportTestConfig(t)
	w.conf.Options().SidecarPath = filepath.Join(w.conf.OriginalsPath(), ".photoprism")
	w.conf.Options().DisableExifTool = true
	require.NoError(t, fs.MkdirAll(w.conf.SidecarPath()))
	writeInsta360StackMedia(t, w.conf, w.conf.OriginalsPath(), insta360StackLeft)
	writeInsta360StackMedia(t, w.conf, w.conf.OriginalsPath(), insta360StackRight)
	media, err := NewMediaFile(filepath.Join(w.conf.OriginalsPath(), insta360StackLeft))
	require.NoError(t, err)
	output, err := w.ToAvc(media, encode.SoftwareAvc, false, false)
	require.NoError(t, err)
	require.NotNil(t, output)
	assert.False(t, output.InSidecar())
	assert.Equal(t, projection.Equirectangular, output.visualProjection)
	assert.Equal(t, projection.Equirectangular, output.VisualProjection(""))
}
