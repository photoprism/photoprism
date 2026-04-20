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

func TestUnescapeLabelTerm(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", unescapeLabelTerm(""))
	})
	t.Run("WithoutEscape", func(t *testing.T) {
		assert.Equal(t, "cake", unescapeLabelTerm("cake"))
	})
	t.Run("EscapedOperators", func(t *testing.T) {
		assert.Equal(t, "!weird", unescapeLabelTerm(`\!weird`))
		assert.Equal(t, "cat&dog", unescapeLabelTerm(`cat\&dog`))
		assert.Equal(t, "a|b", unescapeLabelTerm(`a\|b`))
		assert.Equal(t, `a\b`, unescapeLabelTerm(`a\\b`))
	})
	t.Run("UnknownEscapePreserved", func(t *testing.T) {
		assert.Equal(t, `a\bc`, unescapeLabelTerm(`a\bc`))
	})
	t.Run("TrailingEscapePreserved", func(t *testing.T) {
		assert.Equal(t, `a\`, unescapeLabelTerm(`a\`))
	})
}

func TestResolveLabelGroup(t *testing.T) {
	t.Run("SingleExact", func(t *testing.T) {
		ids := resolveLabelGroup("cake")
		assert.Equal(t, []uint{LabelFixtures.Get("cake").ID}, ids)
	})
	t.Run("OrAlternatives", func(t *testing.T) {
		ids := resolveLabelGroup("cake|cow")
		assert.Contains(t, ids, LabelFixtures.Get("cake").ID)
		assert.Contains(t, ids, LabelFixtures.Get("cow").ID)
	})
	t.Run("CategoryExpansion", func(t *testing.T) {
		ids := resolveLabelGroup("landscape")
		assert.Contains(t, ids, LabelFixtures.Get("landscape").ID)
		assert.Contains(t, ids, LabelFixtures.Get("flower").ID)
	})
	t.Run("UnknownReturnsEmpty", func(t *testing.T) {
		assert.Empty(t, resolveLabelGroup("totally-unknown-label-name"))
	})
	t.Run("EscapedAmpersandName", func(t *testing.T) {
		ids := resolveLabelGroup(`construction\&failure`)
		assert.Equal(t, []uint{LabelFixtures.Get("construction&failure").ID}, ids)
	})
}

func TestParseLabelFilter(t *testing.T) {
	cakeID := LabelFixtures.Get("cake").ID
	cowID := LabelFixtures.Get("cow").ID
	flowerID := LabelFixtures.Get("flower").ID
	landscapeID := LabelFixtures.Get("landscape").ID

	t.Run("Empty", func(t *testing.T) {
		include, exclude, sawPositive, err := ParseLabelFilter("")
		require.NoError(t, err)
		assert.False(t, sawPositive)
		assert.Empty(t, include)
		assert.Empty(t, exclude)
	})
	t.Run("SinglePositive", func(t *testing.T) {
		include, exclude, sawPositive, err := ParseLabelFilter("cake")
		require.NoError(t, err)
		assert.True(t, sawPositive)
		require.Len(t, include, 1)
		assert.Equal(t, []uint{cakeID}, include[0])
		assert.Empty(t, exclude)
	})
	t.Run("SingleNegative", func(t *testing.T) {
		include, exclude, sawPositive, err := ParseLabelFilter("!cake")
		require.NoError(t, err)
		assert.False(t, sawPositive)
		assert.Empty(t, include)
		require.Len(t, exclude, 1)
		assert.Equal(t, []uint{cakeID}, exclude[0])
	})
	t.Run("IncludeAndExclude", func(t *testing.T) {
		include, exclude, sawPositive, err := ParseLabelFilter("cake&!flower")
		require.NoError(t, err)
		assert.True(t, sawPositive)
		require.Len(t, include, 1)
		assert.Equal(t, []uint{cakeID}, include[0])
		require.Len(t, exclude, 1)
		assert.Equal(t, []uint{flowerID}, exclude[0])
	})
	t.Run("MultipleIncludes", func(t *testing.T) {
		include, exclude, sawPositive, err := ParseLabelFilter("cake&cow")
		require.NoError(t, err)
		assert.True(t, sawPositive)
		require.Len(t, include, 2)
		assert.Equal(t, []uint{cakeID}, include[0])
		assert.Equal(t, []uint{cowID}, include[1])
		assert.Empty(t, exclude)
	})
	t.Run("OrWithinPositiveGroup", func(t *testing.T) {
		include, _, sawPositive, err := ParseLabelFilter("cake|cow")
		require.NoError(t, err)
		assert.True(t, sawPositive)
		require.Len(t, include, 1)
		assert.ElementsMatch(t, []uint{cakeID, cowID}, include[0])
	})
	t.Run("UnknownPositiveShortCircuits", func(t *testing.T) {
		_, _, sawPositive, err := ParseLabelFilter("cake&totally-unknown-label")
		assert.ErrorIs(t, err, ErrLabelNotFound)
		assert.True(t, sawPositive)
	})
	t.Run("UnknownNegativeIsNoOp", func(t *testing.T) {
		include, exclude, sawPositive, err := ParseLabelFilter("cake&!totally-unknown-label")
		require.NoError(t, err)
		assert.True(t, sawPositive)
		require.Len(t, include, 1)
		assert.Empty(t, exclude)
	})
	t.Run("NegativeCategoryExpansion", func(t *testing.T) {
		_, exclude, _, err := ParseLabelFilter("!landscape")
		require.NoError(t, err)
		require.Len(t, exclude, 1)
		assert.Contains(t, exclude[0], landscapeID)
		assert.Contains(t, exclude[0], flowerID)
	})
	t.Run("EscapedLeadingBang", func(t *testing.T) {
		// A leading \! is a literal label name. No label named "!weird"
		// exists in fixtures, so the lookup short-circuits.
		_, _, sawPositive, err := ParseLabelFilter(`\!weird`)
		assert.ErrorIs(t, err, ErrLabelNotFound)
		assert.True(t, sawPositive)
	})
	t.Run("NegatedEscapedBang", func(t *testing.T) {
		// "!\!weird" = NOT a label literally named "!weird"; unknown negative is a no-op.
		include, exclude, sawPositive, err := ParseLabelFilter(`!\!weird`)
		require.NoError(t, err)
		assert.False(t, sawPositive)
		assert.Empty(t, include)
		assert.Empty(t, exclude)
	})
	t.Run("EscapedAmpersandName", func(t *testing.T) {
		include, _, sawPositive, err := ParseLabelFilter(`construction\&failure`)
		require.NoError(t, err)
		assert.True(t, sawPositive)
		require.Len(t, include, 1)
		assert.Equal(t, []uint{LabelFixtures.Get("construction&failure").ID}, include[0])
	})
	t.Run("LoneBangDropped", func(t *testing.T) {
		include, exclude, sawPositive, err := ParseLabelFilter("!")
		require.NoError(t, err)
		assert.False(t, sawPositive)
		assert.Empty(t, include)
		assert.Empty(t, exclude)
	})
}
