package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestFileName(t *testing.T) {
	c := config.TestConfig()
	t.Run("sidecar", func(t *testing.T) {
		assert.Equal(t, c.SidecarPath()+"/test.jpg", FileName("sidecar", "test.jpg"))
	})
	t.Run("import", func(t *testing.T) {
		assert.Equal(t, c.ImportPath()+"/test.jpg", FileName("import", "test.jpg"))
	})
	t.Run("examples", func(t *testing.T) {
		assert.Equal(t, c.ExamplesPath()+"/test.jpg", FileName("examples", "test.jpg"))
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

func TestRelName(t *testing.T) {
	c := config.TestConfig()
	t.Run("SidecarPath", func(t *testing.T) {
		assert.Equal(t, "foo/test.jpg", RelName(FileName("sidecar", "foo/test.jpg"), c.SidecarPath()))
	})
}

func TestRootPath(t *testing.T) {
	c := config.TestConfig()
	t.Run("SidecarPath", func(t *testing.T) {
		assert.Equal(t, c.SidecarPath(), RootPath(FileName("sidecar", "test.jpg")))
	})
}

func TestRoot(t *testing.T) {
	t.Run("SidecarPath", func(t *testing.T) {
		assert.Equal(t, entity.RootSidecar, Root(FileName("sidecar", "test.jpg")))
	})
}

func TestRootRelName(t *testing.T) {
	t.Run("SidecarPath", func(t *testing.T) {
		assert.Equal(t, "foo/test.jpg", RootRelName(FileName("sidecar", "foo/test.jpg")))
	})
}
