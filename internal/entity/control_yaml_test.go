package entity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestPhoto_LoadFromYaml_ControlCharacters(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "control.yml")
	data := "UID: ps6sg6be2lvl0yh0\n" +
		"Title: \"Sea\\0side\\e\"\n" +
		"TitleSrc: manual\n" +
		"Caption: \"First line\\0\\r\\nSecond line\"\n" +
		"CaptionSrc: manual\n" +
		"Exposure: \"1/60\\0\"\n" +
		"Details:\n" +
		"  Notes: \"Notes\\0\\nnext\"\n"
	require.NoError(t, os.WriteFile(fileName, []byte(data), fs.ModeFile))

	m := Photo{}
	require.NoError(t, m.LoadFromYaml(fileName))
	assert.Equal(t, "Seaside", m.PhotoTitle)
	assert.Equal(t, "First line\nSecond line", m.PhotoCaption)
	assert.Equal(t, "1/60", m.PhotoExposure)
	require.NotNil(t, m.Details)

	m.Details.PhotoID = 900000012
	deleteDetails(t, m.Details.PhotoID)
	t.Cleanup(func() { deleteDetails(t, 900000012) })
	require.NoError(t, m.Details.Save())
	assert.Equal(t, "Notes\nnext", findDetails(t, 900000012).Notes)
}

func TestAlbum_LoadFromYaml_ControlCharacters(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "control.yml")
	data := "UID: as6sg6bxpoaaaaz1\n" +
		"Slug: control-album\n" +
		"Type: album\n" +
		"Title: \"Control\\0 Album\\e\"\n" +
		"Description: \"First\\0\\r\\nSecond\"\n" +
		"Location: \"Berlin\\x01\"\n"
	require.NoError(t, os.WriteFile(fileName, []byte(data), fs.ModeFile))

	m := Album{}
	require.NoError(t, m.LoadFromYaml(fileName))
	assert.Equal(t, "Control Album", m.AlbumTitle)
	assert.Equal(t, "control-album", m.AlbumSlug)
	assert.Equal(t, "First\nSecond", m.AlbumDescription)
	assert.Equal(t, "Berlin", m.AlbumLocation)
}
