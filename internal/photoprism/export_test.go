package photoprism

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

// exportTestConfig installs a disposable media library with software encoding.
func exportTestConfig(t *testing.T) *Convert {
	t.Helper()
	old := Config()
	conf := config.NewMinimalTestConfig(t.TempDir())
	conf.Options().FFmpegEncoder = encode.SoftwareAvc.String()
	conf.Options().FFmpegSize = 720
	require.NoError(t, conf.CreateDirectories())
	SetConfig(conf)
	t.Cleanup(func() { SetConfig(old) })
	return NewConvert(conf)
}

// exportTestVideo creates a small real clip without hardware dependencies.
func exportTestVideo(t *testing.T, w *Convert, name, codec string) *entity.File {
	t.Helper()
	if w.conf.FFmpegBin() == "" {
		t.Skip("ffmpeg unavailable")
	}
	dest := filepath.Join(w.conf.OriginalsPath(), name)
	require.NoError(t, fs.MkdirAll(filepath.Dir(dest)))
	args := []string{"-v", "error", "-f", "lavfi", "-i", "testsrc2=size=64x48:rate=5", "-t", "0.4", "-threads", "1", "-c:v", codec}
	if codec == "libx265" {
		args = append(args, "-x265-params", "pools=1:frame-threads=1:log-level=error")
	}
	args = append(args, "-y", dest)
	data, err := exec.Command(w.conf.FFmpegBin(), args...).CombinedOutput() //nolint:gosec // executable and paths are controlled disposable fixtures
	require.NoError(t, err, "%s", data)
	return &entity.File{FileRoot: entity.RootOriginals, FileName: name, FileHash: fs.Hash(dest), FileType: filepath.Ext(name)[1:], FileVideo: true, MediaType: entity.MediaVideo, FileWidth: 64, FileHeight: 48, FileDuration: 400 * time.Millisecond}
}

// TestExportProbe checks codec-based choices and read-only probing.
func TestExportProbe(t *testing.T) {
	w := exportTestConfig(t)
	t.Run("AvcHasNoDuplicate", func(t *testing.T) {
		f := exportTestVideo(t, w, "already.mov", "libx264")
		probe, err := w.ProbeExport(f)
		require.NoError(t, err)
		assert.True(t, exportAvc(probe.Codec))
		assert.Nil(t, probe.Avc)
		assert.Greater(t, probe.Duration, 0.0)
		entries, err := os.ReadDir(w.conf.CachePath())
		require.NoError(t, err)
		for _, entry := range entries {
			assert.NotEqual(t, "export", entry.Name())
		}
	})
	t.Run("HevcRequiresPreparation", func(t *testing.T) {
		f := exportTestVideo(t, w, "hevc.mp4", "libx265")
		probe, err := w.ProbeExport(f)
		require.NoError(t, err)
		require.NotNil(t, probe.Avc)
		assert.True(t, probe.Avc.Preparable)
		assert.False(t, probe.Avc.Ready)
	})
	t.Run("DisabledConverter", func(t *testing.T) {
		f := exportTestVideo(t, w, "disabled.mp4", "libx265")
		w.conf.Options().DisableFFmpeg = true
		defer func() { w.conf.Options().DisableFFmpeg = false }()
		probe, err := w.ProbeExport(f)
		require.NoError(t, err)
		assert.Nil(t, probe.Avc)
	})
	t.Run("Unavailable", func(t *testing.T) {
		_, err := w.ProbeExport(nil)
		assert.ErrorIs(t, err, ErrExportUnavailable)
		_, err = w.ProbeExport(&entity.File{FileHash: "../"})
		assert.ErrorIs(t, err, ErrExportUnavailable)
	})
	t.Run("CodecNames", func(t *testing.T) {
		for _, codec := range []string{"avc1", "AVC3", "h264", "H.264", "x264"} {
			assert.True(t, exportAvc(codec))
		}
		for _, codec := range []string{"", "hevc", "hvc1", "vp9"} {
			assert.False(t, exportAvc(codec))
		}
	})
	t.Run("NonMp4Probe", func(t *testing.T) {
		f := exportTestVideo(t, w, "avc.mkv", "libx264")
		probe, err := w.ProbeExport(f)
		require.NoError(t, err)
		assert.True(t, exportAvc(probe.Codec))
		assert.Nil(t, probe.Avc)
	})
	t.Run("NonAvcMkvOffersConversion", func(t *testing.T) {
		f := exportTestVideo(t, w, "hevc.mkv", "libx265")
		probe, err := w.ProbeExport(f)
		require.NoError(t, err)
		require.NotNil(t, probe.Avc)
		assert.True(t, probe.Avc.Preparable)
	})
}

