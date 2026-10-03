package event

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/clean"
)

func TestFormat(t *testing.T) {
	assert.Equal(t, "", Format(nil))
	assert.Equal(t, "user", Format([]string{"user"}))
	assert.Equal(t, "user › michael › not found", Format([]string{"user", "michael", "not found"}))

	result := Format([]string{"user", "%s not found"}, clean.LogQuote("michael"))
	expected := "user › 'michael' not found"

	assert.Equal(t, expected, result)

	t.Log(result)
}

func TestFormatSegmentSeparator(t *testing.T) {
	sep := string(clean.FieldSep)
	t.Run("WithoutArgs", func(t *testing.T) {
		// The separator count follows the segment list, whatever the segments carry.
		noSep := Format([]string{"ip", "users", "value", "denied"})
		withSep := Format([]string{"ip", "users", "a" + sep + "b", "denied"})
		assert.Equal(t, strings.Count(noSep, sep), strings.Count(withSep, sep))
		assert.NotContains(t, withSep, "a"+sep+"b")
	})
	t.Run("WithArgs", func(t *testing.T) {
		noSep := Format([]string{"ip", "session %s", "value", "denied"}, "ref")
		withSep := Format([]string{"ip", "session %s", "a" + sep + "b", "denied"}, "ref")
		assert.Equal(t, strings.Count(noSep, sep), strings.Count(withSep, sep))
	})
	t.Run("ArgumentIsTheCallersToSanitize", func(t *testing.T) {
		// A value the caller quoted keeps what it holds, because quotes already show its extent.
		out := Format([]string{"ip", "users", "%s", "denied"}, clean.LogQuote("a"+sep+"b"))
		assert.Contains(t, out, "'a"+sep+"b'")
	})
}

func TestFormatSegmentVerb(t *testing.T) {
	t.Run("WithoutArgs", func(t *testing.T) {
		// Without arguments the segments are text, so a verb among them is not a verb.
		assert.Equal(t, "ip › users › %s%!q › denied", Format([]string{"ip", "users", "%s%!q", "denied"}))
	})
	t.Run("WithArgs", func(t *testing.T) {
		assert.Equal(t, "ip › session ref › denied", Format([]string{"ip", "session %s", "denied"}, "ref"))
	})
}

func TestFormatSegments(t *testing.T) {
	sep := string(clean.FieldSep)
	t.Run("SliceIsReusedWhenNothingChanges", func(t *testing.T) {
		ev := []string{"ip", "users", "denied"}
		out := formatSegments(ev)
		assert.Equal(t, ev, out)
		assert.Same(t, &ev[0], &out[0])
	})
	t.Run("CallerSliceIsNotModified", func(t *testing.T) {
		ev := []string{"ip", "users", "a" + sep + "b"}
		out := formatSegments(ev)
		assert.Equal(t, "a?b", out[2])
		assert.Equal(t, "a"+sep+"b", ev[2])
	})
	t.Run("EverySegmentIsChecked", func(t *testing.T) {
		ev := []string{"a" + sep + "b", "c", "d" + sep + "e"}
		assert.Equal(t, []string{"a?b", "c", "d?e"}, formatSegments(ev))
	})
}
