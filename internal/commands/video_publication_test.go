package commands

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// TestVideoPreserveMode checks destination metadata and failures before publication.
func TestVideoPreserveMode(t *testing.T) {
	t.Run("RegularDestination", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "clip.mp4")
		require.NoError(t, os.WriteFile(dest, []byte("original"), fs.ModeFile))
		require.NoError(t, os.Chmod(dest, fs.ModeSecretFile))
		staged, err := fs.CreateStageFile(dest)
		require.NoError(t, err)
		require.NoError(t, videoPreserveMode(staged, dest))
		info, err := os.Stat(staged)
		require.NoError(t, err)
		require.Equal(t, fs.ModeSecretFile, info.Mode().Perm())
		require.Error(t, videoPreserveMode(filepath.Join(dir, "missing"), dest))
	})
	t.Run("NewDestination", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "clip.mp4")
		staged, err := fs.CreateStageFile(dest)
		require.NoError(t, err)
		before, err := os.Stat(staged)
		require.NoError(t, err)
		require.NoError(t, videoPreserveMode(staged, dest))
		after, err := os.Stat(staged)
		require.NoError(t, err)
		require.Equal(t, before.Mode(), after.Mode())
	})
	t.Run("SymlinkDestination", func(t *testing.T) {
		dir := t.TempDir()
		dest := filepath.Join(dir, "clip.mp4")
		require.NoError(t, os.Symlink(filepath.Join(dir, "absent"), dest))
		staged, err := fs.CreateStageFile(dest)
		require.NoError(t, err)
		require.NoError(t, videoPreserveMode(staged, dest))
		require.True(t, fs.IsSymlink(dest))
	})
	t.Run("InvalidParent", func(t *testing.T) {
		parent := filepath.Join(t.TempDir(), "file")
		require.NoError(t, os.WriteFile(parent, nil, fs.ModeFile))
		require.Error(t, videoPreserveMode("unused", filepath.Join(parent, "clip.mp4")))
	})
}

// TestVideoPublicationMode checks reserved siblings and permissions on replacement and new output.
func TestVideoPublicationMode(t *testing.T) {
	for _, tc := range []struct {
		name     string
		remux    bool
		separate bool
		sidecar  bool
		noBackup bool
		newFile  bool
		noExt    bool
	}{
		{name: "RemuxSameFile", remux: true},
		{name: "RemuxSeparateFile", remux: true, separate: true},
		{name: "RemuxSidecar", remux: true, separate: true, sidecar: true},
		{name: "RemuxNewFile", remux: true, separate: true, newFile: true},
		{name: "TrimWithBackup"},
		{name: "TrimWithoutBackup", noBackup: true},
		{name: "TrimNewSidecar", separate: true, sidecar: true, newFile: true},
		{name: "TrimExtensionFallback", separate: true, sidecar: true, newFile: true, noExt: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, mode := range []os.FileMode{fs.ModeSecretFile, 0o640} {
				conf, _ := remuxPlanFixture(t, "clip.mp4")
				src := filepath.Join(conf.OriginalsPath(), "clip.mp4")
				dest := src
				if tc.separate {
					dest = filepath.Join(conf.OriginalsPath(), "clip.out.mp4")
				}
				if tc.noExt {
					dest = filepath.Join(conf.OriginalsPath(), "clip-output")
				}
				if !tc.newFile {
					require.NoError(t, os.WriteFile(dest, []byte("original"), fs.ModeFile))
					require.NoError(t, os.Chmod(dest, mode))
				}
				control := filepath.Join(conf.OriginalsPath(), "mode-control")
				require.NoError(t, os.WriteFile(control, nil, fs.ModeFile))
				want, err := os.Stat(control)
				require.NoError(t, err)
				stub := filepath.Join(t.TempDir(), "ffmpeg")
				require.NoError(t, os.WriteFile(stub, []byte(`#!/bin/sh
for output do :; done
test -f "$output" || exit 41
dir=${output%/*}
for reservation in "$dir"/.clip*.tmp*; do
  test -f "$reservation" || exit 42
done
case "$output" in *.mp4) ;; *) exit 43 ;; esac
printf 'converted' > "$output"
`), fs.ModeDir))
				conf.Options().FFmpegBin = stub
				if tc.remux {
					err = videoRemuxFile(conf, nil, videoRemuxPlan{SrcPath: src, DestPath: dest, Sidecar: tc.sidecar}, true)
				} else {
					err = videoTrimFile(conf, nil, videoTrimPlan{SrcPath: src, DestPath: dest, Sidecar: tc.sidecar, Duration: 3 * time.Second}, time.Second, tc.noBackup)
				}
				require.ErrorContains(t, err, "missing filename", "publication must reach the reindex boundary")
				data, err := os.ReadFile(dest) //nolint:gosec // the fixture owns this temporary path
				require.NoError(t, err)
				require.Equal(t, "converted", string(data))
				got, err := os.Stat(dest)
				require.NoError(t, err)
				if tc.newFile {
					require.Equal(t, want.Mode().Perm(), got.Mode().Perm())
				} else {
					require.Equal(t, mode, got.Mode().Perm())
				}
				if !tc.remux && !tc.noBackup && !tc.sidecar {
					backup, statErr := os.Stat(dest + ".backup")
					require.NoError(t, statErr)
					require.Equal(t, fs.ModeBackupFile, backup.Mode().Perm())
				}
				staged, err := filepath.Glob(filepath.Join(conf.OriginalsPath(), ".*.tmp*"))
				require.NoError(t, err)
				require.Empty(t, staged)
			}
		})
	}
}

