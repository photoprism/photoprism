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
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// remuxPlanFixture creates an isolated selection whose indexed codecs need no external probe.
func remuxPlanFixture(t *testing.T, names ...string) (*config.Config, []search.Photo) {
	t.Helper()
	conf := config.NewMinimalTestConfig(t.TempDir())
	require.NoError(t, os.MkdirAll(conf.OriginalsPath(), fs.ModeDir))
	saved := ffmpeg.Exclude()
	ffmpeg.SetExclude(video.NewFormats(""))
	t.Cleanup(func() { ffmpeg.SetExclude(saved) })
	results := make([]search.Photo, 0, len(names))
	for _, name := range names {
		path := filepath.Join(conf.OriginalsPath(), name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), fs.ModeDir))
		require.NoError(t, os.WriteFile(path, []byte("original "+name), fs.ModeFile))
		results = append(results, search.Photo{Files: []entity.File{{FileRoot: entity.RootOriginals, FileName: name, FileVideo: true, FileCodec: video.CodecAvc1, FileType: fs.FileType(name).String()}}})
	}
	return conf, results
}

// TestVideoBuildRemuxPlans_Batch validates the complete plan without modifying selected files.
func TestVideoBuildRemuxPlans_Batch(t *testing.T) {
	t.Run("DuplicateOutputs", func(t *testing.T) {
		for _, force := range []bool{false, true} {
			conf, results := remuxPlanFixture(t, "clip.mts", "clip.tod")
			plans, preflight, _, err := videoBuildRemuxPlans(conf, results, force)
			require.ErrorContains(t, err, "selected more than once")
			for _, name := range []string{"clip.mts", "clip.tod", "clip.mp4"} {
				assert.Contains(t, err.Error(), filepath.Join(conf.OriginalsPath(), name))
			}
			assert.Empty(t, plans)
			assert.Empty(t, preflight)
			for _, name := range []string{"clip.mts", "clip.tod"} {
				data, readErr := os.ReadFile(filepath.Join(conf.OriginalsPath(), name)) // #nosec G304 -- the fixture owns this temporary originals directory.
				require.NoError(t, readErr)
				assert.Equal(t, "original "+name, string(data))
			}
			assert.NoFileExists(t, filepath.Join(conf.OriginalsPath(), "clip.mp4"))
		}
	})
	t.Run("ReadOnlySidecarOutputs", func(t *testing.T) {
		conf, results := remuxPlanFixture(t, "clip.mts", "clip.tod")
		conf.Options().ReadOnly = true
		require.NoError(t, os.MkdirAll(conf.SidecarPath(), fs.ModeDir))
		plans, preflight, _, err := videoBuildRemuxPlans(conf, results, false)
		require.ErrorContains(t, err, "selected more than once")
		assert.Empty(t, plans)
		assert.Empty(t, preflight)
	})
	t.Run("OtherInput", func(t *testing.T) {
		conf, results := remuxPlanFixture(t, "clip.mts", "clip.mp4")
		plans, preflight, _, err := videoBuildRemuxPlans(conf, results, true)
		require.ErrorContains(t, err, "another selected input")
		assert.Empty(t, plans)
		assert.Empty(t, preflight)
		data, readErr := os.ReadFile(filepath.Join(conf.OriginalsPath(), "clip.mp4"))
		require.NoError(t, readErr)
		assert.Equal(t, "original clip.mp4", string(data))
	})
	t.Run("SkippedInputIsPreserved", func(t *testing.T) {
		conf, results := remuxPlanFixture(t, "clip.mts", "clip.mp4")
		ffmpeg.SetExclude(video.NewFormats("mp4"))
		plans, _, skipped, err := videoBuildRemuxPlans(conf, results, true)
		require.ErrorContains(t, err, "another selected input")
		assert.Empty(t, plans)
		assert.Equal(t, 1, skipped)
	})
	t.Run("NoForceKeepsExistingOutput", func(t *testing.T) {
		conf, results := remuxPlanFixture(t, "clip.mts", "clip.mp4")
		plans, _, skipped, err := videoBuildRemuxPlans(conf, results, false)
		require.NoError(t, err)
		require.Len(t, plans, 1)
		assert.Equal(t, 1, skipped)
		assert.Equal(t, plans[0].SrcPath, plans[0].DestPath)
	})
	t.Run("IndependentAndSameFile", func(t *testing.T) {
		conf, results := remuxPlanFixture(t, "one.mts", "two.mp4", "other/one.tod")
		for _, force := range []bool{false, true} {
			plans, preflight, skipped, err := videoBuildRemuxPlans(conf, results, force)
			require.NoError(t, err)
			require.Len(t, plans, 3)
			assert.Len(t, preflight, 3)
			assert.Zero(t, skipped)
			assert.Equal(t, plans[1].SrcPath, plans[1].DestPath)
		}
	})
}

