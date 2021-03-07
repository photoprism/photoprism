package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestFileName(t *testing.T) {
	conf := config.TestConfig()
	t.Run("sidecar", func(t *testing.T) {
		assert.Equal(t, conf.SidecarPath()+"/test.jpg", FileName("sidecar", "test.jpg"))
	})
	t.Run("import", func(t *testing.T) {
		assert.Equal(t, conf.ImportPath()+"/test.jpg", FileName("import", "test.jpg"))
	})
	t.Run("examples", func(t *testing.T) {
		assert.Equal(t, conf.ExamplesPath()+"/test.jpg", FileName("examples", "test.jpg"))
	})

}

func TestCacheName(t *testing.T) {
	t.Run("cacheKey empty", func(t *testing.T) {
		r, err := CacheName("abcdghoj", "test", "")
		assert.Error(t, err)
		assert.Empty(t, r)
	})

	t.Run("success", func(t *testing.T) {
		r, err := CacheName("abcdghoj", "test", "juh")
		if err != nil {
			t.Fatal(err)
		}
		assert.Contains(t, r, "test/a/b/c/abcdghoj_juh")
	})
	t.Run("filehash too short", func(t *testing.T) {
		r, err := CacheName("ab", "test", "juh")
		assert.Error(t, err)
		assert.Empty(t, r)
	})
}
