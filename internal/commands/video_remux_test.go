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
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// TestVideoBuildRemuxPlans covers filtered remux selections.
func TestVideoBuildRemuxPlans(t *testing.T) {
	t.Run("CountsExcludedFilesAsSkipped", func(t *testing.T) {
		conf := get.Config()
		require.NotNil(t, conf)

		saved := ffmpeg.Exclude()
		ffmpeg.SetExclude(video.NewFormats("avi"))
		t.Cleanup(func() { ffmpeg.SetExclude(saved) })

		relPath := "testdata/remux-excluded.avi"
		absPath := filepath.Join(conf.OriginalsPath(), filepath.FromSlash(relPath))
		require.NoError(t, os.MkdirAll(filepath.Dir(absPath), fs.ModeDir))
		require.NoError(t, os.WriteFile(absPath, []byte("test"), fs.ModeFile))
		t.Cleanup(func() {
			_ = os.Remove(absPath)
		})

		results := []search.Photo{{
			PhotoUID: "ptest-remux-excluded",
			Files: []entity.File{{
				FileRoot:  entity.RootOriginals,
				FileName:  relPath,
				FileVideo: true,
				FileCodec: video.CodecAvc1,
				FileType:  fs.VideoAVI.String(),
				FileSize:  4,
			}},
		}}

		plans, preflight, skipped, err := videoBuildRemuxPlans(conf, results, false)
		require.NoError(t, err)
		assert.Empty(t, plans)
		assert.Empty(t, preflight)
		assert.Equal(t, 1, skipped)
	})
}

// TestVideoRemuxFile_PublishesPlan checks publication without an implicit backup destination.
func TestVideoRemuxFile_PublishesPlan(t *testing.T) {
	for _, sameFile := range []bool{false, true} {
		name := "SeparateOutput"
		if sameFile {
			name = "SameFile"
		}
		t.Run(name, func(t *testing.T) {
			conf, _ := remuxPlanFixture(t, "clip.mts")
			src := filepath.Join(conf.OriginalsPath(), "clip.mts")
			dest := filepath.Join(conf.OriginalsPath(), "clip.mp4")
			if sameFile {
				dest = src
			}
			backup := src + ".backup"
			require.NoError(t, os.WriteFile(backup, []byte("existing backup"), fs.ModeFile))
			stub := filepath.Join(t.TempDir(), "ffmpeg")
			require.NoError(t, os.WriteFile(stub, []byte(`#!/bin/sh
for output do :; done
printf 'remuxed' > "$output"
`), fs.ModeDir))
			conf.Options().FFmpegBin = stub

			// Empty IndexPath stops at the reindex boundary after publication.
			err := videoRemuxFile(conf, nil, videoRemuxPlan{SrcPath: src, DestPath: dest}, true)
			require.ErrorContains(t, err, "missing filename")
			data, err := os.ReadFile(dest) // #nosec G304 -- the fixture owns this temporary path.
			require.NoError(t, err)
			assert.Equal(t, "remuxed", string(data))
			want, err := os.Stat(backup)
			require.NoError(t, err)
			got, err := os.Stat(dest)
			require.NoError(t, err)
			assert.Equal(t, want.Mode().Perm(), got.Mode().Perm())
			t.Logf("remux mode: %04o", got.Mode().Perm())
			data, err = os.ReadFile(backup) // #nosec G304 -- the fixture owns this temporary path.
			require.NoError(t, err)
			assert.Equal(t, "existing backup", string(data))
			if !sameFile {
				data, err = os.ReadFile(src) // #nosec G304 -- the fixture owns this temporary path.
				require.NoError(t, err)
				assert.Equal(t, "original clip.mts", string(data))
			}
		})
	}
}

// TestVideoRemuxFile_SidecarFolder verifies that a sidecar remux creates its missing folder.
func TestVideoRemuxFile_SidecarFolder(t *testing.T) {
	conf, _ := remuxPlanFixture(t, "clip.mts")
	src := filepath.Join(conf.OriginalsPath(), "clip.mts")
	dest := filepath.Join(conf.SidecarPath(), "2026", "clip.mp4")
	stub := filepath.Join(t.TempDir(), "ffmpeg")
	require.NoError(t, os.WriteFile(stub, []byte(`#!/bin/sh
for output do :; done
printf 'remuxed' > "$output"
`), fs.ModeDir))
	conf.Options().FFmpegBin = stub
	require.True(t, conf.SidecarPathIsAbs())
	require.NoDirExists(t, filepath.Dir(dest))

	// Empty IndexPath stops at the reindex boundary after publication.
	err := videoRemuxFile(conf, nil, videoRemuxPlan{SrcPath: src, DestPath: dest, Sidecar: true}, false)
	require.ErrorContains(t, err, "missing filename")
	data, err := os.ReadFile(dest) // #nosec G304 -- the fixture owns this temporary path.
	require.NoError(t, err)
	assert.Equal(t, "remuxed", string(data))
}

