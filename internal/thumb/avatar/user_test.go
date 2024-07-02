package avatar

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestSetUserAvatarURL(t *testing.T) {
	t.Run("PNG", func(t *testing.T) {
		admin := entity.UserFixtures.Get("alice")
		imageUrl := "https://dl.photoprism.app/icons/logo/256.png"
		err := SetUserImageURL(&admin, imageUrl, entity.SrcAuto)
		assert.NoError(t, err)
	})
	t.Run("JPEG", func(t *testing.T) {
		admin := entity.UserFixtures.Get("alice")
		imageUrl := "https://dl.photoprism.app/img/team/avatar.jpg"
		err := SetUserImageURL(&admin, imageUrl, entity.SrcOIDC)
		assert.NoError(t, err)
	})
	t.Run("NotFound", func(t *testing.T) {
		admin := entity.UserFixtures.Get("alice")
		imageUrl := "https://dl.photoprism.app/img/team/avatar-invalid.jpg"
		err := SetUserImageURL(&admin, imageUrl, entity.SrcAuto)
		assert.Error(t, err)
	})
}

func TestSetUserAvatarImage(t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		admin := entity.UserFixtures.Get("alice")
		fileName := fs.Abs("testdata/avatar.png")
		err := SetUserImage(&admin, fileName, entity.SrcAuto)
		assert.NoError(t, err)
	})
}
