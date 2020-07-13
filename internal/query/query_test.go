package query

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)

	db := entity.InitTestDb(os.Getenv("PHOTOPRISM_TEST_DRIVER"), os.Getenv("PHOTOPRISM_TEST_DSN"))
	defer db.Close()

	code := m.Run()

	os.Exit(code)
}

func TestLikeAny(t *testing.T) {
	t.Run("table spoon usa img json", func(t *testing.T) {
		where := LikeAny("k.keyword", "table spoon usa img json")
		assert.Equal(t, "k.keyword LIKE 'json%' OR k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%' OR k.keyword = 'usa'", where)
	})

	t.Run("cat dog", func(t *testing.T) {
		where := LikeAny("k.keyword", "cat dog")
		assert.Equal(t, "k.keyword = 'cat' OR k.keyword = 'dog'", where)
	})

	t.Run("cats dogs", func(t *testing.T) {
		where := LikeAny("k.keyword", "cats dogs")
		assert.Equal(t, "k.keyword LIKE 'cats%' OR k.keyword = 'cat' OR k.keyword LIKE 'dogs%' OR k.keyword = 'dog'", where)
	})

	t.Run("spoon", func(t *testing.T) {
		where := LikeAny("k.keyword", "spoon")
		assert.Equal(t, "k.keyword LIKE 'spoon%'", where)
	})

	t.Run("img", func(t *testing.T) {
		where := LikeAny("k.keyword", "img")
		assert.Equal(t, "", where)
	})

	t.Run("empty", func(t *testing.T) {
		where := LikeAny("k.keyword", "")
		assert.Equal(t, "", where)
	})
}

func TestAnySlug(t *testing.T) {
	t.Run("table spoon usa img json", func(t *testing.T) {
		where := AnySlug("custom_slug", "table spoon usa img json", " ")
		assert.Equal(t, "custom_slug = 'table' OR custom_slug = 'spoon' OR custom_slug = 'usa' OR custom_slug = 'img' OR custom_slug = 'json'", where)
	})

	t.Run("cat dog", func(t *testing.T) {
		where := AnySlug("custom_slug", "cat dog", " ")
		assert.Equal(t, "custom_slug = 'cat' OR custom_slug = 'dog'", where)
	})

	t.Run("cats dogs", func(t *testing.T) {
		where := AnySlug("custom_slug", "cats dogs", " ")
		assert.Equal(t, "custom_slug = 'cats' OR custom_slug = 'cat' OR custom_slug = 'dogs' OR custom_slug = 'dog'", where)
	})

	t.Run("spoon", func(t *testing.T) {
		where := AnySlug("custom_slug", "spoon", " ")
		assert.Equal(t, "custom_slug = 'spoon'", where)
	})

	t.Run("img", func(t *testing.T) {
		where := AnySlug("custom_slug", "img", " ")
		assert.Equal(t, "custom_slug = 'img'", where)
	})

	t.Run("empty", func(t *testing.T) {
		where := AnySlug("custom_slug", "", " ")
		assert.Equal(t, "", where)
	})

	t.Run("comma separated", func(t *testing.T) {
		where := AnySlug("custom_slug", "botanical-garden,landscape,bay", ",")
		assert.Equal(t, "custom_slug = 'botanical-garden' OR custom_slug = 'landscape' OR custom_slug = 'bay'", where)
	})

	t.Run("len = 0", func(t *testing.T) {
		where := AnySlug("custom_slug", " ", "")
		assert.Equal(t, "custom_slug = '' OR custom_slug = ''", where)
	})
}
