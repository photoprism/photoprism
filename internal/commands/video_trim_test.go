package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// TestVideoTrimFastStart verifies which extensions get the faststart flag.
func TestVideoTrimFastStart(t *testing.T) {
	assert.True(t, videoTrimFastStart("clip.mp4"))
	assert.True(t, videoTrimFastStart("clip.MOV"))
	assert.True(t, videoTrimFastStart("clip.m4v"))
	assert.True(t, videoTrimFastStart("clip.qt"))
	assert.False(t, videoTrimFastStart("clip.mkv"))
	assert.False(t, videoTrimFastStart(""))
}

// TestVideoTrimSidecar verifies that a trim of read-only originals runs when its sidecar folder does not
// exist yet: the plan and its free-space check create no folder, and the trim creates it.
func TestVideoTrimSidecar(t *testing.T) {
	conf, results, folder := transcodePlanFixture(t)
	saved := *conf.Options()
	t.Cleanup(func() { *conf.Options() = saved })
	conf.Options().ReadOnly = true
	results[0].Files[0].FileDuration = 3 * time.Second
	sidecarDir := filepath.Join(conf.SidecarPath(), folder)

	plans, preflight, err := videoBuildTrimPlans(conf, results, time.Second)
	require.NoError(t, err)
	require.Len(t, plans, 1)
	require.True(t, plans[0].Sidecar)
	assert.Equal(t, filepath.Join(sidecarDir, "clip.avi"), plans[0].DestPath)
	require.NoError(t, videoCheckFreeSpace(preflight))
	require.NoDirExists(t, sidecarDir)

	stub := filepath.Join(t.TempDir(), "ffmpeg")
	require.NoError(t, os.WriteFile(stub, []byte(`#!/bin/sh
for output do :; done
printf 'trimmed' > "$output"
`), fs.ModeDir))
	conf.Options().FFmpegBin = stub

	// The reindex that follows needs the indexed file, which this fixture does not create.
	plan := plans[0]
	plan.IndexPath = ""
	require.ErrorContains(t, videoTrimFile(conf, nil, plan, time.Second, true), "missing filename")

	data, err := os.ReadFile(plan.DestPath) // #nosec G304 -- the fixture owns this sidecar folder.
	require.NoError(t, err)
	assert.Equal(t, "trimmed", string(data))
	assert.FileExists(t, filepath.Join(conf.OriginalsPath(), folder, "clip.avi"))
}

// TestVideoTrimFile_SidecarLink verifies that a sidecar trim does not replace a link at its destination.
func TestVideoTrimFile_SidecarLink(t *testing.T) {
	dir := t.TempDir()
	conf, _ := remuxPlanFixture(t, "clip.mp4")
	src := filepath.Join(conf.OriginalsPath(), "clip.mp4")
	dest := filepath.Join(dir, "clip.mp4")
	require.NoError(t, os.Symlink(filepath.Join(dir, "absent"), dest))
	stub := filepath.Join(t.TempDir(), "ffmpeg")
	require.NoError(t, os.WriteFile(stub, []byte(`#!/bin/sh
for output do :; done
printf 'trimmed' > "$output"
`), fs.ModeDir))
	conf.Options().FFmpegBin = stub

	plan := videoTrimPlan{SrcPath: src, DestPath: dest, Duration: 3 * time.Second, Sidecar: true}
	require.ErrorContains(t, videoTrimFile(conf, nil, plan, time.Second, true), "already exists")
	assert.True(t, fs.IsSymlink(dest))

	// The staged output is removed.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}
