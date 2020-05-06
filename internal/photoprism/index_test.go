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

	tf := classify.New(conf.ResourcesPath(), conf.DisableTensorFlow())
	nd := nsfw.New(conf.NSFWModelPath())
	convert := NewConvert(conf)

	ind := NewIndex(conf, tf, nd, convert)
	imp := NewImport(conf, ind, convert)
	opt := ImportOptionsMove(conf.ImportPath())

	imp.Start(opt)

	indexOpt := IndexOptionsAll()

	ind.Start(indexOpt)
}
