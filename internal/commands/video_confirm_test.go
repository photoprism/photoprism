package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// videoConfirmFixture creates a video original and its index rows, and returns the photo UID and the file path.
func videoConfirmFixture(t *testing.T, name, fileType, codec string, duration time.Duration) (photoUID, fileName string) {
	t.Helper()

	conf := get.Config()
	dir, err := os.MkdirTemp(conf.OriginalsPath(), "video-confirm-")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	fileName = filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(fileName, []byte("video fixture "+name), fs.ModeFile))

	photo := entity.NewPhoto(false)
	photo.PhotoType, photo.PhotoQuality, photo.PhotoName = entity.MediaVideo, 3, "confirmation fixture"
	require.NoError(t, photo.Create())

	file := entity.File{FileUID: rnd.GenerateUID(entity.FileUID), PhotoUID: photo.PhotoUID, PhotoID: photo.ID,
		FileName: filepath.Join(filepath.Base(dir), name), FileRoot: entity.RootOriginals, FileHash: rnd.GenerateUID('h'),
		FileSize: 32, FileVideo: true, FilePrimary: true, FileType: fileType, FileCodec: codec, FileDuration: duration}
	require.NoError(t, file.Create())

	t.Cleanup(func() {
		_ = entity.UnscopedDb().Where("photo_uid = ?", photo.PhotoUID).Delete(&entity.File{}).Error
		_ = entity.UnscopedDb().Delete(&photo).Error
	})

	return photo.PhotoUID, fileName
}

// assertVideoRefused runs a video command without confirmation: without a terminal it must exit 2, and answered
// with "n" it must exit 0. The exit status is what detects a command that proceeds, since the fixture is no real
// video; the original and the missing output are checked as well.
func assertVideoRefused(t *testing.T, cmd *cli.Command, args []string, fileName string, outputs ...string) {
	t.Helper()

	t.Setenv("PHOTOPRISM_CLI", "")

	before, err := os.ReadFile(fileName)
	require.NoError(t, err)
	stat, err := os.Stat(fileName)
	require.NoError(t, err)

	unchanged := func() {
		t.Helper()

		data, readErr := os.ReadFile(fileName)
		require.NoError(t, readErr)
		assert.Equal(t, before, data)

		after, statErr := os.Stat(fileName)
		require.NoError(t, statErr)
		assert.Equal(t, stat.ModTime(), after.ModTime())

		for _, output := range outputs {
			assert.NoFileExists(t, output)
		}
	}

	t.Run("NoTerminal", func(t *testing.T) {
		_, runErr := RunWithTestContext(cmd, args)

		var exit cli.ExitCoder
		require.ErrorAs(t, runErr, &exit)
		assert.Equal(t, 2, exit.ExitCode())
		assert.Contains(t, runErr.Error(), "could not ask for confirmation")
		unchanged()
	})
	t.Run("AnsweredNo", func(t *testing.T) {
		pipeResetAnswers(t, "n\n")

		_, runErr := RunWithTestContext(cmd, args)

		assert.NoError(t, runErr)
		unchanged()
	})
}

func TestVideoTrimCommand_Confirm(t *testing.T) {
	photoUID, fileName := videoConfirmFixture(t, "clip.mp4", fs.VideoMp4.String(), "avc1", 10*time.Second)
	assertVideoRefused(t, VideoTrimCommand, []string{"trim", "uid:" + photoUID, "2"}, fileName)
}

func TestVideoTranscodeCommand_Confirm(t *testing.T) {
	conf := get.Config()
	photoUID, fileName := videoConfirmFixture(t, "clip.avi", fs.VideoAVI.String(), "mp4v", 10*time.Second)

	// An indexed video has its preview image in the sidecar folder, so the folder exists.
	sidecarDir := filepath.Join(conf.SidecarPath(), filepath.Base(filepath.Dir(fileName)))
	require.NoError(t, fs.MkdirAll(sidecarDir))
	t.Cleanup(func() { _ = os.RemoveAll(sidecarDir) })
	dest := filepath.Join(sidecarDir, filepath.Base(fileName)+fs.ExtAvc)

	assertVideoRefused(t, VideoTranscodeCommand, []string{"transcode", "uid:" + photoUID}, fileName, dest)
}

func TestVideoRemuxCommand_Confirm(t *testing.T) {
	photoUID, fileName := videoConfirmFixture(t, "clip.mkv", fs.VideoMkv.String(), "avc1", 10*time.Second)
	assertVideoRefused(t, VideoRemuxCommand, []string{"remux", "uid:" + photoUID}, fileName, fs.StripKnownExt(fileName)+fs.ExtMp4)
}
