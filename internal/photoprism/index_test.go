package photoprism

import (
	"testing"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/config"
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

	found, updated := ind.Start(indexOpt)
	assert.GreaterOrEqual(t, len(found), 0)
	assert.GreaterOrEqual(t, updated, 0)

	t.Logf("index run 1: found %s", english.Plural(updated, "file", "files"))
	t.Logf("index run 1: updated %s", english.Plural(updated, "file", "files"))

	time.Sleep(time.Second)

	found, updated = ind.Start(indexOpt)
	assert.GreaterOrEqual(t, len(found), 0)
	assert.GreaterOrEqual(t, updated, 0)

	t.Logf("index run 2: found %s", english.Plural(updated, "file", "files"))
	t.Logf("index run 2: updated %s", english.Plural(updated, "file", "files"))

	time.Sleep(time.Second)

	found, updated = ind.Start(indexOpt)
	assert.GreaterOrEqual(t, len(found), 0)
	assert.GreaterOrEqual(t, updated, 0)

	t.Logf("index run 3: found %s", english.Plural(updated, "file", "files"))
	t.Logf("index run 3: updated %s", english.Plural(updated, "file", "files"))
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
