package entity

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// openTestDb opens a new SQLite database in a temporary directory and closes it when the test ends.
func openTestDb(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = CloseDb(db) })

	return db
}

func TestGormConfig(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := GormConfig()
		assert.Equal(t, logger.Discard, c.Logger)
		assert.True(t, c.DisableForeignKeyConstraintWhenMigrating)
		assert.True(t, c.IgnoreRelationshipsWhenMigrating)
		assert.Equal(t, 0, c.NowFunc().Nanosecond())
		assert.Equal(t, time.UTC, c.NowFunc().Location())
	})
}

func TestGormDialector(t *testing.T) {
	t.Run("MySQL", func(t *testing.T) {
		d, err := GormDialector(dsn.DriverMySQL, "user:pass@tcp(localhost:3306)/db", nil)
		require.NoError(t, err)
		assert.Equal(t, "mysql", d.Name())
		m, ok := d.(mysqlDialector)
		require.True(t, ok)
		assert.Equal(t, uint(255), m.DefaultStringSize)
		assert.True(t, m.DisableDatetimePrecision)
	})
	t.Run("SQLite", func(t *testing.T) {
		d, err := GormDialector(dsn.DriverSQLite3, "test.db", nil)
		require.NoError(t, err)
		assert.Equal(t, "sqlite", d.Name())
	})
	t.Run("PostgreSQL", func(t *testing.T) {
		d, err := GormDialector("postgresql", "postgres://u:p@localhost/db", nil)
		require.NoError(t, err)
		assert.Equal(t, "postgres", d.Name())
		p, ok := d.(postgresDialector)
		require.True(t, ok)
		assert.Equal(t, "postgres://u:p@localhost/db", p.DSN)
		assert.Nil(t, p.Conn)
	})
	t.Run("Unsupported", func(t *testing.T) {
		d, err := GormDialector("oracle", "host=localhost", nil)
		assert.Error(t, err)
		assert.Nil(t, d)
		d, err = GormDialector("", "", nil)
		assert.Error(t, err)
		assert.Nil(t, d)
	})
}

