package photoprism

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestPublishConverting(t *testing.T) {
	// receive returns the next index.converting event, or nil if none is published.
	receive := func(t *testing.T, publish func()) event.Data {
		t.Helper()

		s := event.Subscribe("index.converting")
		defer event.Unsubscribe(s)

		publish()

		select {
		case msg := <-s.Receiver:
			return msg.Fields
		case <-time.After(250 * time.Millisecond):
			return nil
		}
	}

	t.Run("LibraryFile", func(t *testing.T) {
		mf, err := NewMediaFile(filepath.Join(Config().SamplesPath(), "elephants.jpg"))
		require.NoError(t, err)

		fields := receive(t, func() { publishConverting(mf, "elephants.jpg", "") })

		require.NotNil(t, fields)
		assert.Equal(t, "elephants.jpg", fields["fileName"])
		assert.Equal(t, "elephants.jpg", fields["baseName"])
		assert.Equal(t, "", fields["xmpName"])
	})
	t.Run("SidecarFile", func(t *testing.T) {
		sidecarName := filepath.Join(Config().SidecarPath(), "convert-event", "reunion.jpg")
		require.NoError(t, os.MkdirAll(filepath.Dir(sidecarName), fs.ModeDir))
		require.NoError(t, fs.Copy(filepath.Join(Config().SamplesPath(), "elephants.jpg"), sidecarName, false))
		t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(sidecarName)) })

		mf, err := NewMediaFile(sidecarName)
		require.NoError(t, err)

		fields := receive(t, func() { publishConverting(mf, "convert-event/reunion.jpg", "") })

		require.NotNil(t, fields)
		assert.Equal(t, "reunion.jpg", fields["baseName"])
	})
	t.Run("ImportFile", func(t *testing.T) {
		importName := filepath.Join(Config().ImportPath(), "convert-event", "reunion.jpg")
		require.NoError(t, os.MkdirAll(filepath.Dir(importName), fs.ModeDir))
		require.NoError(t, fs.Copy(filepath.Join(Config().SamplesPath(), "elephants.jpg"), importName, false))
		t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(importName)) })

		mf, err := NewMediaFile(importName)
		require.NoError(t, err)

		fields := receive(t, func() { publishConverting(mf, "convert-event/reunion.jpg", "") })

		require.NotNil(t, fields)
		assert.Equal(t, "convert-event/reunion.jpg", fields["fileName"])
	})
	t.Run("CachedFile", func(t *testing.T) {
		cacheName := filepath.Join(t.TempDir(), "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818.mov")
		require.NoError(t, os.WriteFile(cacheName, []byte("not a video"), fs.ModeFile))

		mf, err := NewMediaFile(cacheName)
		require.NoError(t, err)

		assert.Nil(t, receive(t, func() { publishConverting(mf, mf.RootRelName(), "") }))
	})
	t.Run("NoFile", func(t *testing.T) {
		assert.Nil(t, receive(t, func() { publishConverting(nil, "elephants.jpg", "") }))
	})
}
