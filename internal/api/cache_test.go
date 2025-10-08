package api

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/service/http/header"
)

func TestAddVideoCacheHeader(t *testing.T) {
	t.Run("Public", func(t *testing.T) {
		r := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(r)
		AddVideoCacheHeader(c, true)
		h := r.Header()
		s := h[header.CacheControl][0]
		assert.Equal(t, "public, max-age=21600, immutable", s)
	})
	t.Run("Private", func(t *testing.T) {
		r := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(r)
		AddVideoCacheHeader(c, false)
		h := r.Header()
		s := h[header.CacheControl][0]
		assert.Equal(t, "private, max-age=21600, immutable", s)
	})
}

func TestRemoveFromFolderCache(t *testing.T) {
	cache := get.FolderCache()
	cache.Flush()

	root := "originals"
	key := fmt.Sprintf("folder:%s:%t:%t", root, true, false)

	cache.SetDefault(key, FoldersResponse{Root: root})

	RemoveFromFolderCache(root)

	_, ok := cache.Get(key)
	assert.False(t, ok)
}

func TestRemoveFromAlbumCoverCache(t *testing.T) {
	cache := get.CoverCache()
	cache.Flush()

	uid := rnd.GenerateUID(entity.AlbumUID)

	for thumbName := range thumb.Sizes {
		key := CacheKey(albumCover, uid, string(thumbName))
		cache.SetDefault(key, ThumbCache{FileName: "cached-file", ShareName: "share"})
	}

	conf := get.Config()
	shareDir := path.Join(conf.ThumbCachePath(), "share")

	if err := fs.MkdirAll(shareDir); err != nil {
		t.Fatalf("mkdir %s: %v", shareDir, err)
	}

	sharePreview := path.Join(shareDir, uid+fs.ExtJpeg)

	if err := os.WriteFile(sharePreview, []byte("preview"), fs.ModeFile); err != nil {
		t.Fatalf("write %s: %v", sharePreview, err)
	}

	RemoveFromAlbumCoverCache(uid)

	for thumbName := range thumb.Sizes {
		key := CacheKey(albumCover, uid, string(thumbName))
		_, ok := cache.Get(key)
		assert.False(t, ok)
	}

	_, err := os.Stat(sharePreview)
	assert.True(t, os.IsNotExist(err))
}

func TestRemoveFromAlbumCoverCacheInvalidUID(t *testing.T) {
	cache := get.CoverCache()
	cache.Flush()

	uid := "" // empty string fails rnd.IsAlnum
	key := CacheKey(albumCover, uid, thumb.Tile500.String())
	cache.SetDefault(key, ThumbCache{FileName: "file", ShareName: "share"})

	RemoveFromAlbumCoverCache(uid)

	_, ok := cache.Get(key)
	assert.True(t, ok)
}

func TestRemoveFromLabelCoverCache(t *testing.T) {
	cache := get.CoverCache()
	cache.Flush()

	uid := rnd.GenerateUID(entity.LabelUID)

	for thumbName := range thumb.Sizes {
		key := CacheKey(labelCover, uid, string(thumbName))
		cache.SetDefault(key, ThumbCache{FileName: "cached-file", ShareName: "share"})
	}

	RemoveFromLabelCoverCache(uid)

	for thumbName := range thumb.Sizes {
		key := CacheKey(labelCover, uid, string(thumbName))
		_, ok := cache.Get(key)
		assert.False(t, ok)
	}
}

func TestRemoveFromLabelCoverCacheInvalidUID(t *testing.T) {
	cache := get.CoverCache()
	cache.Flush()

	uid := ""
	key := CacheKey(labelCover, uid, thumb.Tile500.String())
	cache.SetDefault(key, ThumbCache{FileName: "file", ShareName: "share"})

	RemoveFromLabelCoverCache(uid)

	_, ok := cache.Get(key)
	assert.True(t, ok)
}
