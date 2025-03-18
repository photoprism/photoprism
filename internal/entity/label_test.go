package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/classify"
)

func TestNewLabel(t *testing.T) {
	t.Run("name Unicorn2000 priority 5", func(t *testing.T) {
		label := NewLabel("Unicorn2000", 5)
		assert.Equal(t, "Unicorn2000", label.LabelName)
		assert.Equal(t, "unicorn2000", label.LabelSlug)
		assert.Equal(t, 5, label.LabelPriority)
	})
	t.Run("name Unknown", func(t *testing.T) {
		label := NewLabel("", -6)
		assert.Equal(t, "Unknown", label.LabelName)
		assert.Equal(t, "unknown", label.LabelSlug)
		assert.Equal(t, -6, label.LabelPriority)
	})
}

func TestFlushLabelCache(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		FlushLabelCache()
	})
}

func TestLabel_SetName(t *testing.T) {
	t.Run("set name", func(t *testing.T) {
		entity := LabelFixtures["landscape"]

		assert.Equal(t, "Landscape", entity.LabelName)
		assert.Equal(t, "landscape", entity.LabelSlug)
		assert.Equal(t, "landscape", entity.CustomSlug)

		entity.SetName("Landschaft")

		assert.Equal(t, "Landschaft", entity.LabelName)
		assert.Equal(t, "landscape", entity.LabelSlug)
		assert.Equal(t, "landschaft", entity.CustomSlug)
	})
	t.Run("new name empty", func(t *testing.T) {
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

func TestFirstOrCreateLabel(t *testing.T) {
	label := LabelFixtures.Get("flower")
	result := FirstOrCreateLabel(&label)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.LabelName != label.LabelName {
		t.Errorf("LabelName should be the same: %s %s", result.LabelName, label.LabelName)
	}

	if result.LabelSlug != label.LabelSlug {
		t.Errorf("LabelName should be the same: %s %s", result.LabelSlug, label.LabelSlug)
	}
}

func TestLabel_UpdateClassify(t *testing.T) {
	t.Run("update priority and label slug", func(t *testing.T) {
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
	t.Run("update custom slug", func(t *testing.T) {
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
	t.Run("update name and Categories", func(t *testing.T) {
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
	t.Run("label not deleted", func(t *testing.T) {
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

func TestLabel_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		label := &Label{LabelSlug: "to-be-updated", LabelName: "Update Me Please"}

		err := label.Save()

		if err != nil {
			t.Fatal(err)
		}

		err = label.Update("LabelSlug", "my-unique-slug")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "my-unique-slug", label.LabelSlug)
	})
}
