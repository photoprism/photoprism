package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/nsfw"
	"github.com/stretchr/testify/assert"
)

func TestNewImport(t *testing.T) {
	conf := config.TestConfig()

	tf := classify.New(conf.AssetsPath(), conf.TensorFlowOff())
	nd := nsfw.New(conf.NSFWModelPath())
	convert := NewConvert(conf)

	ind := NewIndex(conf, tf, nd, convert, NewFiles(), NewPhotos())
	imp := NewImport(conf, ind, convert)

	assert.IsType(t, &Import{}, imp)
}

func TestImport_DestinationFilename(t *testing.T) {
	conf := config.TestConfig()

	conf.InitializeTestData(t)

	tf := classify.New(conf.AssetsPath(), conf.TensorFlowOff())
	nd := nsfw.New(conf.NSFWModelPath())
	convert := NewConvert(conf)

	ind := NewIndex(conf, tf, nd, convert, NewFiles(), NewPhotos())

	imp := NewImport(conf, ind, convert)

	rawFile, err := NewMediaFile(conf.ImportPath() + "/raw/IMG_2567.CR2")

	if err != nil {
		t.Fatal(err)
	}

	fileName, err := imp.DestinationFilename(rawFile, rawFile)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, conf.OriginalsPath()+"/2019/07/20190705_153230_C167C6FD.cr2", fileName)
}

func TestImport_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	conf.InitializeTestData(t)

	tf := classify.New(conf.AssetsPath(), conf.TensorFlowOff())
	nd := nsfw.New(conf.NSFWModelPath())
	convert := NewConvert(conf)

	ind := NewIndex(conf, tf, nd, convert, NewFiles(), NewPhotos())

	imp := NewImport(conf, ind, convert)

	opt := ImportOptionsMove(conf.ImportPath())

	imp.Start(opt)
}
