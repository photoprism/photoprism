package entity

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestNewLabel(t *testing.T) {
	t.Run("NameUnicornNum2000PriorityFive", func(t *testing.T) {
		label := NewLabel("Unicorn2000", 5)
		assert.Equal(t, "Unicorn2000", label.LabelName)
		assert.Equal(t, "unicorn2000", label.LabelSlug)
		assert.Equal(t, 5, label.LabelPriority)
	})
	t.Run("NameUnknown", func(t *testing.T) {
		label := NewLabel("", -6)
		assert.Equal(t, "Unknown", label.LabelName)
		assert.Equal(t, "unknown", label.LabelSlug)
		assert.Equal(t, -6, label.LabelPriority)
	})
}

func TestLabel_TableName(t *testing.T) {
	label := &Label{}
	assert.Equal(t, "labels", label.TableName())
}

func TestLabel_SaveForm(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		label := createTestLabel(t, "save-form")
		frm := &form.Label{
			LabelName:        "Sunrise Field",
			LabelPriority:    7,
			LabelFavorite:    true,
			LabelDescription: "desc",
			LabelNotes:       "notes",
			Thumb:            "thumb.jpg",
			ThumbSrc:         "manual",
		}

		require.NoError(t, label.SaveForm(frm))
		assert.Equal(t, "Sunrise Field", label.LabelName)
		assert.Equal(t, 7, label.LabelPriority)
		assert.True(t, label.LabelFavorite)
		assert.Equal(t, "desc", label.LabelDescription)
		assert.Equal(t, "notes", label.LabelNotes)
		assert.Equal(t, "thumb.jpg", label.Thumb)
		assert.Equal(t, "manual", label.ThumbSrc)
	})
	t.Run("InvalidForm", func(t *testing.T) {
		label := createTestLabel(t, "save-form-invalid")
		err := label.SaveForm(&form.Label{})
		assert.Error(t, err)
	})
}

func TestFlushLabelCache(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		FlushLabelCache()
	})
}

func TestLabel_SetName(t *testing.T) {
	t.Run("SetName", func(t *testing.T) {
		entity := LabelFixtures["landscape"]

		assert.Equal(t, "Landscape", entity.LabelName)
		assert.Equal(t, "landscape", entity.LabelSlug)
		assert.Equal(t, "landscape", entity.CustomSlug)

		entity.SetName("Landschaft")

		assert.Equal(t, "Landschaft", entity.LabelName)
		assert.Equal(t, "landscape", entity.LabelSlug)
		assert.Equal(t, "landschaft", entity.CustomSlug)
	})
	t.Run("NewNameEmpty", func(t *testing.T) {
		entity := LabelFixtures["flower"]

		assert.Equal(t, "Flower", entity.LabelName)
		assert.Equal(t, "flower", entity.LabelSlug)
		assert.Equal(t, "flower", entity.CustomSlug)

		assert.False(t, entity.SetName(""))

		assert.Equal(t, "Flower", entity.LabelName)
		assert.Equal(t, "flower", entity.LabelSlug)
		assert.Equal(t, "flower", entity.CustomSlug)
	})
}

func TestLabel_HasID(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var label *Label
		assert.False(t, label.HasID())
	})
	t.Run("Missing", func(t *testing.T) {
		label := &Label{ID: 1}
		assert.False(t, label.HasID())
	})
	t.Run("Persisted", func(t *testing.T) {
		label := createTestLabel(t, "has-id")
		assert.True(t, label.HasID())
	})
}

func TestLabel_HasUID(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var label *Label
		assert.False(t, label.HasUID())
	})
	t.Run("Invalid", func(t *testing.T) {
		label := &Label{LabelUID: "invalid"}
		assert.False(t, label.HasUID())
	})
	t.Run("Valid", func(t *testing.T) {
		uid := rnd.GenerateUID(LabelUID)
		label := &Label{LabelUID: uid}
		assert.True(t, label.HasUID())
	})
}