func TestOpenDb(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db := openTestDb(t)
		assert.Equal(t, "sqlite", db.Name())
		require.NoError(t, db.Exec("SELECT 1").Error)
	})
	t.Run("Unsupported", func(t *testing.T) {
		db, err := OpenDb("oracle", "host=localhost")
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

func TestOpenDbConn(t *testing.T) {
	t.Run("SharedPool", func(t *testing.T) {
		db := openTestDb(t)
		pool, err := db.DB()
		require.NoError(t, err)

		conn, err := OpenDbConn(dsn.DriverSQLite3, "", pool)
		require.NoError(t, err)
		require.NoError(t, conn.Exec("CREATE TABLE shared_pool (id INTEGER)").Error)
		assert.True(t, db.Migrator().HasTable("shared_pool"))
	})
	t.Run("Unsupported", func(t *testing.T) {
		db, err := OpenDbConn("oracle", "", nil)
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

func TestCloseDb(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.NoError(t, CloseDb(nil))
	})
	t.Run("Success", func(t *testing.T) {
		db, err := OpenDb(dsn.DriverSQLite3, filepath.Join(t.TempDir(), "close.db"))
		require.NoError(t, err)
		require.NoError(t, CloseDb(db))
		pool, err := db.DB()
		require.NoError(t, err)
		assert.Error(t, pool.Ping())
	})
}

func TestDisableAssociationSaves(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Error(t, DisableAssociationSaves(nil))
	})
	t.Run("Callbacks", func(t *testing.T) {
		db := openTestDb(t)

		for _, name := range associationCallbacks {
			assert.Nil(t, db.Callback().Create().Get(name), name)
			assert.Nil(t, db.Callback().Update().Get(name), name)
		}

		assert.NotNil(t, db.Callback().Update().Get("photoprism:omit_blank_fields"))
		assert.NotNil(t, db.Callback().Update().Get("photoprism:save_fallback"))
	})
	t.Run("RelatedRowsNotSaved", func(t *testing.T) {
		photo := Photo{PhotoUID: rnd.GenerateUID(PhotoUID), PhotoName: "Associations", OriginalName: "associations"}
		photo.Details = &Details{Keywords: "associations"}
		photo.Labels = []PhotoLabel{{LabelID: LabelFixtures.Get("flower").ID, LabelSrc: SrcManual}}
		require.NoError(t, Db().Create(&photo).Error)
		t.Cleanup(func() { _ = UnscopedDb().Delete(&photo).Error })

		var count int64
		require.NoError(t, UnscopedDb().Model(&Details{}).Where("photo_id = ?", photo.ID).Count(&count).Error)
		assert.Zero(t, count)
		require.NoError(t, UnscopedDb().Model(&PhotoLabel{}).Where("photo_id = ?", photo.ID).Count(&count).Error)
		assert.Zero(t, count)
	})
	t.Run("AssociationAppendNotSaved", func(t *testing.T) {
		label := LabelFixtures.Get("cake")
		category := LabelFixtures.Get("cow")
		cond := "label_id = ? AND category_id = ?"
		t.Cleanup(func() { _ = Db().Where(cond, label.ID, category.ID).Delete(&Category{}).Error })

		// The Association API saves through the removed callbacks, so Append reports no error and writes no row.
		require.NoError(t, Db().Model(&label).Association("LabelCategories").Append(&category))

		var count int64
		require.NoError(t, Db().Model(&Category{}).Where(cond, label.ID, category.ID).Count(&count).Error)
		assert.Zero(t, count)
	})
	t.Run("AssociationReplaceRemovesLinks", func(t *testing.T) {
		label := NewLabel("Association Replace", 0)
		linked := NewLabel("Association Linked", 0)
		replacement := NewLabel("Association Replacement", 0)

		for _, m := range []*Label{label, linked, replacement} {
			require.NoError(t, m.Create())
			t.Cleanup(func() { _ = UnscopedDb().Delete(m).Error })
		}

		t.Cleanup(func() { _ = Db().Where("label_id = ?", label.ID).Delete(&Category{}).Error })
		require.NoError(t, Db().Create(&Category{LabelID: label.ID, CategoryID: linked.ID}).Error)

		// Replace deletes the links it does not keep, but saves the new ones through the removed callbacks.
		require.NoError(t, Db().Model(label).Association("LabelCategories").Replace(replacement))

		var count int64
		require.NoError(t, Db().Model(&Category{}).Where("label_id = ?", label.ID).Count(&count).Error)
		assert.Zero(t, count)
	})
}

func TestIsSave(t *testing.T) {
	t.Run("Save", func(t *testing.T) {
		stmt := Db().Session(&gorm.Session{DryRun: true}).Save(&Album{ID: 1, AlbumUID: rnd.GenerateUID(AlbumUID)}).Statement
		assert.True(t, isSave(stmt))
	})
	t.Run("Updates", func(t *testing.T) {
		stmt := Db().Session(&gorm.Session{DryRun: true}).Model(&Album{ID: 1}).Updates(Values{"album_title": "Title"}).Statement
		assert.False(t, isSave(stmt))
	})
}

func TestOmitBlankFields(t *testing.T) {
	t.Run("KeepsCreatedAt", func(t *testing.T) {
		createdAt := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
		album := Album{AlbumTitle: "Omit Blank Fields", AlbumSlug: "omit-blank-" + rnd.Base36(6), AlbumType: AlbumManual, CreatedAt: createdAt}
		require.NoError(t, Db().Create(&album).Error)
		t.Cleanup(func() { _ = UnscopedDb().Delete(&album).Error })

		saved := Album{ID: album.ID, AlbumUID: album.AlbumUID, AlbumTitle: "Omit Blank Fields Saved", AlbumSlug: album.AlbumSlug, AlbumType: AlbumManual}
		require.NoError(t, Db().Save(&saved).Error)

		found := Album{}
		require.NoError(t, UnscopedDb().Where("id = ?", album.ID).First(&found).Error)
		assert.Equal(t, "Omit Blank Fields Saved", found.AlbumTitle)
		assert.True(t, createdAt.Equal(found.CreatedAt), "created_at kept: %s", found.CreatedAt)
	})
	t.Run("KeepsStoredReference", func(t *testing.T) {
		photo := Photo{PhotoUID: rnd.GenerateUID(PhotoUID), PhotoName: "References", OriginalName: "references", CameraID: UnknownCamera.ID}
		require.NoError(t, photo.Create())
		t.Cleanup(func() { _ = UnscopedDb().Delete(&photo).Error })

		cellID := photo.CellID
		photo.CellID = ""
		photo.Cell = nil
		require.NoError(t, Db().Save(&photo).Error)

		found := Photo{}
		require.NoError(t, UnscopedDb().Where("id = ?", photo.ID).First(&found).Error)
		assert.Equal(t, cellID, found.CellID)
	})
}

func TestSaveFallback(t *testing.T) {
	t.Run("CreatesMissingRow", func(t *testing.T) {
		// The album has an ID but no row, so the update matches nothing and the row is created with its hooks.
		album := Album{ID: 9876543, AlbumTitle: "Save Fallback", AlbumSlug: "save-fallback-" + rnd.Base36(6), AlbumType: AlbumManual}
		t.Cleanup(func() { _ = UnscopedDb().Delete(&Album{}, "id = ?", album.ID).Error })
		require.NoError(t, Db().Save(&album).Error)

		found := Album{}
		require.NoError(t, UnscopedDb().Where("id = ?", album.ID).First(&found).Error)
		assert.True(t, rnd.IsUID(found.AlbumUID, AlbumUID), "uid set by hook: %s", found.AlbumUID)
	})
	t.Run("UnchangedRow", func(t *testing.T) {
		// MariaDB reports no affected row when an update changes nothing, which the ORM would follow with an insert.
		db := openTestDb(t)
		require.NoError(t, db.AutoMigrate(&Album{}))
		require.NoError(t, db.Callback().Update().After("gorm:update").Before("photoprism:save_fallback").
			Register("test:unchanged", func(tx *gorm.DB) { tx.RowsAffected = 0 }))

		album := Album{AlbumTitle: "Save Fallback Unchanged", AlbumSlug: "save-fallback-unchanged", AlbumType: AlbumManual}
		require.NoError(t, db.Create(&album).Error)

		counter := &statementCounter{}
		db.Logger = counter
		res := db.Save(&album)
		require.NoError(t, res.Error)
		assert.Equal(t, int64(1), res.RowsAffected)

		for _, stmt := range counter.sql {
			assert.False(t, strings.HasPrefix(stmt, "INSERT"), stmt)
		}

		var count int64
		require.NoError(t, db.Model(&Album{}).Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})
	t.Run("NoSave", func(t *testing.T) {
		// Updates of a missing row do not create it.
		res := Db().Model(&Album{ID: 9876544}).Updates(Values{"album_title": "Missing"})
		require.NoError(t, res.Error)
		assert.Zero(t, res.RowsAffected)

		var count int64
		require.NoError(t, UnscopedDb().Model(&Album{}).Where("id = ?", 9876544).Count(&count).Error)
		assert.Zero(t, count)
	})
}

// dryRunTestDb returns a SQLite connection that builds statements without running them, so that their SQL
// does not depend on the database the tests run on.
func dryRunTestDb(t *testing.T) func() *gorm.DB {
	t.Helper()

	db := openTestDb(t)
	return func() *gorm.DB { return db.Session(&gorm.Session{DryRun: true}) }
}

func TestBuildWhere(t *testing.T) {
	dryRun := dryRunTestDb(t)

	t.Run("Conditions", func(t *testing.T) {
		stmt := dryRun().Model(&Photo{}).Where("photo_title = ?\n\tOR photo_title = ?", "a", "b").Where("photo_quality > ?", 2).Find(&Photos{}).Statement
		assert.Contains(t, stmt.SQL.String(), "WHERE (photo_title = ?\n\tOR photo_title = ?) AND (photo_quality > ?) AND `photos`.`deleted_at` IS NULL")
	})
	t.Run("Expression", func(t *testing.T) {
		stmt := dryRun().Model(&Label{}).Where("?", gorm.Expr("label_name = ? OR label_slug = ?", "a", "b")).Where("label_priority >= ?", 0).Find(&Labels{}).Statement
		assert.Contains(t, stmt.SQL.String(), "WHERE ( label_name = ? OR label_slug = ?) AND (label_priority >= ?) AND `labels`.`deleted_at` IS NULL")
	})
	t.Run("Single", func(t *testing.T) {
		stmt := dryRun().Unscoped().Model(&Photo{}).Where("photo_title = ?", "a").Find(&Photos{}).Statement
		assert.Contains(t, stmt.SQL.String(), "WHERE photo_title = ?")
	})
	t.Run("Or", func(t *testing.T) {
		stmt := dryRun().Unscoped().Model(&Photo{}).Where("photo_title = ?", "a").Or("photo_title = ?", "b").Find(&Photos{}).Statement
		assert.Contains(t, stmt.SQL.String(), "WHERE (photo_title = ?) OR photo_title = ?")
	})
	t.Run("LeadingSlice", func(t *testing.T) {
		stmt := dryRun().Unscoped().Model(&Photo{}).Where("? IN (photo_title, photo_name)", []string{"a"}).Where("photo_quality > ?", 2).Find(&Photos{}).Statement
		assert.Contains(t, stmt.SQL.String(), "WHERE ( (?) IN (photo_title, photo_name)) AND (photo_quality > ?)")
	})
	t.Run("Update", func(t *testing.T) {
		stmt := dryRun().Model(&Photo{}).Where("photo_title = ?\nOR photo_title = ?", "a", "b").Where("photo_quality > ?", 2).UpdateColumn("photo_private", true).Statement
		assert.Contains(t, stmt.SQL.String(), "WHERE (photo_title = ?\nOR photo_title = ?) AND (photo_quality > ?)")
	})
}

func TestBindPrimaryKeys(t *testing.T) {
	dryRun := dryRunTestDb(t)

	t.Run("Update", func(t *testing.T) {
		stmt := dryRun().Model(&PhotoLabel{PhotoID: 1}).UpdateColumn("uncertainty", 10).Statement
		assert.Contains(t, stmt.SQL.String(), "`photos_labels`.`label_id` = ?")
		assert.Contains(t, stmt.SQL.String(), "`photo_id` = ?")
	})
	t.Run("Query", func(t *testing.T) {
		stmt := dryRun().First(&PhotoLabel{PhotoID: 1}).Statement
		assert.Contains(t, stmt.SQL.String(), "`photos_labels`.`label_id` = ?")
		assert.Contains(t, stmt.SQL.String(), "`photo_id` = ?")
	})
	t.Run("FirstZero", func(t *testing.T) {
		// Like before, a zero first key part binds only the parts that are set.
		stmt := dryRun().First(&PhotoLabel{LabelID: 2}).Statement
		assert.NotContains(t, stmt.SQL.String(), "`photo_id` = ?")
		assert.Contains(t, stmt.SQL.String(), "`label_id` = ?")
	})
	t.Run("AllSet", func(t *testing.T) {
		stmt := dryRun().Model(&PhotoLabel{PhotoID: 1, LabelID: 2}).UpdateColumn("uncertainty", 10).Statement
		assert.Equal(t, 1, strings.Count(stmt.SQL.String(), "label_id"))
	})
	t.Run("NoneSet", func(t *testing.T) {
		err := Db().Model(&PhotoLabel{}).UpdateColumn("uncertainty", 10).Error
		assert.ErrorIs(t, err, gorm.ErrMissingWhereClause)
	})
	t.Run("Rows", func(t *testing.T) {
		// Only the row with the zero key part is updated, not every row with the same photo.
		photoID := uint(9876501)
		labels := []PhotoLabel{{PhotoID: photoID, LabelID: 0, LabelSrc: SrcManual}, {PhotoID: photoID, LabelID: 1, LabelSrc: SrcManual}}
		require.NoError(t, UnscopedDb().Create(&labels).Error)
		t.Cleanup(func() { _ = UnscopedDb().Where("photo_id = ?", photoID).Delete(&PhotoLabel{}).Error })

		require.NoError(t, UnscopedDb().Model(&PhotoLabel{PhotoID: photoID}).UpdateColumn("uncertainty", 42).Error)

		var updated int64
		require.NoError(t, UnscopedDb().Model(&PhotoLabel{}).Where("photo_id = ? AND uncertainty = 42", photoID).Count(&updated).Error)
		assert.Equal(t, int64(1), updated)
	})
}
