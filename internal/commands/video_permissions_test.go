package commands

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestVideoTrimFileCreationMode checks publication into originals with ordinary creation modes.
func TestVideoTrimFileCreationMode(t *testing.T) {
	conf, _ := remuxPlanFixture(t, "clip.mp4")
	src := filepath.Join(conf.OriginalsPath(), "clip.mp4")
	control := filepath.Join(conf.OriginalsPath(), "mode-control")
	require.NoError(t, os.WriteFile(control, nil, fs.ModeFile))
	want, err := os.Stat(control)
	require.NoError(t, err)
	stub := filepath.Join(t.TempDir(), "ffmpeg")
	require.NoError(t, os.WriteFile(stub, []byte(`#!/bin/sh
for output do :; done
printf 'trimmed' > "$output"
`), fs.ModeDir))
	conf.Options().FFmpegBin = stub
	plan := videoTrimPlan{SrcPath: src, DestPath: src, Duration: 3 * time.Second}
	require.ErrorContains(t, videoTrimFile(conf, nil, plan, time.Second, true), "missing filename")
	got, err := os.Stat(src)
	require.NoError(t, err)
	require.Equal(t, want.Mode().Perm(), got.Mode().Perm())
	t.Logf("trim mode: %04o", got.Mode().Perm())
}

// TestVideoTranscodeActionCreationMode checks the CLI action's output permissions after conversion.
func TestVideoTranscodeActionCreationMode(t *testing.T) {
	ffmpegBin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is required for the one-frame transcode fixture")
	}
	conf := get.Config()
	saved := *conf.Options()
	t.Cleanup(func() { *conf.Options() = saved })
	conf.Options().DisableFaces = true
	conf.Options().DisableClassification = true
	conf.Options().TranscodeTimeout = 20
	dir, err := os.MkdirTemp(conf.OriginalsPath(), "video-mode-")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(dir)) })
	src := filepath.Join(dir, "clip.avi")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, ffmpegBin, "-v", "error", "-f", "lavfi", "-i", "color=c=black:s=32x32:r=1", "-frames:v", "1", "-c:v", "mpeg4", src) //nolint:gosec // trusted test binary and generated fixture path
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	media, err := photoprism.NewMediaFile(src)
	require.NoError(t, err)
	photo := entity.NewPhoto(false)
	photo.PhotoType, photo.PhotoQuality, photo.PhotoName = entity.MediaVideo, 3, "permission fixture"
	require.NoError(t, photo.Create())
	file := entity.File{FileUID: rnd.GenerateUID(entity.FileUID), PhotoUID: photo.PhotoUID, PhotoID: photo.ID, FileName: media.RelName(conf.OriginalsPath()), FileRoot: entity.RootOriginals, FileHash: media.Hash(), FileSize: media.FileSize(), FileVideo: true, FilePrimary: true, FileType: "avi", FileCodec: "mp4v"}
	require.NoError(t, file.Create())
	t.Cleanup(func() {
		require.NoError(t, entity.UnscopedDb().Where("photo_uid = ?", photo.PhotoUID).Delete(&entity.File{}).Error)
		require.NoError(t, entity.UnscopedDb().Delete(&photo).Error)
	})
	dest, err := fs.FileName(src, conf.SidecarPath(), conf.OriginalsPath(), fs.ExtAvc)
	require.NoError(t, err)
	outputDir := filepath.Join(conf.SidecarPath(), filepath.Base(dir))
	require.Equal(t, outputDir, filepath.Dir(dest))
	require.NoError(t, fs.MkdirAll(outputDir))
	t.Cleanup(func() { require.NoError(t, os.RemoveAll(outputDir)) })
	control := filepath.Join(dir, "mode-control")
	require.NoError(t, os.WriteFile(control, nil, fs.ModeFile))
	want, err := os.Stat(control)
	require.NoError(t, err)
	_, err = RunWithTestContext(VideoTranscodeCommand, []string{"transcode", "--yes", "uid:" + photo.PhotoUID})
	require.NoError(t, err)
	got, err := os.Stat(dest)
	require.NoError(t, err, "the command must actually produce output")
	require.Positive(t, got.Size())
	require.Zero(t, got.Mode().Perm()&^want.Mode().Perm())
	t.Logf("transcode mode: %04o", got.Mode().Perm())
	require.NoError(t, os.Chmod(dest, fs.ModeSecretFile))
	_, err = RunWithTestContext(VideoTranscodeCommand, []string{"transcode", "--yes", "uid:" + photo.PhotoUID})
	require.NoError(t, err)
	got, err = os.Stat(dest)
	require.NoError(t, err)
	require.Equal(t, fs.ModeSecretFile, got.Mode().Perm(), "existing output keeps its permissions")
}