func TestLabel_Skip(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var label *Label
		assert.True(t, label.Skip())
	})
	t.Run("MissingID", func(t *testing.T) {
		label := &Label{}
		assert.True(t, label.Skip())
	})
	t.Run("Deleted", func(t *testing.T) {
		label := createTestLabel(t, "skip-deleted")
		now := time.Now()
		label.DeletedAt = &now
		assert.True(t, label.Skip())
	})
	t.Run("Active", func(t *testing.T) {
		label := createTestLabel(t, "skip-active")
		assert.False(t, label.Skip())
	})
}

func TestLabel_InvalidName(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		label := &Label{LabelName: ""}
		assert.True(t, label.InvalidName())
	})
	t.Run("Valid", func(t *testing.T) {
		label := &Label{LabelName: "Valid Name"}
		assert.False(t, label.InvalidName())
	})
}

func TestLabel_GetSlug(t *testing.T) {
	label := &Label{CustomSlug: "custom", LabelSlug: "orig", LabelName: "Name"}
	assert.Equal(t, "custom", label.GetSlug())

	label.CustomSlug = ""
	assert.Equal(t, "orig", label.GetSlug())

	label.LabelSlug = ""
	assert.Equal(t, "name", label.GetSlug())
}

func TestFirstOrCreateLabel(t *testing.T) {
	t.Run("Existing", func(t *testing.T) {
		label := LabelFixtures.Get("flower")
		result := FirstOrCreateLabel(&label)

		if result == nil {
			t.Fatal("result must not be nil")
		}

		if result.LabelName != label.LabelName {
			t.Errorf("LabelName should be the same: %s %s", result.LabelName, label.LabelName)
		}

		if result.LabelSlug != label.LabelSlug {
			t.Errorf("LabelName should be the same: %s %s", result.LabelSlug, label.LabelSlug)
		}
	})
	t.Run("MatchesExistingNameIgnoringCase", func(t *testing.T) {
		result := FirstOrCreateLabel(NewLabel("cow", 0))

		if result == nil {
			t.Fatal("result must not be nil")
		}

		assert.Equal(t, LabelFixtures.Get("cow").ID, result.ID)
		assert.Equal(t, "cow", result.LabelSlug)
	})
	t.Run("CreatesDistinctHomophoneLabels", func(t *testing.T) {
		first := FirstOrCreateLabel(NewLabel("问", 0))
		second := FirstOrCreateLabel(NewLabel("吻", 0))

		if first == nil || second == nil {
			t.Fatal("expected labels")
		}

		t.Cleanup(func() {
			_ = Db().Unscoped().Delete(second).Error
			_ = Db().Unscoped().Delete(first).Error
		})

		assert.Equal(t, "问", first.LabelName)
		assert.Equal(t, "吻", second.LabelName)
		assert.NotEqual(t, first.ID, second.ID)
		assert.Equal(t, "wen", first.LabelSlug)
		assert.Contains(t, second.LabelSlug, "wen-")
		assert.NotEqual(t, first.LabelSlug, second.LabelSlug)
		assert.Equal(t, first.LabelSlug, first.CustomSlug)
		assert.Equal(t, second.LabelSlug, second.CustomSlug)
	})
	t.Run("ReusesRenamedLabelByPreviousName", func(t *testing.T) {
		original := FirstOrCreateLabel(NewLabel("RenameCat", 0))
		require.NotNil(t, original)

		t.Cleanup(func() {
			_ = Db().Unscoped().Delete(original).Error
			FlushLabelCache()
		})

		require.True(t, original.SetName("RenameKatze"))
		require.NoError(t, Db().Save(original).Error)
		FlushLabelCache()

		assert.Equal(t, "renamecat", original.LabelSlug)
		assert.Equal(t, "renamekatze", original.CustomSlug)

		again := FirstOrCreateLabel(NewLabel("RenameCat", 0))
		require.NotNil(t, again)
		assert.Equal(t, original.ID, again.ID)
		assert.Equal(t, "RenameKatze", again.LabelName)

		var rows []Label
		require.NoError(t, UnscopedDb().
			Where("label_slug LIKE ? OR custom_slug LIKE ?", "renamecat%", "renamecat%").
			Find(&rows).Error)
		assert.Len(t, rows, 1, "rename + re-add must not create a duplicate")
	})
}

