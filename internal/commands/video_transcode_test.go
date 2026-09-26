package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// transcodePlanFixture copies a sample video to a new originals folder as clip.avi and returns the config,
// a matching search result, and the folder's name relative to the originals and sidecar paths.
func transcodePlanFixture(t *testing.T) (*config.Config, []search.Photo, string) {
	t.Helper()

	conf := get.Config()
	saved := ffmpeg.Exclude()
	ffmpeg.SetExclude(video.NewFormats(""))
	t.Cleanup(func() { ffmpeg.SetExclude(saved) })

	dir, err := os.MkdirTemp(conf.OriginalsPath(), "video-transcode-")
	require.NoError(t, err)
	folder := filepath.Base(dir)
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
		_ = os.RemoveAll(filepath.Join(conf.SidecarPath(), folder))
	})

	src := filepath.Join(dir, "clip.avi")
	require.NoError(t, fs.Copy(filepath.Join(conf.SamplesPath(), "christmas.mp4"), src, false))
	require.NoDirExists(t, filepath.Join(conf.SidecarPath(), folder))

	result := search.Photo{PhotoType: entity.MediaVideo, Files: []entity.File{{FileRoot: entity.RootOriginals,
		FileName: folder + "/clip.avi", FileVideo: true, FileCodec: "mp4v", FileType: fs.VideoAVI.String(),
		MediaType: entity.MediaVideo, FileSize: fs.FileSize(src)}}}

	return conf, []search.Photo{result}, folder
}

func TestVideoBuildTranscodePlans(t *testing.T) {
	t.Run("MissingSidecarFolder", func(t *testing.T) {
		conf, results, folder := transcodePlanFixture(t)

		plans, preflight, err := videoBuildTranscodePlans(conf, get.Convert(), results, false)
		require.NoError(t, err)
		require.Len(t, plans, 1)

		// The plan names the file the converter writes, and its preflight creates no folder.
		assert.Equal(t, filepath.Join(conf.SidecarPath(), folder, "clip.avi.avc"), plans[0].DestPath)
		require.NoError(t, videoCheckFreeSpace(preflight))
		assert.NoDirExists(t, filepath.Join(conf.SidecarPath(), folder))
	})
	t.Run("SameOutput", func(t *testing.T) {
		conf, results, _ := transcodePlanFixture(t)

		// Two results with the same output, like the lenses of an Insta360 capture, are planned once.
		plans, preflight, err := videoBuildTranscodePlans(conf, get.Convert(), append(results, results[0]), false)
		require.NoError(t, err)
		assert.Len(t, plans, 1)
		assert.Len(t, preflight, 1)
	})
	t.Run("ExistingSidecar", func(t *testing.T) {
		conf, results, folder := transcodePlanFixture(t)
		existing := filepath.Join(conf.SidecarPath(), folder, "clip.avi.avc")
		require.NoError(t, fs.Copy(filepath.Join(conf.SamplesPath(), "christmas.mp4"), existing, false))

		plans, _, err := videoBuildTranscodePlans(conf, get.Convert(), results, false)
		require.NoError(t, err)
		assert.Empty(t, plans)

		// A sidecar transcode is replaced with --force.
		plans, _, err = videoBuildTranscodePlans(conf, get.Convert(), results, true)
		require.NoError(t, err)
		require.Len(t, plans, 1)
		assert.Equal(t, existing, plans[0].DestPath)
	})
	t.Run("ExistingOriginal", func(t *testing.T) {
		conf, results, folder := transcodePlanFixture(t)
		existing := filepath.Join(conf.OriginalsPath(), folder, "clip.avc")
		require.NoError(t, fs.Copy(filepath.Join(conf.SamplesPath(), "christmas.mp4"), existing, false))

		// The converter keeps a transcode next to the original even with --force.
		for _, force := range []bool{false, true} {
			plans, _, err := videoBuildTranscodePlans(conf, get.Convert(), results, force)
			require.NoError(t, err)
			assert.Empty(t, plans)
		}
	})
}

func TestVideoTranscodeTarget(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		conf, _, folder := transcodePlanFixture(t)
		src := filepath.Join(conf.OriginalsPath(), folder, "clip.avi")

		dest, existing, err := videoTranscodeTarget(get.Convert(), src)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(conf.SidecarPath(), folder, "clip.avi.avc"), dest)
		assert.Nil(t, existing)
	})
	t.Run("Existing", func(t *testing.T) {
		conf, _, folder := transcodePlanFixture(t)
		src := filepath.Join(conf.OriginalsPath(), folder, "clip.avi")
		want := filepath.Join(conf.SidecarPath(), folder, "clip.avi.avc")
		require.NoError(t, fs.Copy(filepath.Join(conf.SamplesPath(), "christmas.mp4"), want, false))

		dest, existing, err := videoTranscodeTarget(get.Convert(), src)
		require.NoError(t, err)
		assert.Equal(t, want, dest)
		require.NotNil(t, existing)
		assert.Equal(t, want, existing.FileName())
	})
	t.Run("MissingFile", func(t *testing.T) {
		_, _, err := videoTranscodeTarget(get.Convert(), filepath.Join(t.TempDir(), "clip.avi"))
		assert.Error(t, err)
	})
	t.Run("NoConvert", func(t *testing.T) {
		_, _, err := videoTranscodeTarget(nil, "clip.avi")
		assert.Error(t, err)
	})
}
