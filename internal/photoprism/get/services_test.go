package get

import (
	"errors"
	"net/url"
	"testing"

	gc "github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/oidc"
	"github.com/photoprism/photoprism/internal/auth/session"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
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
	origConf := Config()
	origFactory := newOIDCClient

	t.Cleanup(func() {
		newOIDCClient = origFactory
		SetConfig(origConf)
	})

	t.Run("CachesSuccess", func(t *testing.T) {
		tempConf := config.NewMinimalTestConfig(t.TempDir())
		SetConfig(tempConf)

		calls := 0
		expected := &oidc.Client{}

		newOIDCClient = func(_ *url.URL, _ string, _ string, _ string, _ string, _ bool) (*oidc.Client, error) {
			calls++
			return expected, nil
		}

		client := OIDC()
		assert.Same(t, expected, client)
		assert.Equal(t, 1, calls)
		assert.Same(t, expected, OIDC())
		assert.Equal(t, 1, calls)
	})
	t.Run("RetriesAfterFailure", func(t *testing.T) {
		tempConf := config.NewMinimalTestConfig(t.TempDir())
		SetConfig(tempConf)

		calls := 0
		expected := &oidc.Client{}

		newOIDCClient = func(_ *url.URL, _ string, _ string, _ string, _ string, _ bool) (*oidc.Client, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("service discovery failed")
			}

			return expected, nil
		}

		assert.Nil(t, OIDC())
		assert.Equal(t, 1, calls)

		client := OIDC()
		assert.Same(t, expected, client)
		assert.Equal(t, 2, calls)
		assert.Same(t, expected, OIDC())
		assert.Equal(t, 2, calls)
	})
}
