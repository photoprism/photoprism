package entity

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// renameTestLabel stores a label with independent identity and metadata.
func renameTestLabel(t *testing.T, name string) *Label {
	t.Helper()
	label := NewLabel(name, 3)
	label.LabelDescription, label.LabelNotes = "Description control", "Notes control"
	require.NoError(t, label.Create())
	t.Cleanup(func() { UnscopedDb().Delete(&Label{}, "id = ?", label.ID) })
	var stored Label
	require.NoError(t, Db().First(&stored, label.ID).Error)
	return &stored
}

// TestLabel_Rename checks display-name updates against the global label naming rules.
func TestLabel_Rename(t *testing.T) {
	t.Run("PreserveProperties", func(t *testing.T) {
		label := renameTestLabel(t, "Original "+rnd.GenerateUID('l'))
		before := *label
		require.NoError(t, label.Rename("Renamed Label Control"))
		var stored Label
		require.NoError(t, Db().First(&stored, label.ID).Error)
		assert.Equal(t, "Renamed Label Control", stored.LabelName)
		assert.Equal(t, "renamed-label-control", stored.CustomSlug)
		before.LabelName, before.CustomSlug = stored.LabelName, stored.CustomSlug
		assert.Equal(t, before, stored)
		require.NoError(t, label.Rename("Renamed Label Control"))
	})
	t.Run("ExistingNameKeepsRows", func(t *testing.T) {
		target := renameTestLabel(t, "Existing "+rnd.GenerateUID('l'))
		source := renameTestLabel(t, "Source "+rnd.GenerateUID('l'))
		control := renameTestLabel(t, "Control "+rnd.GenerateUID('l'))
		uid, slug := source.LabelUID, source.LabelSlug
		frm, err := form.NewLabel(control)
		require.NoError(t, err)
		frm.LabelName = target.LabelName
		require.NoError(t, control.SaveForm(frm))
		require.NoError(t, source.Rename(target.LabelName))
		assert.Equal(t, control.LabelName, source.LabelName)
		assert.Equal(t, control.CustomSlug, source.CustomSlug)
		assert.Equal(t, uid, source.LabelUID)
		assert.Equal(t, slug, source.LabelSlug)
		assert.NotEqual(t, target.ID, source.ID)
		var n int
		require.NoError(t, Db().Model(&Label{}).Where("id IN (?)", []uint{target.ID, source.ID, control.ID}).Count(&n).Error)
		assert.Equal(t, 3, n)
	})
	t.Run("DistinctNamesShareSlugBase", func(t *testing.T) {
		suffix := rnd.GenerateUID('l')
		target := renameTestLabel(t, "问 "+suffix)
		source := renameTestLabel(t, "Source "+suffix)
		originalSlug := source.LabelSlug
		require.NoError(t, source.Rename("吻 "+suffix))
		assert.NotEqual(t, target.CustomSlug, source.CustomSlug)
		assert.Equal(t, originalSlug, source.LabelSlug)
	})
	t.Run("MissingRow", func(t *testing.T) {
		label := renameTestLabel(t, "Missing "+rnd.GenerateUID('l'))
		before := *label
		require.NoError(t, UnscopedDb().Delete(label).Error)
		require.ErrorIs(t, label.Rename("Missing Rename"), gorm.ErrRecordNotFound)
		assert.Equal(t, before, *label)
	})
	t.Run("EmptyCanonicalSlug", func(t *testing.T) {
		label := renameTestLabel(t, "Empty "+rnd.GenerateUID('l'))
		require.NoError(t, UnscopedDb().Model(label).UpdateColumn("label_slug", "").Error)
		label.LabelSlug = ""
		require.NoError(t, label.Rename("Empty Slug Control"))
		var stored Label
		require.NoError(t, Db().First(&stored, label.ID).Error)
		assert.Equal(t, "empty-slug-control", stored.LabelSlug)
		assert.Equal(t, stored.LabelSlug, label.LabelSlug)
	})
	t.Run("WriteFailure", func(t *testing.T) {
		label := renameTestLabel(t, "Failed "+rnd.GenerateUID('l'))
		before := *label
		Db().Callback().Update().Before("gorm:begin_transaction").Register("test:rename-failure", func(scope *gorm.Scope) {
			if scope.TableName() == (Label{}).TableName() {
				_ = scope.Err(errors.New("rename write control"))
			}
		})
		t.Cleanup(func() { Db().Callback().Update().Remove("test:rename-failure") })
		require.ErrorContains(t, label.Rename("Failed Rename"), "rename write control")
		assert.Equal(t, before, *label)
		var stored Label
		require.NoError(t, Db().First(&stored, label.ID).Error)
		assert.Equal(t, before, stored)
	})
	t.Run("Invalid", func(t *testing.T) {
		label := renameTestLabel(t, "Invalid "+rnd.GenerateUID('l'))
		before := *label
		var missing *Label
		require.ErrorIs(t, missing.Rename("Name"), ErrInvalidName)
		require.ErrorIs(t, (&Label{}).Rename("Name"), ErrInvalidName)
		statements := countStatements(t, func() { require.ErrorIs(t, label.Rename(""), ErrInvalidName) })
		assert.Empty(t, statements)
		assert.Equal(t, before, *label)
	})
}

