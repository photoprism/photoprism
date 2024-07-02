package get

import (
	"testing"

	gc "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/auth/oidc"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/session"
)

func TestConfig(t *testing.T) {
	assert.Equal(t, conf, Config())
}

func TestFolderCache(t *testing.T) {
	assert.IsType(t, &gc.Cache{}, FolderCache())
}

func TestCoverCache(t *testing.T) {
	assert.IsType(t, &gc.Cache{}, CoverCache())
}

func TestThumbCache(t *testing.T) {
	assert.IsType(t, &gc.Cache{}, ThumbCache())
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

func TestCleanUp(t *testing.T) {
	assert.IsType(t, &photoprism.CleanUp{}, CleanUp())
}

func TestNsfwDetector(t *testing.T) {
	assert.IsType(t, &nsfw.Detector{}, NsfwDetector())
}

func TestQuery(t *testing.T) {
	assert.IsType(t, &query.Query{}, Query())
}

func TestResample(t *testing.T) {
	assert.IsType(t, &photoprism.Thumbs{}, Thumbs())
}

func TestSession(t *testing.T) {
	assert.IsType(t, &session.Session{}, Session())
}

func TestOIDC(t *testing.T) {
	assert.IsType(t, &oidc.Client{}, OIDC())
}