// TestExportAppleKey rejects generic metadata IDs while accepting explicit Apple identifiers.
func TestExportAppleKey(t *testing.T) {
	w := exportTestConfig(t)
	f := &entity.File{FileHash: "a123456789012345678901234567890123456789", MediaType: entity.MediaImage}
	assert.Empty(t, w.exportAppleKey(f))
	dir, err := fs.CachePath(w.conf.CachePath(), f.FileHash, "json", true)
	require.NoError(t, err)
	name := filepath.Join(dir, f.FileHash+"_exiftool.json")
	for _, tc := range []struct{ name, data, key string }{
		{"Identifier", `[{"ContentIdentifier":"pair-id"}]`, "pair-id"},
		{"GroupedTag", `[{"MakerNotes:ContentIdentifier":"pair-id"}]`, "pair-id"},
		{"GroupUUID", `{"MediaGroupUUID":"pair-id"}`, "pair-id"},
		{"GenericIDs", `[{"DocumentID":"not-a-pair","InstanceID":"not-a-pair"}]`, ""},
		{"Malformed", `{`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, os.WriteFile(name, []byte(tc.data), fs.ModeFile))
			assert.Equal(t, tc.key, w.exportAppleKey(f))
		})
	}
}

// exportTestSidecar resolves an output and closes its download handle.
func exportTestSidecar(t *testing.T, w *Convert, f *entity.File) (string, ExportOutput) {
	t.Helper()
	file, offset, output, err := w.ExportFile(f, "avc")
	require.NoError(t, err)
	defer file.Close()
	require.Zero(t, offset)
	return file.Name(), output
}

// blockAvcDestination holds the shared sidecar coordinator until the test releases it.
func blockAvcDestination(t *testing.T, w *Convert, f *entity.File) (*transportCall, string, func()) {
	t.Helper()
	media, err := NewMediaFile(filepath.Join(w.conf.OriginalsPath(), f.FileName))
	require.NoError(t, err)
	name, err := w.avcName(media)
	require.NoError(t, err)
	key := avcDestinationKey(name)
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	call := avcConversions.start(key, transportRequest{}, func() (string, error) { <-release; return "", nil })
	return call, key, unblock
}

