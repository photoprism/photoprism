package vision

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
)

// sampleImageDataUrl returns a data URL built from a local sample image.
func sampleImageDataUrl(t *testing.T) string {
	data, err := os.ReadFile(samplesPath + "/chameleon_lime.jpg")
	require.NoError(t, err)
	return media.DataUrl(bytes.NewReader(data))
}

func TestNewApiRequestImages_SrcLocal(t *testing.T) {
	req, err := NewApiRequestImages(Files{samplesPath + "/chameleon_lime.jpg"}, scheme.Data, media.SrcLocal)
	require.NoError(t, err)
	require.Len(t, req.Images, 1)
	assert.True(t, strings.HasPrefix(req.Images[0], "data:image/"))
}

func TestNewApiRequestImages_SrcRemote(t *testing.T) {
	dataUrl := sampleImageDataUrl(t)
	t.Run("DataUrl", func(t *testing.T) {
		req, err := NewApiRequestImages(Files{dataUrl}, scheme.Data, media.SrcRemote)
		require.NoError(t, err)
		require.Len(t, req.Images, 1)
		assert.True(t, strings.HasPrefix(req.Images[0], "data:image/"))
	})
	t.Run("LocalPathRejected", func(t *testing.T) {
		_, err := NewApiRequestImages(Files{samplesPath + "/chameleon_lime.jpg"}, scheme.Data, media.SrcRemote)
		assert.Error(t, err)
	})
	t.Run("AbsolutePathRejected", func(t *testing.T) {
		_, err := NewApiRequestImages(Files{"/private/example.txt"}, scheme.Data, media.SrcRemote)
		assert.Error(t, err)
	})
	t.Run("FileSchemeRejected", func(t *testing.T) {
		_, err := NewApiRequestImages(Files{"file:///private/example.txt"}, scheme.Data, media.SrcRemote)
		assert.Error(t, err)
	})
	t.Run("NonImageDataUrlRejected", func(t *testing.T) {
		_, err := NewApiRequestImages(Files{"data:text/plain;base64,aGVsbG8="}, scheme.Data, media.SrcRemote)
		assert.ErrorIs(t, err, ErrNotAnImage)
	})
	t.Run("PrivateHostRejected", func(t *testing.T) {
		// PhotoPrism's own fetch (ReadUrlImage, AllowPrivate=false) blocks link-local /
		// cloud-metadata targets before connecting.
		_, err := NewApiRequestImages(Files{"https://169.254.169.254/latest/meta-data/"}, scheme.Data, media.SrcRemote)
		assert.Error(t, err)
	})
	t.Run("InvalidSource", func(t *testing.T) {
		_, err := NewApiRequestImages(Files{dataUrl}, scheme.Data, "")
		assert.Error(t, err)
	})
}

func TestNewApiRequestOllama_Source(t *testing.T) {
	dataUrl := sampleImageDataUrl(t)
	t.Run("SrcLocal", func(t *testing.T) {
		req, err := NewApiRequestOllama(Files{samplesPath + "/chameleon_lime.jpg"}, scheme.Base64, media.SrcLocal)
		require.NoError(t, err)
		require.Len(t, req.Images, 1)
		assert.NotEmpty(t, req.Images[0])
	})
	t.Run("SrcRemoteDataUrl", func(t *testing.T) {
		req, err := NewApiRequestOllama(Files{dataUrl}, scheme.Base64, media.SrcRemote)
		require.NoError(t, err)
		require.Len(t, req.Images, 1)
		assert.NotEmpty(t, req.Images[0])
	})
	t.Run("SrcRemoteLocalPathRejected", func(t *testing.T) {
		_, err := NewApiRequestOllama(Files{"/private/example.txt"}, scheme.Base64, media.SrcRemote)
		assert.Error(t, err)
	})
}

func TestNewApiRequestUrl_Source(t *testing.T) {
	t.Run("SrcRemoteHttpUrl", func(t *testing.T) {
		req, err := NewApiRequestUrl("https://example.org/image.jpg", scheme.Https, media.SrcRemote)
		require.NoError(t, err)
		assert.Equal(t, "https://example.org/image.jpg", req.Url)
	})
	t.Run("SrcRemoteDataUrl", func(t *testing.T) {
		req, err := NewApiRequestUrl(sampleImageDataUrl(t), scheme.Data, media.SrcRemote)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(req.Url, "data:image/"))
	})
	t.Run("SrcRemoteLocalPathRejected", func(t *testing.T) {
		_, err := NewApiRequestUrl("/private/example.txt", scheme.Https, media.SrcRemote)
		assert.Error(t, err)
	})
	t.Run("InvalidSource", func(t *testing.T) {
		_, err := NewApiRequestUrl("https://example.org/image.jpg", scheme.Https, "")
		assert.Error(t, err)
	})
}

func TestRemoteImageData(t *testing.T) {
	t.Run("DataUrl", func(t *testing.T) {
		data, err := remoteImageData(sampleImageDataUrl(t))
		require.NoError(t, err)
		assert.NotEmpty(t, data)
	})
	t.Run("LocalPathRejected", func(t *testing.T) {
		_, err := remoteImageData("/private/example.txt")
		assert.Error(t, err)
	})
	t.Run("NonImageRejected", func(t *testing.T) {
		_, err := remoteImageData("data:text/plain;base64,aGVsbG8=")
		assert.ErrorIs(t, err, ErrNotAnImage)
	})
	t.Run("PrivateHostRejected", func(t *testing.T) {
		_, err := remoteImageData("https://169.254.169.254/latest/meta-data/")
		assert.Error(t, err)
	})
}
