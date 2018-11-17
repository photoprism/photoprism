package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestMediaFile_GetColors(t *testing.T) {
	conf := test.NewConfig()

	conf.InitializeTestData(t)

	if mediaFile1, err := NewMediaFile(conf.GetImportPath() + "/dog.jpg"); err == nil {
		names, vibrantHex, mutedHex := mediaFile1.GetColors()

		t.Log(names, vibrantHex, mutedHex)

		assert.IsType(t, []string{}, names)
		assert.Equal(t, "#e0ed21", vibrantHex)
		assert.Equal(t, "#977d67", mutedHex)
	} else {
		t.Error(err)
	}
}