func TestLabel_UpdateClassify(t *testing.T) {
	t.Run("UpdatePriorityAndLabelSlug", func(t *testing.T) {
		classifyLabel := &classify.Label{Name: "classify", Uncertainty: 30, Source: "manual", Priority: 5}
		result := &Label{LabelName: "label", LabelSlug: "", CustomSlug: "customslug", LabelPriority: 4}

		assert.Equal(t, 4, result.LabelPriority)
		assert.Equal(t, "", result.LabelSlug)
		assert.Equal(t, "customslug", result.CustomSlug)
		assert.Equal(t, "label", result.LabelName)

		err := result.UpdateClassify(*classifyLabel)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 5, result.LabelPriority)
		assert.Equal(t, "customslug", result.LabelSlug)
		assert.Equal(t, "classify", result.CustomSlug)
		assert.Equal(t, "Classify", result.LabelName)
	})
	t.Run("UpdateCustomSlug", func(t *testing.T) {
		classifyLabel := &classify.Label{Name: "classify", Uncertainty: 30, Source: "manual", Priority: 5}
		result := &Label{LabelName: "label12", LabelSlug: "labelslug", CustomSlug: "", LabelPriority: 5}

		assert.Equal(t, 5, result.LabelPriority)
		assert.Equal(t, "labelslug", result.LabelSlug)
		assert.Equal(t, "", result.CustomSlug)
		assert.Equal(t, "label12", result.LabelName)

		err := result.UpdateClassify(*classifyLabel)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 5, result.LabelPriority)
		assert.Equal(t, "labelslug", result.LabelSlug)
		assert.Equal(t, "classify", result.CustomSlug)
		assert.Equal(t, "Classify", result.LabelName)

	})
	t.Run("UpdateNameAndCategories", func(t *testing.T) {
		classifyLabel := &classify.Label{Name: "classify", Uncertainty: 30, Source: "manual", Priority: 5, Categories: []string{"flower", "plant"}}
		result := &Label{LabelName: "label34", LabelSlug: "labelslug2", CustomSlug: "labelslug2", LabelPriority: 5, LabelCategories: []*Label{LabelFixtures.Pointer("flower")}}

		assert.Equal(t, 5, result.LabelPriority)
		assert.Equal(t, "labelslug2", result.LabelSlug)
		assert.Equal(t, "labelslug2", result.CustomSlug)
		assert.Equal(t, "label34", result.LabelName)

		err := result.UpdateClassify(*classifyLabel)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 5, result.LabelPriority)
		assert.Equal(t, "labelslug2", result.LabelSlug)
		assert.Equal(t, "classify", result.CustomSlug)
		assert.Equal(t, "Classify", result.LabelName)

	})
}

func TestLabel_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		label := createTestLabel(t, "update")
		oldPriority := label.LabelPriority
		require.NoError(t, label.Update("LabelPriority", oldPriority+5))
		require.NoError(t, Db().First(label, label.ID).Error)
		assert.Equal(t, oldPriority+5, label.LabelPriority)
	})
	t.Run("NilLabel", func(t *testing.T) {
		var label *Label
		err := label.Update("LabelPriority", 1)
		assert.EqualError(t, err, "label must not be nil - you may have found a bug")
	})
	t.Run("MissingID", func(t *testing.T) {
		label := NewLabel("missing", 0)
		err := label.Update("LabelPriority", 1)
		assert.EqualError(t, err, "label ID must not be empty - you may have found a bug")
	})
}

