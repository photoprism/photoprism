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
	"github.com/photoprism/photoprism/pkg/txt"
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
	t.Run("Homophones", func(t *testing.T) {
		label := NewLabel("老板", 10)
		assert.Equal(t, "老板", label.LabelName)
		assert.Equal(t, "lao-ban", label.LabelSlug)
		assert.Equal(t, 10, label.LabelPriority)
		label = NewLabel("老伴", 10)
		assert.Equal(t, "老伴", label.LabelName)
		assert.Equal(t, "lao-ban", label.LabelSlug)
		assert.Equal(t, 10, label.LabelPriority)
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
	t.Run("First", func(t *testing.T) {
		// Find flower
		label := LabelFixtures.Get("flower")
		result := FirstOrCreateLabel(&label)

		if !assert.NotNil(t, result) {
			t.Fatal("result should not be nil")
		}
		assert.Equal(t, label.LabelName, result.LabelName, "LabelName should be the same")
		assert.Equal(t, label.LabelSlug, result.LabelSlug, "LabelSlug should be the same")

		// Find Batch Delete
		label = *NewLabel("Batch Delete", 10)
		resultBase := FirstOrCreateLabel(&label)
		if !assert.NotNil(t, resultBase) {
			t.Fatal("resultBase should not be nil")
		}
		assert.Equal(t, LabelFixtures.Get("batchdelete").LabelUID, resultBase.LabelUID)

		// Find Batch Delete with -
		label = *NewLabel("Batch-Delete", 10)
		result = FirstOrCreateLabel(&label)
		if !assert.NotNil(t, result) {
			t.Fatal("result should not be nil")
		}
		assert.Equal(t, LabelFixtures.Get("batchdelete").LabelUID, result.LabelUID)

		// Find Batch Delete lowercase
		label = *NewLabel("batch delete'", 10)
		result = FirstOrCreateLabel(&label)
		if !assert.NotNil(t, result) {
			t.Fatal("result should not be nil")
		}
		assert.Equal(t, LabelFixtures.Get("batchdelete").LabelUID, result.LabelUID)

		// Find Batch Delete with a unicode graphic character
		label = *NewLabel("BATCH DELETE🢱", 10)
		result = FirstOrCreateLabel(&label)
		if !assert.NotNil(t, result) {
			t.Fatal("result should not be nil")
		}
		assert.Equal(t, LabelFixtures.Get("batchdelete").LabelUID, result.LabelUID)
	})

	t.Run("Homophones", func(t *testing.T) {
		// Add a homophone
		label := NewLabel("老板", 10)
		assert.Equal(t, "老板", label.LabelName)
		assert.Equal(t, "lao-ban", label.LabelSlug)
		assert.Equal(t, 10, label.LabelPriority)

		result1 := FirstOrCreateLabel(label)

		if !assert.NotNil(t, result1) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.Equal(t, label.LabelName, result1.LabelName, "LabelName should be the same")
		assert.Equal(t, label.LabelSlug, result1.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, "", result1.MaxHomophone)

		// Add another homophone
		label2 := NewLabel("老伴", 10)
		assert.Equal(t, "老伴", label2.LabelName)
		assert.Equal(t, "lao-ban", label2.LabelSlug)
		assert.Equal(t, 10, label2.LabelPriority)

		result2 := FirstOrCreateLabel(label2)
		if !assert.NotNil(t, result2) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.NotEqual(t, label.LabelName, label2.LabelName, "LabelName should not be the same")
		assert.NotEqual(t, label.LabelSlug, label2.LabelSlug, "LabelSlug should not be the same")
		assert.Equal(t, "lao-ban-c-a", result2.LabelSlug)
		assert.Equal(t, "a", result2.MaxHomophone)

		// Add the homophone in ascii
		label3 := NewLabel("lao-ban", 10)
		result3 := FirstOrCreateLabel(label3)
		if !assert.NotNil(t, result3) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.NotEqual(t, label.LabelName, label3.LabelName, "LabelName should be the same")
		assert.NotEqual(t, label.LabelSlug, label3.LabelSlug, "LabelSlug should not be the same")
		assert.Equal(t, "lao-ban-c-b", result3.LabelSlug)
		assert.Equal(t, "b", result3.MaxHomophone)

		assert.NotEqual(t, result1.LabelUID, result2.LabelUID)
		assert.NotEqual(t, result1.LabelUID, result3.LabelUID)
		assert.NotEqual(t, result2.LabelUID, result3.LabelUID)

		// Make sure that we find the correct homophone
		label1a := NewLabel("老板", 10)
		result1a := FirstOrCreateLabel(label1a)
		if !assert.NotNil(t, result1a) {
			t.Fatal("result should not be nil")
		}
		assert.Equal(t, result1.LabelUID, result1a.LabelUID)
		assert.Equal(t, "b", result1a.MaxHomophone)

		label2a := NewLabel("老伴", 10)
		result2a := FirstOrCreateLabel(label2a)
		if !assert.NotNil(t, result2a) {
			t.Fatal("result should not be nil")
		}
		assert.Equal(t, result2.LabelUID, result2a.LabelUID)
		assert.Equal(t, "b", result2a.MaxHomophone)

		label3a := NewLabel("lao-ban", 10)
		result3a := FirstOrCreateLabel(label3a)
		if !assert.NotNil(t, result3a) {
			t.Fatal("result should not be nil")
		}
		assert.Equal(t, result3.LabelUID, result3a.LabelUID)
		assert.Equal(t, "b", result3a.MaxHomophone)

		assert.NoError(t, UnscopedDb().Delete(&result1).Error)
		assert.NoError(t, UnscopedDb().Delete(&result2).Error)
		assert.NoError(t, UnscopedDb().Delete(&result3).Error)
	})
	t.Run("ExceedHomophones", func(t *testing.T) {
		label := NewLabel("送钟", 10)
		result := FirstOrCreateLabel(label)
		//t.Logf("result = %+v", result)
		assert.Equal(t, "song-zhong", result.LabelSlug)
		assert.NotNil(t, result)
		label = NewLabel("song-zhong", 10)
		result = FirstOrCreateLabel(label)
		assert.NotNil(t, result)
		assert.Equal(t, "song-zhong-c-a", result.LabelSlug)
		// Force increment to maximum
		label = NewLabel("song-zhong-c-z", 10)
		result = FirstOrCreateLabel(label)
		assert.NotNil(t, result)
		assert.Equal(t, "song-zhong-c-z", result.LabelSlug)
		// This one should fail as we have run out of incrementers
		label = NewLabel("送终", 10)
		result = FirstOrCreateLabel(label)
		//t.Logf("result = %+v", result)
		assert.Nil(t, result)
	})
	t.Run("UnicodeEmoji", func(t *testing.T) {
		// Test emitocons
		label := NewLabel("😖😕", 10)
		assert.Equal(t, "😖😕", label.LabelName)
		assert.Equal(t, "_5cpzrfxqt5mjk", label.LabelSlug)
		assert.Equal(t, 10, label.LabelPriority)

		result := FirstOrCreateLabel(label)

		if !assert.NotNil(t, result) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.Equal(t, "😖😕", result.LabelName, "LabelName should be the same")
		assert.Equal(t, "_5cpzrfxqt5mjk", result.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, 10, result.LabelPriority)

		label = NewLabel("😖😕", 1)
		result2 := FirstOrCreateLabel(label)
		if !assert.NotNil(t, result2) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.Equal(t, "😖😕", result2.LabelName, "LabelName should be the same")
		assert.Equal(t, "_5cpzrfxqt5mjk", result2.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, result.LabelUID, result2.LabelUID, "LabelUID should be the same")
		assert.Equal(t, 10, result2.LabelPriority)

		// Test Unicode AND Emoticons
		label = NewLabel("😖அஆஇ😕", 10)
		assert.Equal(t, "😖அஆஇ😕", label.LabelName)
		assert.Equal(t, "aaai", label.LabelSlug)
		assert.Equal(t, 10, label.LabelPriority)

		result = FirstOrCreateLabel(label)

		if !assert.NotNil(t, result) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.Equal(t, "😖அஆஇ😕", result.LabelName, "LabelName should be the same")
		assert.Equal(t, "aaai", result.LabelSlug, "LabelSlug should be the same")

		label = NewLabel("😖அஆஇ😕", 1)
		result2 = FirstOrCreateLabel(label)
		if !assert.NotNil(t, result2) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.Equal(t, "😖அஆஇ😕", result2.LabelName, "LabelName should be the same")
		assert.Equal(t, "aaai", result2.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, result.LabelUID, result2.LabelUID, "LabelUID should be the same")
		assert.Equal(t, 10, result2.LabelPriority)

		// Unicode Only to find with Emoticons
		label = NewLabel("அஆஇ", 1)
		result2 = FirstOrCreateLabel(label)
		if !assert.NotNil(t, result2) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.Equal(t, "😖அஆஇ😕", result2.LabelName, "LabelName should be the same")
		assert.Equal(t, "aaai", result2.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, result.LabelUID, result2.LabelUID, "LabelUID should be the same")
		assert.Equal(t, 10, result2.LabelPriority)

		// Test Unicode Only
		label = NewLabel("அஆஇண", 10)
		assert.Equal(t, "அஆஇண", label.LabelName)
		assert.Equal(t, "aaainn", label.LabelSlug)
		assert.Equal(t, 10, label.LabelPriority)

		result = FirstOrCreateLabel(label)

		if !assert.NotNil(t, result) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.Equal(t, "அஆஇண", result.LabelName, "LabelName should be the same")
		assert.Equal(t, "aaainn", result.LabelSlug, "LabelSlug should be the same")

		label = NewLabel("அஆஇண", 1)
		result2 = FirstOrCreateLabel(label)
		if !assert.NotNil(t, result2) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.Equal(t, "அஆஇண", result2.LabelName, "LabelName should be the same")
		assert.Equal(t, "aaainn", result2.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, result.LabelUID, result2.LabelUID, "LabelUID should be the same")
		assert.Equal(t, 10, result2.LabelPriority)

	})
	t.Run("Renamed", func(t *testing.T) {
		// Find cow
		label := LabelFixtures.Get("cow")
		result := FirstOrCreateLabel(&label)

		if !assert.NotNil(t, result) {
			t.Fatal("result should not be nil")
		}
		assert.Equal(t, label.LabelName, result.LabelName, "LabelName should be the same")
		assert.Equal(t, label.LabelSlug, result.LabelSlug, "LabelSlug should be the same")

		label1 := NewLabel("Cow", 5)
		result1 := FirstOrCreateLabel(label1)
		require.NotNil(t, result1)

		assert.Equal(t, label.ID, result1.ID)
	})
	t.Run("AmpersandVsAnd", func(t *testing.T) {
		// Add base record
		label := NewLabel("Fire and Station", 10)
		assert.Equal(t, "Fire and Station", label.LabelName)
		assert.Equal(t, "fire-and-station", label.LabelSlug)
		assert.Equal(t, 10, label.LabelPriority)

		result1 := FirstOrCreateLabel(label)

		if !assert.NotNil(t, result1) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.Equal(t, label.LabelName, result1.LabelName, "LabelName should be the same")
		assert.Equal(t, label.LabelSlug, result1.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, "", result1.MaxHomophone)

		// Find Base Record with slug
		label2 := NewLabel("Fire & Station", 10)
		assert.Equal(t, "Fire & Station", label2.LabelName)
		assert.Equal(t, "fire-and-station", label2.LabelSlug)
		assert.Equal(t, 10, label2.LabelPriority)

		result2 := FirstOrCreateLabel(label2)
		if !assert.NotNil(t, result2) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.NotEqual(t, label.LabelName, label2.LabelName, "LabelName should not be the same")
		assert.Equal(t, label.LabelSlug, label2.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, "fire-and-station", result2.LabelSlug)
		assert.Equal(t, "", result2.MaxHomophone)

		assert.Equal(t, result1.LabelUID, result2.LabelUID)

		assert.NoError(t, UnscopedDb().Delete(&result1).Error)
		assert.NoError(t, UnscopedDb().Delete(&result2).Error)
	})
	t.Run("AtVsAt", func(t *testing.T) {
		// Add base record
		label := NewLabel("老伴 at 伤害", 10)
		assert.Equal(t, "老伴 at 伤害", label.LabelName)
		assert.Equal(t, "lao-ban-at-shang-hai", label.LabelSlug)
		assert.Equal(t, 10, label.LabelPriority)

		result1 := FirstOrCreateLabel(label)

		if !assert.NotNil(t, result1) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.Equal(t, label.LabelName, result1.LabelName, "LabelName should be the same")
		assert.Equal(t, label.LabelSlug, result1.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, "", result1.MaxHomophone)

		// Find Base Record with slug
		label2 := NewLabel("老伴 @ 伤害", 10)
		assert.Equal(t, "老伴 @ 伤害", label2.LabelName)
		assert.Equal(t, "lao-ban-at-shang-hai", label2.LabelSlug)
		assert.Equal(t, 10, label2.LabelPriority)

		result2 := FirstOrCreateLabel(label2)
		if !assert.NotNil(t, result2) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.NotEqual(t, label.LabelName, label2.LabelName, "LabelName should not be the same")
		assert.Equal(t, label.LabelSlug, label2.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, "lao-ban-at-shang-hai", result2.LabelSlug)
		assert.Equal(t, "", result2.MaxHomophone)

		assert.Equal(t, result1.LabelUID, result2.LabelUID)

		label3 := NewLabel("老伴 @ 上海", 10)
		assert.Equal(t, "老伴 @ 上海", label3.LabelName)
		assert.Equal(t, "lao-ban-at-shang-hai", label3.LabelSlug)
		assert.Equal(t, 10, label3.LabelPriority)

		result3 := FirstOrCreateLabel(label3)
		if !assert.NotNil(t, result3) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.NotEqual(t, label.LabelName, label3.LabelName, "LabelName should not be the same")
		assert.NotEqual(t, label.LabelSlug, label3.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, "lao-ban-at-shang-hai-c-a", result3.LabelSlug)
		assert.Equal(t, "a", result3.MaxHomophone)

		assert.NotEqual(t, result1.LabelUID, result3.LabelUID)

		assert.NoError(t, UnscopedDb().Delete(&result1).Error)
		assert.NoError(t, UnscopedDb().Delete(&result2).Error)
		assert.NoError(t, UnscopedDb().Delete(&result3).Error)
	})
	t.Run("RemovedRunes", func(t *testing.T) {
		// Add base record
		label := NewLabel("Fire Station", 10)
		assert.Equal(t, "Fire Station", label.LabelName)
		assert.Equal(t, "fire-station", label.LabelSlug)
		assert.Equal(t, 10, label.LabelPriority)

		result1 := FirstOrCreateLabel(label)

		if !assert.NotNil(t, result1) {
			t.Fatal("result must not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.Equal(t, label.LabelName, result1.LabelName, "LabelName should be the same")
		assert.Equal(t, label.LabelSlug, result1.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, "", result1.MaxHomophone)

		// Find Base Record with slug
		label2 := NewLabel("fire-station", 10)
		assert.Equal(t, "Fire-Station", label2.LabelName)
		assert.Equal(t, "fire-station", label2.LabelSlug)
		assert.Equal(t, 10, label2.LabelPriority)

		result2 := FirstOrCreateLabel(label2)
		if !assert.NotNil(t, result2) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.NotEqual(t, label.LabelName, label2.LabelName, "LabelName should not be the same")
		assert.Equal(t, label.LabelSlug, label2.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, "fire-station", result2.LabelSlug)
		assert.Equal(t, "", result2.MaxHomophone)

		// Find Base Record due to character removal
		label3 := NewLabel("Fire+Station", 10)
		assert.Equal(t, "Fire+Station", label3.LabelName)
		assert.Equal(t, "fire-station", label3.LabelSlug)
		assert.Equal(t, 10, label3.LabelPriority)
		result3 := FirstOrCreateLabel(label3)
		if !assert.NotNil(t, result3) {
			t.Fatal("result should not be nil")
		}
		//t.Logf("result = %+v", result)
		assert.NotEqual(t, label.LabelName, label3.LabelName, "LabelName should be the same")
		assert.Equal(t, label.LabelSlug, label3.LabelSlug, "LabelSlug should be the same")
		assert.Equal(t, "fire-station", result3.LabelSlug)
		assert.Equal(t, "", result3.MaxHomophone)

		assert.Equal(t, result1.LabelUID, result2.LabelUID)
		assert.Equal(t, result1.LabelUID, result3.LabelUID)
		assert.Equal(t, result2.LabelUID, result3.LabelUID)

		assert.NoError(t, UnscopedDb().Delete(&result1).Error)
		assert.NoError(t, UnscopedDb().Delete(&result2).Error)
		assert.NoError(t, UnscopedDb().Delete(&result3).Error)
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

func TestFindLabels(t *testing.T) {
	t.Run("SuccessSingular", func(t *testing.T) {
		label := LabelFixtures.Get("flower")
		labels, err := FindLabels(label.LabelName, txt.Or, false)
		assert.NoError(t, err)
		if assert.Len(t, labels, 1) {
			assert.Equal(t, label.ID, labels[0].ID)
		}
	})
	t.Run("SuccessPlural", func(t *testing.T) {
		label := LabelFixtures.Get("flower")
		labels, err := FindLabels(label.LabelName+"s", txt.Or, false)
		assert.NoError(t, err)
		if assert.Len(t, labels, 1) {
			assert.Equal(t, label.ID, labels[0].ID)
		}
	})
	t.Run("SuccessMultiple", func(t *testing.T) {
		label1 := LabelFixtures.Get("landscape")
		label2 := LabelFixtures.Get("flower")
		labels, err := FindLabels(fmt.Sprintf("%s%s%s", label1.LabelName, txt.Or, label2.LabelName), txt.Or, false)
		assert.NoError(t, err)
		found1 := false
		found2 := false
		if assert.Len(t, labels, 2) {
			for _, label := range labels {
				if label.ID == label1.ID {
					found1 = true
				} else if label.ID == label2.ID {
					found2 = true
				} else {
					assert.Failf(t, "unable to match", "%+v", label)
				}
			}
			assert.True(t, found1, "Unable to find %+v", label1)
			assert.True(t, found2, "Unable to find %+v", label2)
		}
	})
	t.Run("SuccessHomophone", func(t *testing.T) {
		label1 := LabelFixtures.Get("shanghai1")
		label2 := LabelFixtures.Get("shanghai2")
		labels, err := FindLabels(label1.LabelName, txt.Or, false)
		assert.NoError(t, err)
		found1 := false
		found2 := false
		if assert.Len(t, labels, 1) {
			for _, label := range labels {
				if label.ID == label1.ID {
					found1 = true
				} else if label.ID == label2.ID {
					found2 = true
				} else {
					assert.Failf(t, "unable to match", "%+v", label)
				}
			}
			assert.True(t, found1, "Unable to find %+v", label1)
			assert.False(t, found2, "Able to find %+v", label2)
		}
	})
	t.Run("SuccessHomophones", func(t *testing.T) {
		label1 := LabelFixtures.Get("shanghai1")
		label2 := LabelFixtures.Get("shanghai2")
		labels, err := FindLabels(fmt.Sprintf("%s%s%s", label1.LabelName, txt.Or, label2.LabelName), txt.Or, false)
		assert.NoError(t, err)
		found1 := false
		found2 := false
		if assert.Len(t, labels, 2) {
			for _, label := range labels {
				if label.ID == label1.ID {
					found1 = true
				} else if label.ID == label2.ID {
					found2 = true
				} else {
					assert.Failf(t, "unable to match", "%+v", label)
				}
			}
			assert.True(t, found1, "Unable to find %+v", label1)
			assert.True(t, found2, "Unable to find %+v", label2)
		}
	})
	t.Run("SuccessHomophoneSlug", func(t *testing.T) {
		label1 := LabelFixtures.Get("shanghai1")
		label2 := LabelFixtures.Get("shanghai2")
		labels, err := FindLabels(label1.LabelSlug, txt.Or, false)
		assert.NoError(t, err)
		found1 := false
		found2 := false
		if assert.Len(t, labels, 1) {
			for _, label := range labels {
				if label.ID == label1.ID {
					found1 = true
				} else if label.ID == label2.ID {
					found2 = true
				} else {
					assert.Failf(t, "unable to match", "%+v", label)
				}
			}
			assert.True(t, found1, "Unable to find %+v", label1)
			assert.False(t, found2, "Able to find %+v", label2)
		}
	})
	t.Run("SuccessHomophoneSlug-A", func(t *testing.T) {
		label1 := LabelFixtures.Get("shanghai1")
		label2 := LabelFixtures.Get("shanghai2")
		labels, err := FindLabels(label2.LabelSlug, txt.Or, false)
		assert.NoError(t, err)
		found1 := false
		found2 := false
		if assert.Len(t, labels, 1) {
			for _, label := range labels {
				if label.ID == label1.ID {
					found1 = true
				} else if label.ID == label2.ID {
					found2 = true
				} else {
					assert.Failf(t, "unable to match", "%+v", label)
				}
			}
			assert.False(t, found1, "Able to find %+v", label1)
			assert.True(t, found2, "Unable to find %+v", label2)
		}
	})

}
