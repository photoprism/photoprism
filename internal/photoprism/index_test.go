package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/nsfw"
)

func TestIndex_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	conf.InitializeTestData(t)

	tf := classify.New(conf.ResourcesPath(), conf.TensorFlowDisabled())
	nd := nsfw.New(conf.NSFWModelPath())

	ind := NewIndex(conf, tf, nd)

	convert := NewConvert(conf)

	imp := NewImport(conf, ind, convert)

	imp.Start(conf.ImportPath())

	opt := IndexOptionsAll()

	ind.Start(opt)
}