func TestLabel_SaveForm_CollidingSlug(t *testing.T) {
	base := FirstOrCreateLabel(NewLabel("问", 0))
	other := createTestLabel(t, "save-form-collision")

	if base == nil {
		t.Fatal("expected base label")
	}

	t.Cleanup(func() {
		_ = Db().Unscoped().Delete(base).Error
	})

	originalSlug := other.LabelSlug
	frm := &form.Label{LabelName: "吻"}
	require.NoError(t, other.SaveForm(frm))

	assert.Equal(t, "吻", other.LabelName)
	assert.Equal(t, originalSlug, other.LabelSlug)
	assert.NotEqual(t, "wen", other.CustomSlug)
	assert.Contains(t, other.CustomSlug, "wen-")
}

func TestLabel_Updates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		label := createTestLabel(t, "updates")
		err := label.Updates(&Label{LabelDescription: "updated", LabelNotes: "notes"})
		require.NoError(t, err)
		require.NoError(t, Db().First(label, label.ID).Error)
		assert.Equal(t, "updated", label.LabelDescription)
		assert.Equal(t, "notes", label.LabelNotes)
	})
	t.Run("NilValues", func(t *testing.T) {
		label := createTestLabel(t, "updates-nil")
		assert.NoError(t, label.Updates(nil))
	})
	t.Run("NilLabel", func(t *testing.T) {
		var label *Label
		err := label.Updates(&Label{LabelDescription: "x"})
		assert.EqualError(t, err, "label must not be nil - you may have found a bug")
	})
	t.Run("MissingID", func(t *testing.T) {
		label := NewLabel("missing", 0)
		err := label.Updates(&Label{LabelDescription: "x"})
		assert.EqualError(t, err, "label ID must not be empty - you may have found a bug")
	})
}

func TestLabel_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		label := NewLabel("Unicorn2000", 5)
		initialDate := label.UpdatedAt
		err := label.Save()

		if err != nil {
			t.Fatal(err)
		}

		afterDate := label.UpdatedAt

		assert.True(t, afterDate.After(initialDate))

	})
}

func TestLabel_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		label := NewLabel("LabelToBeDeleted", 5)
		err := label.Save()
		if err != nil {
			t.Fatal(err)
		}
		assert.False(t, label.Deleted())

		var labels Labels

		if err = Db().Where("label_name = ?", label.LabelName).Find(&labels).Error; err != nil {
			t.Fatal(err)
		}

		assert.Len(t, labels, 1)

		err = label.Delete()
		if err != nil {
			t.Fatal(err)
		}

		if err = Db().Where("label_name = ?", label.LabelName).Find(&labels).Error; err != nil {
			t.Fatal(err)
		}

		assert.Len(t, labels, 0)
	})
}

func TestLabel_Restore(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var deletedAt = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		label := &Label{DeletedAt: &deletedAt, LabelName: "ToBeRestored"}

		if err := label.Save(); err != nil {
			t.Fatal(err)
		}

		assert.True(t, label.Deleted())

		if err := label.Restore(); err != nil {
			t.Fatal(err)
		}

		assert.False(t, label.Deleted())
	})
	t.Run("LabelNotDeleted", func(t *testing.T) {
		label := &Label{DeletedAt: nil, LabelName: "NotDeleted1234"}

		if err := label.Restore(); err != nil {
			t.Fatal(err)
		}

		assert.False(t, label.Deleted())
	})
}

func TestLabel_Links(t *testing.T) {
	t.Run("OneResult", func(t *testing.T) {
		label := LabelFixtures.Get("flower")
		links := label.Links()
		assert.Equal(t, "6jxf3jfn2k", links[0].LinkToken)
	})
}

func createTestLabel(t *testing.T, prefix string) *Label {
	t.Helper()
	name := fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	label := NewLabel(name, 0)
	require.NoError(t, label.Save())

	t.Cleanup(func() {
		_ = Db().Unscoped().Delete(label).Error
	})

	return label
}
