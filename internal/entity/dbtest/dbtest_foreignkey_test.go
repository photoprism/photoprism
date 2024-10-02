package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestDbtestForeignKey_Validate(t *testing.T) {
	dbtestMutex.Lock()
	defer dbtestMutex.Unlock()

	t.Run("Photos_CameraID", func(t *testing.T) {
		m := &entity.Photo{}
		m.CameraID = 123412341234
		stmt := entity.Db()
		res := stmt.Create(m)
		assert.Error(t, res.Error)
		assert.Error(t, res.Error, "foreign key constraint")
	})

	t.Run("Photos_LensID", func(t *testing.T) {
		m := &entity.Photo{}
		m.LensID = 123412341234
		stmt := entity.Db()
		res := stmt.Create(m)
		assert.Error(t, res.Error)
		assert.Error(t, res.Error, "foreign key constraint")
	})
}
