package entity

import (
	"errors"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// failDetailsUpdates makes updates of photo details fail with err as long as fail returns true.
func failDetailsUpdates(t *testing.T, err error, fail func() bool) {
	t.Helper()

	name := "test:details-update-failure"

	Db().Callback().Update().Before("gorm:begin_transaction").Register(name, func(scope *gorm.Scope) {
		if scope.TableName() == (Details{}).TableName() && fail() {
			_ = scope.Err(err)
		}
	})
	t.Cleanup(func() { Db().Callback().Update().Remove(name) })
}

// storedKeywords returns the keywords stored for a photo.
func storedKeywords(t *testing.T, photoID uint) string {
	t.Helper()

	var found Details
	require.NoError(t, UnscopedDb().Where("photo_id = ?", photoID).First(&found).Error)

	return found.Keywords
}

func TestPhoto_SaveDetails(t *testing.T) {
	// newDetailsPhoto creates a picture with stored details and removes it afterwards.
	newDetailsPhoto := func(t *testing.T) *Photo {
		m := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID), Details: &Details{Keywords: "stored"}}
		createShapePhoto(t, m)
		require.Equal(t, "stored", storedKeywords(t, m.ID))
		return m
	}

	t.Run("Success", func(t *testing.T) {
		m := newDetailsPhoto(t)
		m.Details.Keywords = "changed"

		require.NoError(t, m.SaveDetails())
		assert.Equal(t, "changed", storedKeywords(t, m.ID))
	})
	t.Run("Locked", func(t *testing.T) {
		m := newDetailsPhoto(t)
		m.Details.Keywords = "changed"

		attempts := 0
		failDetailsUpdates(t, sqlite3.Error{Code: sqlite3.ErrBusy}, func() bool { attempts++; return attempts == 1 })

		require.NoError(t, m.SaveDetails())
		assert.Equal(t, 2, attempts)
		assert.Equal(t, "changed", storedKeywords(t, m.ID))
	})
	t.Run("StillLocked", func(t *testing.T) {
		m := newDetailsPhoto(t)
		m.Details.Keywords = "changed"

		attempts := 0
		failDetailsUpdates(t, sqlite3.Error{Code: sqlite3.ErrBusy}, func() bool { attempts++; return true })

		assert.EqualError(t, m.SaveDetails(), "database is locked")
		assert.Equal(t, 2, attempts)
		assert.Equal(t, "stored", storedKeywords(t, m.ID))
	})
	t.Run("Failed", func(t *testing.T) {
		m := newDetailsPhoto(t)
		m.Details.Keywords = "changed"
		m.Details.CreatedAt = time.Time{}

		failDetailsUpdates(t, errors.New("write control"), func() bool { return true })

		assert.EqualError(t, m.SaveDetails(), "write control")
		assert.Equal(t, "stored", storedKeywords(t, m.ID))
		assert.Equal(t, "changed", m.Details.Keywords, "changes are kept in memory")
		assert.True(t, m.Details.CreatedAt.IsZero(), "created at is not set by the failed save")
	})
	t.Run("PhotoSaveFailed", func(t *testing.T) {
		m := newDetailsPhoto(t)
		m.Details.Keywords = "changed"

		failDetailsUpdates(t, errors.New("write control"), func() bool { return true })

		assert.EqualError(t, m.Save(), "write control")
		assert.Equal(t, "stored", storedKeywords(t, m.ID))
	})
	t.Run("NotLoaded", func(t *testing.T) {
		// Details that were not loaded have no changes, so failing updates do not matter.
		m := &Photo{PhotoUID: rnd.GenerateUID(PhotoUID)}
		failDetailsUpdates(t, errors.New("write control"), func() bool { return true })
		createShapePhoto(t, m)

		require.NotNil(t, m.Details)
		assert.Equal(t, m.ID, m.Details.PhotoID)
		assert.Equal(t, "", storedKeywords(t, m.ID))
	})
	t.Run("MissingRow", func(t *testing.T) {
		m := newDetailsPhoto(t)
		require.NoError(t, UnscopedDb().Where("photo_id = ?", m.ID).Delete(&Details{}).Error)
		m.Details.Keywords = "changed"

		// Creating the missing row stores the changes.
		failDetailsUpdates(t, errors.New("write control"), func() bool { return true })

		require.NoError(t, m.SaveDetails())
		assert.Equal(t, "changed", storedKeywords(t, m.ID))
	})
}