// TestVideoRemuxPath checks entry identity without following the final component.
func TestVideoRemuxPath(t *testing.T) {
	t.Run("DirectoryAlias", func(t *testing.T) {
		root := t.TempDir()
		dir := filepath.Join(root, "actual")
		alias := filepath.Join(root, "alias")
		require.NoError(t, os.Mkdir(dir, fs.ModeDir))
		require.NoError(t, os.Symlink(dir, alias))
		actual, err := videoRemuxPath(filepath.Join(dir, "new", "out.mp4"))
		require.NoError(t, err)
		other, err := videoRemuxPath(filepath.Join(alias, "new", "out.mp4"))
		require.NoError(t, err)
		assert.Equal(t, actual, other)
		link := filepath.Join(dir, "linked.mp4")
		require.NoError(t, os.Symlink("out.mp4", link))
		key, err := videoRemuxPath(link)
		require.NoError(t, err)
		assert.Equal(t, link, key)
	})
	t.Run("Invalid", func(t *testing.T) {
		_, err := videoRemuxPath("")
		require.Error(t, err)
		root := t.TempDir()
		link := filepath.Join(root, "missing")
		require.NoError(t, os.Symlink(filepath.Join(root, "absent"), link))
		_, err = videoRemuxPath(filepath.Join(link, "output.mp4"))
		require.Error(t, err)
		loop := filepath.Join(root, "loop")
		require.NoError(t, os.Symlink(loop, loop))
		_, err = videoRemuxPath(filepath.Join(loop, "output.mp4"))
		require.Error(t, err)
	})
}

// TestVideoValidateRemuxPlans checks aliases and keeps intentional self-remuxing distinct.
func TestVideoValidateRemuxPlans(t *testing.T) {
	t.Run("DirectoryAliases", func(t *testing.T) {
		root := t.TempDir()
		alias := filepath.Join(t.TempDir(), "alias")
		require.NoError(t, os.Symlink(root, alias))
		a, b := filepath.Join(root, "one.mts"), filepath.Join(root, "two.mts")
		require.NoError(t, os.WriteFile(a, []byte("one"), fs.ModeFile))
		require.NoError(t, os.WriteFile(b, []byte("two"), fs.ModeFile))
		plans := []videoRemuxPlan{{SrcPath: a, DestPath: filepath.Join(root, "out.mp4")}, {SrcPath: b, DestPath: filepath.Join(alias, "out.mp4")}}
		require.ErrorContains(t, videoValidateRemuxPlans(plans, []string{a, b}), "selected more than once")
	})
	t.Run("InputSymlink", func(t *testing.T) {
		root := t.TempDir()
		a, b, alias := filepath.Join(root, "one.mts"), filepath.Join(root, "one.mp4"), filepath.Join(root, "input.mov")
		require.NoError(t, os.WriteFile(a, []byte("one"), fs.ModeFile))
		require.NoError(t, os.WriteFile(b, []byte("two"), fs.ModeFile))
		require.NoError(t, os.Symlink(b, alias))
		plans := []videoRemuxPlan{{SrcPath: a, DestPath: b}}
		err := videoValidateRemuxPlans(plans, []string{a, alias})
		require.ErrorContains(t, err, "another selected input")
		for _, name := range []string{a, b, alias} {
			assert.Contains(t, err.Error(), name)
		}
		require.NoError(t, videoValidateRemuxPlans([]videoRemuxPlan{{SrcPath: b, DestPath: b}}, []string{b}))
	})
	t.Run("InvalidInput", func(t *testing.T) {
		require.Error(t, videoValidateRemuxPlans(nil, []string{""}))
		require.Error(t, videoValidateRemuxPlans(nil, []string{filepath.Join(t.TempDir(), "missing")}))
		require.Error(t, videoValidateRemuxPlans([]videoRemuxPlan{{}}, nil))
		require.Error(t, videoValidateRemuxPlans([]videoRemuxPlan{{SrcPath: "input.mp4"}}, nil))
		require.NoError(t, videoValidateRemuxPlans(nil, nil))
	})
}