// TestExportPrepare checks standard sidecar reuse, source identity, and non-destructive failures.
func TestExportPrepare(t *testing.T) {
	w := exportTestConfig(t)
	f := exportTestVideo(t, w, "same.mp4", "libx265")
	t.Run("DownloadDoesNotPrepare", func(t *testing.T) {
		_, _, _, err := w.ExportFile(f, "avc")
		assert.ErrorIs(t, err, ErrExportNotReady)
		entries, err := os.ReadDir(w.conf.SidecarPath())
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
	t.Run("ConcurrentPreparation", func(t *testing.T) {
		for i := 0; i < cap(exportSlots)-1; i++ {
			exportSlots <- struct{}{}
		}
		defer func() {
			for i := 0; i < cap(exportSlots)-1; i++ {
				<-exportSlots
			}
		}()
		_, _, unblock := blockAvcDestination(t, w, f)
		var wg sync.WaitGroup
		errors := make([]error, 4)
		outputs := make([]ExportOutput, 4)
		for i := range errors {
			wg.Add(1)
			go func() { defer wg.Done(); outputs[i], errors[i] = w.PrepareExport(f, "avc") }()
		}
		unblock()
		wg.Wait()
		for i := range errors {
			require.NoError(t, errors[i])
			assert.True(t, outputs[i].Ready)
			assert.Equal(t, "avc1", outputs[i].Codec)
		}
		name, output := exportTestSidecar(t, w, f)
		assert.Equal(t, filepath.Join(w.conf.SidecarPath(), f.FileName+fs.ExtAvc), name)
		assert.Equal(t, "mp4", output.Format)
		assert.NoDirExists(t, filepath.Join(w.conf.CachePath(), "export"))
	})
	t.Run("PlaybackReusesExport", func(t *testing.T) {
		name, _ := exportTestSidecar(t, w, f)
		before, err := os.Stat(name)
		require.NoError(t, err)
		media, err := NewMediaFile(filepath.Join(w.conf.OriginalsPath(), f.FileName))
		require.NoError(t, err)
		file, err := w.ToAvc(media, w.conf.FFmpegEncoder(), false, false)
		require.NoError(t, err)
		assert.Equal(t, name, file.FileName())
		after, err := os.Stat(name)
		require.NoError(t, err)
		assert.True(t, os.SameFile(before, after))
		assert.Equal(t, before.ModTime(), after.ModTime())
	})
	t.Run("ReadySidecarSurvivesDisabledConverterAndReadOnly", func(t *testing.T) {
		w.conf.Options().DisableFFmpeg, w.conf.Options().ReadOnly = true, true
		defer func() { w.conf.Options().DisableFFmpeg, w.conf.Options().ReadOnly = false, false }()
		probe, err := w.ProbeExport(f)
		require.NoError(t, err)
		require.NotNil(t, probe.Avc)
		assert.True(t, probe.Avc.Ready)
		exportTestSidecar(t, w, f)
		_, err = w.PrepareExport(f, "avc")
		require.NoError(t, err)
	})
	t.Run("SettingsReuseTheSameSidecar", func(t *testing.T) {
		name, expected := exportTestSidecar(t, w, f)
		before, err := os.Stat(name)
		require.NoError(t, err)
		w.conf.Options().FFmpegQuality, w.conf.Options().FFmpegSize = 80, 32
		defer func() { w.conf.Options().FFmpegQuality, w.conf.Options().FFmpegSize = 0, 720 }()
		output, err := w.PrepareExport(f, "avc")
		require.NoError(t, err)
		assert.Equal(t, expected, output)
		after, err := os.Stat(name)
		require.NoError(t, err)
		assert.True(t, os.SameFile(before, after))
		assert.Equal(t, before.ModTime(), after.ModTime())
	})
	t.Run("SameBasenameIsSeparate", func(t *testing.T) {
		other := exportTestVideo(t, w, "same.mov", "libx265")
		_, err := w.PrepareExport(other, "avc")
		require.NoError(t, err)
		first, _ := exportTestSidecar(t, w, f)
		second, _ := exportTestSidecar(t, w, other)
		assert.NotEqual(t, first, second)
	})
	t.Run("CorruptSidecarIsNotOverwritten", func(t *testing.T) {
		other := exportTestVideo(t, w, "corrupt.mp4", "libx265")
		_, err := w.PrepareExport(other, "avc")
		require.NoError(t, err)
		name, _ := exportTestSidecar(t, w, other)
		require.NoError(t, os.WriteFile(name, []byte("partial"), fs.ModeFile))
		_, _, _, err = w.ExportFile(other, "avc")
		assert.ErrorIs(t, err, ErrExportNotReady)
		_, err = w.PrepareExport(other, "avc")
		assert.ErrorIs(t, err, ErrExportUnavailable)
		data, err := os.ReadFile(name) //nolint:gosec // name is inside the disposable sidecar path
		require.NoError(t, err)
		assert.Equal(t, "partial", string(data))
	})
	t.Run("InvalidKind", func(t *testing.T) {
		_, err := w.PrepareExport(f, "../mp4")
		assert.ErrorIs(t, err, ErrExportUnavailable)
	})
	t.Run("ChangedSource", func(t *testing.T) {
		copy := *f
		copy.FileHash = "1234567890123456789012345678901234567890"
		_, err := w.PrepareExport(&copy, "avc")
		assert.ErrorIs(t, err, ErrExportSourceChanged)
	})
}

// TestExportMotion verifies direct embedded-video delivery without storing a second file.
func TestExportMotion(t *testing.T) {
	w := exportTestConfig(t)
	native := exportTestVideo(t, w, "native.mp4", "libx265")
	videoBytes, err := os.ReadFile(filepath.Join(w.conf.OriginalsPath(), native.FileName))
	require.NoError(t, err)
	imageBytes, err := os.ReadFile("testdata/flash.jpg")
	require.NoError(t, err)
	name := filepath.Join(w.conf.OriginalsPath(), "motion.jpg")
	require.NoError(t, os.WriteFile(name, append(imageBytes, videoBytes...), fs.ModeFile)) //nolint:gosec // fixture path is confined to the disposable originals directory
	f := &entity.File{FileRoot: entity.RootOriginals, FileName: "motion.jpg", FileHash: fs.Hash(name), FileType: "jpg", MediaType: entity.MediaLive}
	paths := []string{w.conf.CachePath(), w.conf.MediaCachePath(), w.conf.SidecarPath()}
	before := make([][]os.DirEntry, len(paths))
	for i, path := range paths {
		before[i], err = os.ReadDir(path)
		require.NoError(t, err)
	}
	probe, err := w.ProbeExport(f)
	require.NoError(t, err)
	require.NotNil(t, probe.VideoPart)
	assert.Nil(t, probe.Avc)
	stale := *f
	stale.FileHash = "1234567890123456789012345678901234567890"
	_, _, _, err = w.ExportFile(&stale, "video-part")
	assert.ErrorIs(t, err, ErrExportSourceChanged)
	_, err = w.PrepareExport(&stale, "video-part")
	assert.ErrorIs(t, err, ErrExportSourceChanged)
	assert.True(t, probe.VideoPart.Ready)
	assert.Equal(t, int64(len(videoBytes)), probe.VideoPart.Bytes)
	output, err := w.PrepareExport(f, "video-part")
	require.NoError(t, err)
	assert.Equal(t, *probe.VideoPart, output)
	file, offset, downloaded, err := w.ExportFile(f, "video-part")
	require.NoError(t, err)
	defer file.Close()
	assert.Equal(t, name, file.Name())
	assert.Equal(t, int64(len(imageBytes)), offset)
	data, err := io.ReadAll(io.NewSectionReader(file, offset, downloaded.Bytes))
	require.NoError(t, err)
	assert.Equal(t, videoBytes, data)
	for i, path := range paths {
		entries, readErr := os.ReadDir(path)
		require.NoError(t, readErr)
		assert.Equal(t, before[i], entries)
	}
	_, err = w.PrepareExport(f, "avc")
	assert.ErrorIs(t, err, ErrExportUnavailable)
}

// TestExportAnimations verifies standard animation sidecars and reuse without source changes.
func TestExportAnimations(t *testing.T) {
	for _, fixture := range []struct{ name, file string }{
		{"GIF", "2018-04-12 19_24_49.gif"},
		{"APNG", "animated.png"},
		{"WebP", "windows95.webp"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			w := exportTestConfig(t)
			data, err := os.ReadFile(filepath.Join("testdata", fixture.file)) //nolint:gosec // fixed fixture names
			require.NoError(t, err)
			name := filepath.Join(w.conf.OriginalsPath(), fixture.file)
			require.NoError(t, os.WriteFile(name, data, fs.ModeFile)) //nolint:gosec // fixture path is confined to the disposable originals directory
			f := &entity.File{FileRoot: entity.RootOriginals, FileName: fixture.file, FileHash: fs.Hash(name), FileType: filepath.Ext(name)[1:], MediaType: entity.MediaAnimated, FileFrames: 2, FileDuration: time.Second}
			output, err := w.PrepareExport(f, "avc")
			require.NoError(t, err)
			assert.True(t, output.Ready)
			assert.True(t, exportAvc(output.Codec))
			sidecar, _ := exportTestSidecar(t, w, f)
			assert.Equal(t, filepath.Join(w.conf.SidecarPath(), fixture.file+".mp4"), sidecar)
			before, err := os.Stat(sidecar)
			require.NoError(t, err)
			_, err = w.PrepareExport(f, "avc")
			require.NoError(t, err)
			after, err := os.Stat(sidecar)
			require.NoError(t, err)
			assert.True(t, os.SameFile(before, after))
			assert.Equal(t, before.ModTime(), after.ModTime())
			assert.Equal(t, f.FileHash, fs.Hash(name))
			assert.NoDirExists(t, filepath.Join(w.conf.CachePath(), "export"))
		})
	}
}

// TestExportOpenDescriptor serves the verified inode even after its pathname is replaced.
func TestExportOpenDescriptor(t *testing.T) {
	w := exportTestConfig(t)
	f := exportTestVideo(t, w, "descriptor.mp4", "libx265")
	_, err := w.PrepareExport(f, "avc")
	require.NoError(t, err)
	file, _, output, err := w.ExportFile(f, "avc")
	require.NoError(t, err)
	defer file.Close()
	before, err := io.ReadAll(io.NewSectionReader(file, 0, output.Bytes))
	require.NoError(t, err)
	replacement := file.Name() + ".replacement"
	require.NoError(t, os.WriteFile(replacement, make([]byte, len(before)), fs.ModeFile))
	require.NoError(t, os.Rename(replacement, file.Name()))
	after, err := io.ReadAll(io.NewSectionReader(file, 0, output.Bytes))
	require.NoError(t, err)
	assert.Equal(t, before, after)
	_, _, _, err = w.ExportFile(f, "avc")
	assert.ErrorIs(t, err, ErrExportNotReady)
}
