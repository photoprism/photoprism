package avatar

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestSetUserAvatarURL(t *testing.T) {
	thumbPath := fs.Abs("testdata/cache")

	t.Run("PNG", func(t *testing.T) {
		admin := entity.UserFixtures.Get("alice")
		imageUrl := "https://dl.photoprism.app/icons/logo/256.png"
		err := SetUserImageURL(&admin, imageUrl, entity.SrcAuto, thumbPath)
		assert.NoError(t, err)
	})
	t.Run("JPEG", func(t *testing.T) {
		admin := entity.UserFixtures.Get("bob")
		imageUrl := "https://dl.photoprism.app/img/team/avatar.jpg"
		err := SetUserImageURL(&admin, imageUrl, entity.SrcOIDC, thumbPath)
		assert.NoError(t, err)
	})
	t.Run("NotFound", func(t *testing.T) {
		admin := entity.UserFixtures.Get("alice")
		imageUrl := "https://dl.photoprism.app/img/team/avatar-invalid.jpg"
		err := SetUserImageURL(&admin, imageUrl, entity.SrcAuto, thumbPath)
		assert.Error(t, err)
	})
	t.Run("EmptyUrl", func(t *testing.T) {
		admin := entity.UserFixtures.Get("alice")
		err := SetUserImageURL(&admin, "", entity.SrcAuto, thumbPath)
		assert.Nil(t, err)
	})
}

func TestSetUserAvatarImage(t *testing.T) {
	thumbPath := fs.Abs("testdata/cache")

	t.Run("Admin", func(t *testing.T) {
		admin := entity.UserFixtures.Get("friend")
		fileName := fs.Abs("testdata/avatar.png")
		err := SetUserImage(&admin, fileName, entity.SrcAuto, thumbPath)
		assert.NoError(t, err)
	})
	t.Run("FileNotFound", func(t *testing.T) {
		admin := entity.UserFixtures.Get("friend")
		fileName := fs.Abs("testdata/avatar-wrong.png")
		err := SetUserImage(&admin, fileName, entity.SrcAuto, thumbPath)
		assert.Error(t, err)
	})
	t.Run("ThumbPathEmpty", func(t *testing.T) {
		admin := entity.UserFixtures.Get("friend")
		fileName := fs.Abs("testdata/avatar-wrong.png")
		err := SetUserImage(&admin, fileName, entity.SrcAuto, "")
		assert.Error(t, err)
	})
}