// TestPhoto_SaveLabels_AssignmentWrites checks that metadata maintenance keeps saved label assignments read-only.
func TestPhoto_SaveLabels_AssignmentWrites(t *testing.T) {
	WaitForAsyncJobs()
	photo := NewPhoto(false)
	require.NoError(t, photo.Save())
	label := renameTestLabel(t, "Sunflower "+rnd.GenerateUID('l'))
	link := NewPhotoLabel(photo.ID, label.ID, 0, SrcManual)
	require.NoError(t, link.Create())
	t.Cleanup(func() {
		WaitForAsyncJobs()
		UnscopedDb().Delete(&PhotoLabel{}, "photo_id = ?", photo.ID)
		UnscopedDb().Delete(&Details{}, "photo_id = ?", photo.ID)
		UnscopedDb().Delete(&photo)
	})
	photo.PreloadLabels()
	require.Len(t, photo.Labels, 1)
	var writes atomic.Int64
	Db().Callback().Update().Before("gorm:begin_transaction").Register("test:metadata-label-writes", func(scope *gorm.Scope) {
		switch scope.TableName() {
		case (PhotoLabel{}).TableName():
			writes.Add(1)
		case (Label{}).TableName():
			if attrs, ok := scope.InstanceGet("gorm:update_attrs"); ok {
				values := attrs.(map[string]interface{})
				if len(values) == 1 {
					if _, counts := values["photo_count"]; counts {
						return
					}
				}
			}
			writes.Add(1)
		}
	})
	t.Cleanup(func() { Db().Callback().Update().Remove("test:metadata-label-writes") })
	photo.PhotoTitle, photo.TitleSrc = "Metadata Title Control", SrcManual
	require.NoError(t, photo.SaveLabels())
	WaitForAsyncJobs()
	assert.Zero(t, writes.Load())
	require.Len(t, photo.Labels, 1)
	assert.Equal(t, label.ID, photo.Labels[0].LabelID)
	assert.Equal(t, "Metadata Title Control", photo.PhotoTitle)
	stored := FindPhoto(Photo{ID: photo.ID})
	require.NotNil(t, stored)
	assert.Equal(t, photo.PhotoTitle, stored.PhotoTitle)
}

// TestPhoto_SaveLabels_RestoreOnError keeps loaded assignments available after a metadata write failure.
func TestPhoto_SaveLabels_RestoreOnError(t *testing.T) {
	WaitForAsyncJobs()
	photo := NewPhoto(false)
	require.NoError(t, photo.Save())
	label := renameTestLabel(t, "Restore "+rnd.GenerateUID('l'))
	require.NoError(t, NewPhotoLabel(photo.ID, label.ID, 0, SrcManual).Create())
	t.Cleanup(func() {
		WaitForAsyncJobs()
		UnscopedDb().Delete(&PhotoLabel{}, "photo_id = ?", photo.ID)
		UnscopedDb().Delete(&Details{}, "photo_id = ?", photo.ID)
		UnscopedDb().Delete(&photo)
	})
	photo.PreloadLabels()
	before := photo.Labels
	Db().Callback().Update().Before("gorm:begin_transaction").Register("test:metadata-save-error", func(scope *gorm.Scope) {
		if scope.TableName() == (Photo{}).TableName() {
			_ = scope.Err(errors.New("metadata write control"))
		}
	})
	t.Cleanup(func() { Db().Callback().Update().Remove("test:metadata-save-error") })
	require.ErrorContains(t, photo.SaveLabels(), "metadata write control")
	assert.Equal(t, before, photo.Labels)
}