// TestVideoRemuxFile_SidecarLink verifies that a sidecar remux without force does not replace a link at
// its destination.
func TestVideoRemuxFile_SidecarLink(t *testing.T) {
	dir := t.TempDir()
	conf, _ := remuxPlanFixture(t, "clip.mts")
	src := filepath.Join(conf.OriginalsPath(), "clip.mts")
	dest := filepath.Join(dir, "clip.mp4")
	require.NoError(t, os.Symlink(filepath.Join(dir, "absent"), dest))
	stub := filepath.Join(t.TempDir(), "ffmpeg")
	require.NoError(t, os.WriteFile(stub, []byte(`#!/bin/sh
for output do :; done
printf 'remuxed' > "$output"
`), fs.ModeDir))
	conf.Options().FFmpegBin = stub

	err := videoRemuxFile(conf, nil, videoRemuxPlan{SrcPath: src, DestPath: dest, Sidecar: true}, false)
	require.ErrorContains(t, err, "is a symbolic link")
	assert.True(t, fs.IsSymlink(dest))

	// The staged output is removed.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

// TestVideoRemuxFile_Replace verifies which existing outputs a remux replaces.
func TestVideoRemuxFile_Replace(t *testing.T) {
	// newRemux returns a config with an ffmpeg stub, and the source file of a remux.
	newRemux := func(t *testing.T) (*config.Config, string) {
		conf, _ := remuxPlanFixture(t, "clip.mts")
		stub := filepath.Join(t.TempDir(), "ffmpeg")
		require.NoError(t, os.WriteFile(stub, []byte(`#!/bin/sh
for output do :; done
printf 'remuxed' > "$output"
`), fs.ModeDir))
		conf.Options().FFmpegBin = stub
		return conf, filepath.Join(conf.OriginalsPath(), "clip.mts")
	}

	t.Run("SameFile", func(t *testing.T) {
		conf, src := newRemux(t)

		// A remux of a file into its own name replaces it without force.
		err := videoRemuxFile(conf, nil, videoRemuxPlan{SrcPath: src, DestPath: src}, false)
		require.ErrorContains(t, err, "missing filename")
		data, err := os.ReadFile(src) // #nosec G304 -- the fixture owns this temporary path.
		require.NoError(t, err)
		assert.Equal(t, "remuxed", string(data))
	})
	for _, sidecar := range []bool{false, true} {
		t.Run(map[bool]string{false: "ExistingOutput", true: "ExistingSidecar"}[sidecar], func(t *testing.T) {
			conf, src := newRemux(t)
			dest := filepath.Join(t.TempDir(), "clip.mp4")
			require.NoError(t, os.WriteFile(dest, []byte("existing"), fs.ModeFile))

			err := videoRemuxFile(conf, nil, videoRemuxPlan{SrcPath: src, DestPath: dest, Sidecar: sidecar}, false)
			require.EqualError(t, err, "output already exists "+clean.Log(dest))
			data, err := os.ReadFile(dest) // #nosec G304 -- the fixture owns this temporary path.
			require.NoError(t, err)
			assert.Equal(t, "existing", string(data))

			// With force, it is replaced.
			err = videoRemuxFile(conf, nil, videoRemuxPlan{SrcPath: src, DestPath: dest, Sidecar: sidecar}, true)
			require.ErrorContains(t, err, "missing filename")
			data, err = os.ReadFile(dest) // #nosec G304 -- the fixture owns this temporary path.
			require.NoError(t, err)
			assert.Equal(t, "remuxed", string(data))
		})
	}
	t.Run("LinkWithForce", func(t *testing.T) {
		conf, src := newRemux(t)
		dir := t.TempDir()
		target := filepath.Join(dir, "target.mp4")
		require.NoError(t, os.WriteFile(target, []byte("target"), fs.ModeFile))
		dest := filepath.Join(dir, "clip.mp4")
		require.NoError(t, os.Symlink(target, dest))

		// A link at a separate output is refused even with force, and its target is left as it is.
		err := videoRemuxFile(conf, nil, videoRemuxPlan{SrcPath: src, DestPath: dest}, true)
		require.ErrorContains(t, err, "is a symbolic link")
		assert.True(t, fs.IsSymlink(dest))
		data, err := os.ReadFile(target) // #nosec G304 -- the fixture owns this temporary path.
		require.NoError(t, err)
		assert.Equal(t, "target", string(data))
	})
}
