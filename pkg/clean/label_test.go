package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelName(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", LabelName(""))
	})
	t.Run("Fire Station", func(t *testing.T) {
		assert.Equal(t, "fire station", LabelName("Fire Station"))
		assert.Equal(t, "fire station", LabelName("Fire' Station"))
		assert.Equal(t, "fire station", LabelName("😀Fire Station"))
		assert.Equal(t, "fire station", LabelName("Fire Station◨"))
		assert.Equal(t, "fire station", LabelName("Fire-Station"))
	})
	t.Run("ABC", func(t *testing.T) {
		assert.Equal(t, "abc", LabelName("Abc"))
		assert.Equal(t, "abc", LabelName("abc."))
		assert.Equal(t, "abc", LabelName(".abc."))
		assert.Equal(t, "abc", LabelName("😀ABC "))
		assert.Equal(t, "abc", LabelName("aB◨c"))
		assert.Equal(t, "abc", LabelName("abc-"))
	})
	t.Run("Unicode", func(t *testing.T) {
		assert.Equal(t, "送钟", LabelName("送钟"))
		assert.Equal(t, "送终", LabelName("送终"))
		assert.Equal(t, "送 终", LabelName("送-终"))
		assert.Equal(t, "送 终", LabelName(" 送 终 "))
		assert.Equal(t, "送终", LabelName("-送终"))
	})
	t.Run("Mixed", func(t *testing.T) {
		assert.Equal(t, "送钟abc", LabelName("送钟AbC"))
		assert.Equal(t, "даbc", LabelName("ДаBc"))
		assert.Equal(t, "так", LabelName("ТАК"))
	})
	t.Run("Graphic", func(t *testing.T) {
		assert.Equal(t, "😀", LabelName("😀"))
		assert.Equal(t, "◨", LabelName("◨"))
		assert.Equal(t, "😀◨😀", LabelName("😀◨😀"))
		assert.Equal(t, "😀◨😀", LabelName(" 😀◨😀 "))
		assert.Equal(t, "😀 ◨😀", LabelName(" 😀-◨😀 "))
		assert.Equal(t, "😀◨ 😀", LabelName(" 😀◨ 😀 "))
	})
}