// TestVideoPublicationFFmpeg checks reserved-file output with real software remuxing and trimming.
func TestVideoPublicationFFmpeg(t *testing.T) {
	ffmpegBin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is required for the small video fixture")
	}
	for _, name := range []string{"Remux", "Trim"} {
		t.Run(name, func(t *testing.T) {
			conf, _ := remuxPlanFixture(t, "clip.mp4")
			src := filepath.Join(conf.OriginalsPath(), "clip.mp4")
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, ffmpegBin, "-y", "-v", "error", "-f", "lavfi", "-i", "color=c=black:s=32x32:r=2", "-t", "3", "-c:v", "mpeg4", "-g", "1", src) //nolint:gosec // trusted test binary and generated fixture path
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, string(output))
			require.NoError(t, os.Chmod(src, fs.ModeSecretFile))
			conf.Options().FFmpegBin = ffmpegBin
			if name == "Remux" {
				err = videoRemuxFile(conf, nil, videoRemuxPlan{SrcPath: src, DestPath: src}, true)
			} else {
				err = videoTrimFile(conf, nil, videoTrimPlan{SrcPath: src, DestPath: src, Duration: 3 * time.Second}, time.Second, true)
			}
			require.ErrorContains(t, err, "missing filename")
			info, err := os.Stat(src)
			require.NoError(t, err)
			require.Positive(t, info.Size())
			require.Equal(t, fs.ModeSecretFile, info.Mode().Perm())
		})
	}
}

// TestVideoPublicationFailure checks that failed subprocesses leave the original and no staged output.
func TestVideoPublicationFailure(t *testing.T) {
	for _, name := range []string{"Remux", "Trim"} {
		t.Run(name, func(t *testing.T) {
			conf, _ := remuxPlanFixture(t, "clip.mp4")
			src := filepath.Join(conf.OriginalsPath(), "clip.mp4")
			require.NoError(t, os.Chmod(src, fs.ModeSecretFile))
			stub := filepath.Join(t.TempDir(), "ffmpeg")
			require.NoError(t, os.WriteFile(stub, []byte(`#!/bin/sh
for output do :; done
printf 'partial' > "$output"
exit 1
`), fs.ModeDir))
			conf.Options().FFmpegBin = stub
			var err error
			if name == "Remux" {
				err = videoRemuxFile(conf, nil, videoRemuxPlan{SrcPath: src, DestPath: src}, true)
			} else {
				err = videoTrimFile(conf, nil, videoTrimPlan{SrcPath: src, DestPath: src, Duration: 3 * time.Second}, time.Second, true)
			}
			require.Error(t, err)
			require.NotContains(t, err.Error(), "missing filename")
			data, err := os.ReadFile(src) //nolint:gosec // the fixture owns this temporary path
			require.NoError(t, err)
			require.Equal(t, "original clip.mp4", string(data))
			info, err := os.Stat(src)
			require.NoError(t, err)
			require.Equal(t, fs.ModeSecretFile, info.Mode().Perm())
			staged, err := filepath.Glob(filepath.Join(conf.OriginalsPath(), ".*.tmp*"))
			require.NoError(t, err)
			require.Empty(t, staged)
		})
	}
}
