package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/config"
)

func TestNewImport(t *testing.T) {
	conf := config.TestConfig()

	nd := nsfw.NewModel(conf.NSFWModelPath())
	fn := face.NewModel(conf.FaceNetModelPath(), "", conf.DisableTensorFlow())
	convert := NewConvert(conf)

	ind := NewIndex(conf, nd, fn, convert, NewFiles(), NewPhotos())
	imp := NewImport(conf, ind, convert)

	assert.IsType(t, &Import{}, imp)
}

func TestImport_DestinationFilename(t *testing.T) {
	conf := config.TestConfig()

	if err := conf.InitializeTestData(); err != nil {
		t.Fatal(err)
	}

	nd := nsfw.NewModel(conf.NSFWModelPath())
	fn := face.NewModel(conf.FaceNetModelPath(), "", conf.DisableTensorFlow())
	convert := NewConvert(conf)

	ind := NewIndex(conf, nd, fn, convert, NewFiles(), NewPhotos())

	imp := NewImport(conf, ind, convert)

	rawFile, err := NewMediaFile(conf.ImportPath() + "/raw/IMG_2567.CR2")

	if err != nil {
		t.Fatal(err)
	}

	t.Run("NoBasePath", func(t *testing.T) {
		fileName, err := imp.DestinationFilename(rawFile, rawFile, "")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, conf.OriginalsPath()+"/2019/07/20190705_153230_C167C6FD.cr2", fileName)
	})

	t.Run("WithBasePath", func(t *testing.T) {
		fileName, err := imp.DestinationFilename(rawFile, rawFile, "users/guest")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, conf.OriginalsPath()+"/users/guest/2019/07/20190705_153230_C167C6FD.cr2", fileName)
	})
}

func TestImport_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	conf := config.TestConfig()

	conf.InitializeTestData()

	nd := nsfw.NewModel(conf.NSFWModelPath())
	fn := face.NewModel(conf.FaceNetModelPath(), "", conf.DisableTensorFlow())
	convert := NewConvert(conf)

	ind := NewIndex(conf, nd, fn, convert, NewFiles(), NewPhotos())

	imp := NewImport(conf, ind, convert)

	opt := ImportOptionsMove(conf.ImportPath(), "")

	imp.Start(opt)
}
