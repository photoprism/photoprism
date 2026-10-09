package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_Stage(t *testing.T) {
	opt := Opt(true, true, nil)

	t.Run("Main", func(t *testing.T) {
		assert.Equal(t, StageMain, opt.StageName())
		assert.True(t, opt.AutoMigrate)
	})
	t.Run("Pre", func(t *testing.T) {
		pre := opt.Pre()
		assert.Equal(t, StagePre, pre.StageName())
		assert.False(t, pre.AutoMigrate)
		assert.True(t, pre.RunFailed)
	})
	t.Run("Post", func(t *testing.T) {
		post := opt.Post()
		assert.Equal(t, StagePost, post.StageName())
		assert.False(t, post.AutoMigrate)
		assert.True(t, post.RunFailed)
	})
	t.Run("PostMigrationSkipped", func(t *testing.T) {
		m := Migration{ID: "20260101-000001", Stage: StagePost}
		assert.True(t, m.Skip(opt))
		assert.False(t, m.Skip(opt.Post()))
	})
}
