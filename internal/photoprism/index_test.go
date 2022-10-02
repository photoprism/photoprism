package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/face"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/nsfw"
)

func TestIndex_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	conf.InitializeTestData()

	tf := classify.New(conf.AssetsPath(), conf.DisableTensorFlow())
	nd := nsfw.New(conf.NSFWModelPath())
	fn := face.NewNet(conf.FaceNetModelPath(), "", conf.DisableTensorFlow())
	convert := NewConvert(conf)

	ind := NewIndex(conf, tf, nd, fn, convert, NewFiles(), NewPhotos())
	imp := NewImport(conf, ind, convert)
	opt := ImportOptionsMove(conf.ImportPath(), "")

	imp.Start(opt)

	indexOpt := IndexOptionsAll()
	indexOpt.Rescan = false

	ind.Start(indexOpt)
}

func TestIndex_File(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	conf.InitializeTestData()

	tf := classify.New(conf.AssetsPath(), conf.DisableTensorFlow())
	nd := nsfw.New(conf.NSFWModelPath())
	fn := face.NewNet(conf.FaceNetModelPath(), "", conf.DisableTensorFlow())
	convert := NewConvert(conf)

	ind := NewIndex(conf, tf, nd, fn, convert, NewFiles(), NewPhotos())

	err := ind.FileName("xxx", IndexOptionsAll())

	assert.Equal(t, IndexFailed, err.Status)
}
