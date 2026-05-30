package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/video"
)

func TestVideoBuildRemuxPlans(t *testing.T) {
	t.Run("CountsExcludedFilesAsSkipped", func(t *testing.T) {
		conf := get.Config()
		require.NotNil(t, conf)

		saved := ffmpeg.Exclude()
		ffmpeg.SetExclude(video.NewFormats("avi"))
		t.Cleanup(func() { ffmpeg.SetExclude(saved) })

		relPath := "testdata/remux-excluded.avi"
		absPath := fs.Abs(relPath)
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
