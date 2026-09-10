package entity

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var createsavetestMutex = sync.Mutex{}

func TestValidateSaveCreate(t *testing.T) {
	dbtestMutex.Lock()
	defer dbtestMutex.Unlock()

	// Use direct db Create to validate that values are set.
	// Passes on V1 and V2 with BeforeCreate being called.
	t.Run("BeforeCreate_Album", func(t *testing.T) {
		createsavetestMutex.Lock()
		defer createsavetestMutex.Unlock()
		m := &entity.Album{}
		assert.False(t, rnd.IsUID(m.AlbumUID, entity.AlbumUID))
		stmt := entity.Db()
		require.Error(t, stmt.Transaction(func(tx *gorm.DB) error {
			if tx.Error != nil {
				assert.Nil(t, tx.Error)
				t.FailNow()
				return tx.Error
			}
			require.NoError(t, tx.Create(m).Error)

			assert.NotEmpty(t, m.AlbumUID)
			assert.True(t, rnd.IsUID(m.AlbumUID, entity.AlbumUID))
			return errors.New("ForceRollback")
		}))
	})

	// Use direct db Save to validate that values are set.
	// Passes on V1 and V2 with BeforeCreate being called.

	t.Run("BeforeSave_Album", func(t *testing.T) {
		createsavetestMutex.Lock()
		defer createsavetestMutex.Unlock()
		m := &entity.Album{}
		assert.False(t, rnd.IsUID(m.AlbumUID, entity.AlbumUID))
		stmt := entity.Db()
		require.Error(t, stmt.Transaction(func(tx *gorm.DB) error {
			if tx.Error != nil {
				assert.Nil(t, tx.Error)
				t.FailNow()
				return tx.Error
			}
			require.NoError(t, tx.Save(m).Error)

			assert.NotEmpty(t, m.AlbumUID)
			assert.True(t, rnd.IsUID(m.AlbumUID, entity.AlbumUID))
			return errors.New("ForceRollback")
		}))

	})

	// Use direct db FirstOrCreate to validate that values are set.
	// Passes on V1 and V2 with BeforeCreate being called.

	t.Run("FirstOrCreate_Album", func(t *testing.T) {
		createsavetestMutex.Lock()
		defer createsavetestMutex.Unlock()
		m := &entity.Album{AlbumTitle: "Test Before Create"}
		assert.False(t, rnd.IsUID(m.AlbumUID, entity.AlbumUID))
		stmt := entity.Db()
		require.Error(t, stmt.Transaction(func(tx *gorm.DB) error {
			if tx.Error != nil {
				assert.Nil(t, tx.Error)
				t.FailNow()
				return tx.Error
			}

			found := entity.Album{}
			require.NoError(t, tx.FirstOrCreate(&found, m).Error)

			assert.NotEmpty(t, found.AlbumUID)
			assert.True(t, rnd.IsUID(found.AlbumUID, entity.AlbumUID))
			return errors.New("ForceRollback")
		}))

	})

	// Use direct db Save with child structs to validate that values are set.
	// Passes on V1 with BeforeCreate being called.
	t.Run("BeforeSave_PhotoAndDetails", func(t *testing.T) {
		createsavetestMutex.Lock()
		defer createsavetestMutex.Unlock()

		stmt := entity.Db()
		require.Error(t, stmt.Transaction(func(tx *gorm.DB) error {
			if tx.Error != nil {
				assert.Nil(t, tx.Error)
				t.FailNow()
				return tx.Error
			}

			labelotter := entity.Label{LabelName: "otter", LabelSlug: "otter"}
			var deletedAt = gorm.DeletedAt{Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
			labelsnake := entity.Label{LabelName: "snake", LabelSlug: "snake", DeletedAt: deletedAt}

			res := tx.Save(&labelsnake)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			res = tx.Save(&labelotter)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			details := &entity.Details{Keywords: "cow, flower, snake, otter"}
			photo := entity.Photo{ID: 34567, Details: details}

			// PostgreSQL will error all statements after a failure in a transaction.
			// So enable a savepoint to rollback to after the error we are about to create.
			if entity.DbDialect() == dsn.DialectPostgreSQL {
				tx.Exec("SAVEPOINT beforeError")
			}

			// Direct Save.  This will always fail with foreign key constraints on v2.
			log.Info("Expect 2 foreign key violation Error or SQLSTATE from dbtest_foreignkey_test")
			res = tx.Save(&photo)
			assert.Error(t, res.Error)
			if res.Error == nil {
				t.Log("Expected a foreign key error here")
				t.FailNow()
				return res.Error
			}
			if entity.DbDialect() == dsn.DialectPostgreSQL {
				tx.Exec("ROLLBACK TO SAVEPOINT beforeError")
			}

			assert.ErrorContains(t, res.Error, "constraint")

			// do the save the safe way.
			photo.Details = nil
			res = tx.Save(&photo)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			photo.Details = details
			res = tx.Save(&photo)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			photo2 := entity.Photo{ID: 34567}
			res = tx.Preload("Details").First(&photo2)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			assert.NotNil(t, photo2.Details)
			if photo2.Details != nil {
				assert.Equal(t, details.Keywords, photo2.Details.Keywords)
			}

			return errors.New("ForceRollback")
		}))
	})

	// Use direct db Save with child structs to validate that values are set.
	// Passes on V1 and V2 with BeforeCreate being called.
	t.Run("BeforeCreate_PhotoAndDetails", func(t *testing.T) {
		createsavetestMutex.Lock()
		defer createsavetestMutex.Unlock()

		stmt := entity.Db()
		require.Error(t, stmt.Transaction(func(tx *gorm.DB) error {
			if tx.Error != nil {
				assert.Nil(t, tx.Error)
				t.FailNow()
				return tx.Error
			}

			labelotter := entity.Label{LabelName: "otter", LabelSlug: "otter"}
			var deletedAt = gorm.DeletedAt{Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
			labelsnake := entity.Label{LabelName: "snake", LabelSlug: "snake", DeletedAt: deletedAt}

			res := tx.Save(&labelsnake)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			res = tx.Save(&labelotter)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			details := &entity.Details{Keywords: "cow, flower, snake, otter"}
			photo := entity.Photo{ID: 34567, Details: details}

			res = tx.Create(&photo)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			photo2 := entity.Photo{ID: 34567}
			res = tx.Preload("Details").First(&photo2)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			assert.NotNil(t, photo2.Details)
			if photo2.Details != nil {
				assert.Equal(t, details.Keywords, photo2.Details.Keywords)
			}

			return errors.New("ForceRollback")
		}))
	})

	// Use direct db Save with child structs to validate that values are set.
	// Works on V1 and V2 BUT Details are nil.
	t.Run("FirstOrCreate_PhotoAndDetails", func(t *testing.T) {
		createsavetestMutex.Lock()
		defer createsavetestMutex.Unlock()

		stmt := entity.Db()
		require.Error(t, stmt.Transaction(func(tx *gorm.DB) error {
			if tx.Error != nil {
				assert.Nil(t, tx.Error)
				t.FailNow()
				return tx.Error
			}

			labelotter := entity.Label{LabelName: "otter", LabelSlug: "otter"}
			var deletedAt = gorm.DeletedAt{Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
			labelsnake := entity.Label{LabelName: "snake", LabelSlug: "snake", DeletedAt: deletedAt}

			res := tx.Save(&labelsnake)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			res = tx.Save(&labelotter)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			details := &entity.Details{Keywords: "cow, flower, snake, otter"}
			photo := entity.Photo{ID: 34567, Details: details}

			found := entity.Photo{}
			res = tx.FirstOrCreate(&found, photo)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			photo2 := entity.Photo{ID: 34567}
			res = tx.Preload("Details").First(&photo2)
			if res.Error != nil {
				assert.Nil(t, res.Error)
				t.FailNow()
				return res.Error
			}

			// Nil indicates that gorm FirstOrCreate doesn't action included structs.
			assert.Nil(t, photo2.Details)
			if photo2.Details != nil {
				assert.Equal(t, details.Keywords, photo2.Details.Keywords)
			}

			return errors.New("ForceRollback")
		}))
	})

	// Use Entity Save with child structs to validate that values are set.
	// Fails on V2 with foreign key violation.
	t.Run("EntitySave_PhotoAndDetails", func(t *testing.T) {
		createsavetestMutex.Lock()
		defer createsavetestMutex.Unlock()

		stmt := entity.Db()
		labelotter := entity.Label{LabelName: "otter", LabelSlug: "otter"}
		var deletedAt = gorm.DeletedAt{Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
		labelsnake := entity.Label{LabelName: "snake", LabelSlug: "snake", DeletedAt: deletedAt}

		res := stmt.Save(&labelsnake)
		if res.Error != nil {
			assert.Nil(t, res.Error)
			t.FailNow()
		}

		res = stmt.Save(&labelotter)
		if res.Error != nil {
			assert.Nil(t, res.Error)
			t.FailNow()
		}

		details := &entity.Details{Keywords: "cow, flower, snake, otter"}
		photo := entity.Photo{ID: 34567, Details: details}

		log.Info("Expect 2 x foreign key violation Error or SQLSTATE from entity_save")
		err := photo.Save()
		assert.Nil(t, err)
		if err != nil {
			entity.UnscopedDb().Delete(&labelotter)
			entity.UnscopedDb().Delete(&labelsnake)
			t.FailNow()
		}

		photo2 := entity.Photo{ID: 34567}
		res = stmt.Preload("Details").First(&photo2)
		if res.Error != nil {
			assert.Nil(t, res.Error)
			t.FailNow()
		}

		assert.NotNil(t, photo2.Details)
		if photo2.Details != nil {
			assert.Equal(t, details.Keywords, photo2.Details.Keywords)
		}

		details2 := photo2.Details
		entity.UnscopedDb().Delete(&details2)
		entity.UnscopedDb().Delete(&photo2)
		entity.UnscopedDb().Delete(&labelotter)
		entity.UnscopedDb().Delete(&labelsnake)
	})

	// Use Entity Create with child structs to validate that values are set.
	// Works on V1 and V2 with Details being populated.
	t.Run("EntityCreate_PhotoAndDetails", func(t *testing.T) {
		createsavetestMutex.Lock()
		defer createsavetestMutex.Unlock()

		stmt := entity.Db()
		labelotter := entity.Label{LabelName: "otter", LabelSlug: "otter"}
		var deletedAt = gorm.DeletedAt{Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
		labelsnake := entity.Label{LabelName: "snake", LabelSlug: "snake", DeletedAt: deletedAt}

		res := stmt.Save(&labelsnake)
		if res.Error != nil {
			assert.Nil(t, res.Error)
			t.FailNow()
		}

		res = stmt.Save(&labelotter)
		if res.Error != nil {
			assert.Nil(t, res.Error)
			t.FailNow()
		}

		details := &entity.Details{Keywords: "cow, flower, snake, otter"}
		photo := entity.Photo{ID: 34567, Details: details}

		err := photo.Create()
		assert.Nil(t, err)
		if err != nil {
			entity.UnscopedDb().Delete(&labelotter)
			entity.UnscopedDb().Delete(&labelsnake)
			t.FailNow()
		}

		photo2 := entity.Photo{ID: 34567}
		res = stmt.Preload("Details").First(&photo2)
		if res.Error != nil {
			assert.Nil(t, res.Error)
			t.FailNow()
		}

		assert.NotNil(t, photo2.Details)
		if photo2.Details != nil {
			assert.Equal(t, details.Keywords, photo2.Details.Keywords)
		}

		details2 := photo2.Details
		entity.UnscopedDb().Delete(&details2)
		entity.UnscopedDb().Delete(&photo2)
		entity.UnscopedDb().Delete(&labelotter)
		entity.UnscopedDb().Delete(&labelsnake)

	})

	// Use entity Save without child structs to validate that values are set.
	// Works on V1 and V2.
	t.Run("EntitySave_Photo", func(t *testing.T) {
		createsavetestMutex.Lock()
		defer createsavetestMutex.Unlock()

		stmt := entity.Db()
		photo := entity.Photo{ID: 34567}

		err := photo.Save()
		assert.Nil(t, err)
		if err != nil {
			t.FailNow()
		}

		photo2 := entity.Photo{ID: 34567}
		assert.Empty(t, photo2.PhotoUID)
		res := stmt.Preload("Details").First(&photo2)
		if res.Error != nil {
			assert.Nil(t, res.Error)
			t.FailNow()
		}

		assert.NotNil(t, photo2.Details)
		assert.NotNil(t, photo2.PhotoUID)

		details2 := photo2.Details
		entity.UnscopedDb().Delete(&details2)
		entity.UnscopedDb().Delete(&photo2)
	})

}
