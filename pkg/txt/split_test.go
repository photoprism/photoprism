package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimmedSplitWithEscape(t *testing.T) {
	s := ` a\,b , c , \, d ` // escaped comma and escaped separator, spaces
	parts := TrimmedSplitWithEscape(s, ',', EscapeRune)
	// Expect trimming and escape handling; escaped separator stays in same token
	assert.Equal(t, []string{"a,b", "c", ", d"}, parts)
}

func TestUnTrimmedSplitWithEscape(t *testing.T) {
	s := ` a\,b , c `
	parts := UnTrimmedSplitWithEscape(s, ',', EscapeRune)
	// No trimming; spaces preserved around segments
	assert.Equal(t, []string{" a,b ", " c "}, parts)
}

func TestSplitWithEscape(t *testing.T) {
	t.Run("TrimmedEmptyString", func(t *testing.T) {
		testString := ""
		expected := []string{}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedNoSeparator", func(t *testing.T) {
		testString := "I Have No Separators"
		expected := []string{testString}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedOneEscapeNoSeparator", func(t *testing.T) {
		testString := `I Have An \ Escape`
		expected := []string{testString}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedOneSeparatorNoEscape", func(t *testing.T) {
		testString := "I Have A|Separator"
		expected := []string{"I Have A", "Separator"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedOneSeparatorNoEscapeLeadingSpaces", func(t *testing.T) {
		testString := "  I Have A|Separator"
		expected := []string{"I Have A", "Separator"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedOneSeparatorNoEscapeTrailingSpaces", func(t *testing.T) {
		testString := "I Have A|Separator  "
		expected := []string{"I Have A", "Separator"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedOneSeparatorNoEscapeLeadingSeparatorSpaces", func(t *testing.T) {
		testString := "I Have A  |Separator"
		expected := []string{"I Have A", "Separator"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedOneSeparatorNoEscapeTrailingSeparatorSpaces", func(t *testing.T) {
		testString := "I Have A|  Separator"
		expected := []string{"I Have A", "Separator"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedOneSeparatorNoEscapeSpacesEverywhere", func(t *testing.T) {
		testString := " I Have A | Separator "
		expected := []string{"I Have A", "Separator"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedOneSeparatorOneEscapedSeparator", func(t *testing.T) {
		testString := `I Have A|Separator and an \|Escape`
		expected := []string{"I Have A", "Separator and an |Escape"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedMultipleSeparators", func(t *testing.T) {
		testString := "One|Two|Three|Four|Five"
		expected := []string{"One", "Two", "Three", "Four", "Five"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedMultipleSeparatorsLeadingBlank", func(t *testing.T) {
		testString := "|One|Two|Three|Four|Five"
		expected := []string{"One", "Two", "Three", "Four", "Five"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedMultipleSeparatorsTrailingBlank", func(t *testing.T) {
		testString := "One|Two|Three|Four|Five|"
		expected := []string{"One", "Two", "Three", "Four", "Five"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedMultipleSeparatorsLeadingBlankWithSpaces", func(t *testing.T) {
		testString := "   | One | Two | Three | Four | Five "
		expected := []string{"One", "Two", "Three", "Four", "Five"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedMultipleSeparatorsTrailingBlankWithSpaces", func(t *testing.T) {
		testString := "  One | Two | Three | Four | Five |     "
		expected := []string{"One", "Two", "Three", "Four", "Five"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedMultipleSeparatorsTrailingBlankWithSpacesAndEscapes", func(t *testing.T) {
		testString := `  One | Two | Three | Four | Five | Si\x | Sev\|en |    `
		expected := []string{"One", "Two", "Three", "Four", "Five", `Si\x`, "Sev|en"}
		actual := SplitWithEscape(testString, '|', EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedFooBar", func(t *testing.T) {
		testString := ` foo & Bar&BAZ `
		expected := []string{"foo", "Bar", "BAZ"}
		actual := SplitWithEscape(testString, AndRune, EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("TrimmedFooBarNoSeparator", func(t *testing.T) {
		testString := ` foo & Bar&BAZ `
		expected := []string{` foo & Bar&BAZ `}
		actual := SplitWithEscape(testString, OrRune, EscapeRune, true)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedEmptyString", func(t *testing.T) {
		testString := ""
		expected := []string{}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedNoSeparator", func(t *testing.T) {
		testString := "I Have No Separators"
		expected := []string{testString}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedOneEscapeNoSeparator", func(t *testing.T) {
		testString := `I Have An \ Escape`
		expected := []string{testString}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedOneSeparatorNoEscape", func(t *testing.T) {
		testString := "I Have A|Separator"
		expected := []string{"I Have A", "Separator"}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedOneSeparatorNoEscapeLeadingSpaces", func(t *testing.T) {
		testString := "  I Have A|Separator"
		expected := []string{"  I Have A", "Separator"}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedOneSeparatorNoEscapeTrailingSpaces", func(t *testing.T) {
		testString := "I Have A|Separator  "
		expected := []string{"I Have A", "Separator  "}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedOneSeparatorNoEscapeLeadingSeparatorSpaces", func(t *testing.T) {
		testString := "I Have A  |Separator"
		expected := []string{"I Have A  ", "Separator"}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedOneSeparatorNoEscapeTrailingSeparatorSpaces", func(t *testing.T) {
		testString := "I Have A|  Separator"
		expected := []string{"I Have A", "  Separator"}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedOneSeparatorNoEscapeSpacesEverywhere", func(t *testing.T) {
		testString := " I Have A | Separator "
		expected := []string{" I Have A ", " Separator "}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedOneSeparatorOneEscapedSeparator", func(t *testing.T) {
		testString := `I Have A|Separator and an \|Escape`
		expected := []string{"I Have A", "Separator and an |Escape"}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedMultipleSeparators", func(t *testing.T) {
		testString := "One|Two|Three|Four|Five"
		expected := []string{"One", "Two", "Three", "Four", "Five"}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedMultipleSeparatorsLeadingBlank", func(t *testing.T) {
		testString := "|One|Two|Three|Four|Five"
		expected := []string{"", "One", "Two", "Three", "Four", "Five"}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedMultipleSeparatorsTrailingBlank", func(t *testing.T) {
		testString := "One|Two|Three|Four|Five|"
		expected := []string{"One", "Two", "Three", "Four", "Five", ""}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedMultipleSeparatorsLeadingBlankWithSpaces", func(t *testing.T) {
		testString := "   | One | Two | Three | Four | Five "
		expected := []string{"   ", " One ", " Two ", " Three ", " Four ", " Five "}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedMultipleSeparatorsTrailingBlankWithSpaces", func(t *testing.T) {
		testString := "  One | Two | Three | Four | Five |     "
		expected := []string{"  One ", " Two ", " Three ", " Four ", " Five ", "     "}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedMultipleSeparatorsTrailingBlankWithSpacesAndEscapes", func(t *testing.T) {
		testString := `  One | Two | Three | Four | Five | Si\x | Sev\|en |    `
		expected := []string{"  One ", " Two ", " Three ", " Four ", " Five ", ` Si\x `, " Sev|en ", "    "}
		actual := SplitWithEscape(testString, '|', EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedFooBar", func(t *testing.T) {
		testString := ` foo & Bar&BAZ `
		expected := []string{" foo ", " Bar", "BAZ "}
		actual := SplitWithEscape(testString, AndRune, EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
	t.Run("UnTrimmedFooBarNoSeparator", func(t *testing.T) {
		testString := ` foo & Bar&BAZ `
		expected := []string{` foo & Bar&BAZ `}
		actual := SplitWithEscape(testString, OrRune, EscapeRune, false)

		assert.Equal(t, expected, actual)
	})
}
