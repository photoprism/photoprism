package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabelSlugs(t *testing.T) {
	t.Run("PipeSeparated", func(t *testing.T) {
		assert.Equal(t, []string{"cake", "flower"}, LabelSlugs("cake|flower", "|"))
	})
	t.Run("IncludesSingularASCIIForms", func(t *testing.T) {
		assert.Equal(t, []string{"cats", "cat", "dogs", "dog"}, LabelSlugs("cats dogs", " "))
	})
	t.Run("DeduplicatesRepeatedSlugs", func(t *testing.T) {
		assert.Equal(t, []string{"cake"}, LabelSlugs("cake cake", " "))
	})
}

func TestFindLabels(t *testing.T) {
	t.Run("FindsExistingNameIgnoringCase", func(t *testing.T) {
		labels, err := FindLabels("cow", " ")
		require.NoError(t, err)
		require.Len(t, labels, 1)
		assert.Equal(t, LabelFixtures.Get("cow").ID, labels[0].ID)
	})
	t.Run("FindsExactHomophoneName", func(t *testing.T) {
		first := FirstOrCreateLabel(NewLabel("问", 0))
		second := FirstOrCreateLabel(NewLabel("吻", 0))

		require.NotNil(t, first)
		require.NotNil(t, second)

		t.Cleanup(func() {
			_ = Db().Unscoped().Delete(second).Error
			_ = Db().Unscoped().Delete(first).Error
			FlushLabelCache()
		})

		labels, err := FindLabels("吻", " ")
		require.NoError(t, err)
		require.Len(t, labels, 1)
		assert.Equal(t, second.ID, labels[0].ID)
		assert.Equal(t, "吻", labels[0].LabelName)
	})
	t.Run("PreservesPipeSeparatorSemantics", func(t *testing.T) {
		labels, err := FindLabels("potato|couch", "|")
		assert.Error(t, err)
		assert.Empty(t, labels)
	})
	t.Run("MatchesTrailingPipeSlug", func(t *testing.T) {
		labels, err := FindLabels("mall|", "|")
		require.NoError(t, err)
		require.Len(t, labels, 1)
		assert.Equal(t, LabelFixtures.Get("mall|").ID, labels[0].ID)
	})
}

func TestFindLabelIDs(t *testing.T) {
	t.Run("WithoutCategories", func(t *testing.T) {
		labelIDs, err := FindLabelIDs("landscape", " ", false)
		require.NoError(t, err)
		assert.Equal(t, []uint{LabelFixtures.Get("landscape").ID}, labelIDs)
	})
	t.Run("IncludesCategories", func(t *testing.T) {
		labelIDs, err := FindLabelIDs("landscape", " ", true)
		require.NoError(t, err)
		assert.Contains(t, labelIDs, LabelFixtures.Get("landscape").ID)
		assert.Contains(t, labelIDs, LabelFixtures.Get("flower").ID)
		assert.Len(t, labelIDs, 2)
	})
}
