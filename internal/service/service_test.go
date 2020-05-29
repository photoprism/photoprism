package service

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/nsfw"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/session"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestMain(m *testing.M) {
	c := config.TestConfig()

	SetConfig(c)

	code := m.Run()

	_ = c.CloseDb()

	os.Exit(code)
}

func TestConfig(t *testing.T) {
	assert.Equal(t, conf, Config())
}

func TestClassify(t *testing.T) {
	assert.IsType(t, &classify.TensorFlow{}, Classify())
}

func TestConvert(t *testing.T) {
	assert.IsType(t, &photoprism.Convert{}, Convert())
}

func TestImport(t *testing.T) {
	assert.IsType(t, &photoprism.Import{}, Import())
}

func TestIndex(t *testing.T) {
	assert.IsType(t, &photoprism.Index{}, Index())
}

func TestMoments(t *testing.T) {
	assert.IsType(t, &photoprism.Moments{}, Moments())
}

func TestPurge(t *testing.T) {
	assert.IsType(t, &photoprism.Purge{}, Purge())
}

func TestNsfwDetector(t *testing.T) {
	assert.IsType(t, &nsfw.Detector{}, NsfwDetector())
}

func TestQuery(t *testing.T) {
	assert.IsType(t, &query.Query{}, Query())
}

func TestResample(t *testing.T) {
	assert.IsType(t, &photoprism.Resample{}, Resample())
}

func TestSession(t *testing.T) {
	assert.IsType(t, &session.Session{}, Session())
}
