package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetColors(t *testing.T) {
	conf := test.NewConfig()

	conf.InitializeTestData(t)

	if mediaFile1, err := NewMediaFile(conf.ImportPath() + "/dog.jpg"); err == nil {
		names, vibrantHex, mutedHex := mediaFile1.GetColors()

		t.Log(names, vibrantHex, mutedHex)

		assert.IsType(t, []string{}, names)
		assert.Equal(t, "#e0ed21", vibrantHex)
		assert.Equal(t, "#977d67", mutedHex)
		assert.Equal(t, []string([]string{"black", "brown", "grey", "white"}), names);
	} else {
		t.Error(err)
	}

	if mediaFile2, err := NewMediaFile(conf.ImportPath() + "/ape.jpeg"); err == nil {
		names, vibrantHex, mutedHex := mediaFile2.GetColors()

		t.Log(names, vibrantHex, mutedHex)

		assert.IsType(t, []string{}, names)
		assert.Equal(t, "#97c84a", vibrantHex)
		assert.Equal(t, "#6c9a68", mutedHex)
		assert.Equal(t, []string([]string{"grey", "teal", "white"}), names);
	} else {
		t.Error(err)
	}
}
