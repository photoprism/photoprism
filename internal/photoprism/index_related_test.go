package photoprism

import (
	"path/filepath"
	"testing"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/nsfw"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/stretchr/testify/assert"
)

func TestIndexRelated(t *testing.T) {
	conf := config.TestConfig()

	testFile, err := NewMediaFile("testdata/2018-04-12 19:24:49.gif")

	if err != nil {
		t.Fatal(err)
	}

	testRelated, err := testFile.RelatedFiles(true)

	if err != nil {
		t.Fatal(err)
	}

	testToken := rnd.Token(8)
	testPath := filepath.Join(conf.OriginalsPath(), testToken)

	for _, f := range testRelated.Files {
		dest := filepath.Join(testPath, f.BaseName())

		if err := f.Copy(dest); err != nil {
			t.Fatalf("COPY FAILED: %s", err)
		}
	}

	mainFile, err := NewMediaFile(filepath.Join(testPath, "2018-04-12 19:24:49.gif"))

	if err != nil {
		t.Fatal(err)
	}

	related, err := mainFile.RelatedFiles(true)

	if err != nil {
		t.Fatal(err)
	}

	tf := classify.New(conf.AssetsPath(), conf.TensorFlowOff())
	nd := nsfw.New(conf.NSFWModelPath())
	convert := NewConvert(conf)

	ind := NewIndex(conf, tf, nd, convert)
	opt := IndexOptionsAll()

	result := IndexRelated(related, ind, opt)

	assert.False(t, result.Failed())
	assert.False(t, result.Stacked())
	assert.True(t, result.Success())
	assert.Equal(t, IndexAdded, result.Status)

	if photo, err := query.PhotoByUID(result.PhotoUID); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(t, "2018-04-12 19:24:49 +0000 UTC", photo.TakenAt.String())
		assert.Equal(t, "name", photo.TakenSrc)
	}
}
